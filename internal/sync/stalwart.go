package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type StalwartJMAPConnector struct {
	client  *http.Client
	jmapURL string
	token   string
}

func NewStalwartJMAPConnector(jmapURL, token string) *StalwartJMAPConnector {
	return &StalwartJMAPConnector{
		client:  &http.Client{Timeout: 15 * time.Second},
		jmapURL: jmapURL,
		token:   token,
	}
}

// Structures de requêtage JMAP standard (RFC 8620)
type JMAPRequest struct {
	Using       []string        `json:"using"`
	MethodCalls [][]interface{} `json:"methodCalls"`
}

type JMAPResponse struct {
	MethodResponses [][]interface{} `json:"methodResponses"`
}

// Structures de données calquées sur les spécifications de l'objet x:Account de Stalwart
type StalwartAlias struct {
	Enabled     bool   `json:"enabled"`
	Name        string `json:"name"`
	DomainID    string `json:"domainId"`
	Description string `json:"description"`
}

type StalwartAccountObj struct {
	ID             string          `json:"id"`
	Type           string          `json:"@type"` // "User" ou "Group"
	Name           string          `json:"name"`
	EmailAddress   string          `json:"emailAddress"`
	Aliases        []StalwartAlias `json:"aliases"`
	MemberGroupIds []string        `json:"memberGroupIds"` // Pour les Users : les groupes dont ils font partie
}

func (s *StalwartJMAPConnector) FetchIdentities(ctx context.Context) ([]ExternalIdentity, error) {
	// 1. Préparation du payload JMAP combinant Query + Get (Back-reference #ids)
	payload := JMAPRequest{
		Using: []string{
			"urn:ietf:params:jmap:core",
			"urn:stalwart:jmap", // Capacité officielle indiquée dans tes specs
		},
		MethodCalls: [][]interface{}{
			{
				"x:Account/query",
				map[string]interface{}{
					"filter": map[string]interface{}{}, // Filtre vide pour tout récupérer
				},
				"q1",
			},
			{
				"x:Account/get",
				map[string]interface{}{
					"#ids": map[string]string{
						"name":     "x:Account/query",
						"path":     "/ids",
						"resultOf": "q1",
					},
				},
				"g1",
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal jmap request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.jmapURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error during jmap call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stalwart jmap api returned status %d", resp.StatusCode)
	}

	var jmapResp JMAPResponse
	if err := json.NewDecoder(resp.Body).Decode(&jmapResp); err != nil {
		return nil, fmt.Errorf("failed to decode jmap response: %w", err)
	}

	if len(jmapResp.MethodResponses) < 2 {
		return nil, fmt.Errorf("invalid jmap response: missing method responses")
	}

	// 2. Récupération de la liste d'objets retournée par x:Account/get (index 1 de MethodResponses)
	getBatch := jmapResp.MethodResponses[1]
	getResults, ok := getBatch[1].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid x:Account/get response payload")
	}

	listRaw, ok := getResults["list"]
	if !ok {
		return nil, fmt.Errorf("jmap account list missing from response")
	}

	listBytes, err := json.Marshal(listRaw)
	if err != nil {
		return nil, err
	}

	var accounts []StalwartAccountObj
	if err := json.Unmarshal(listBytes, &accounts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal accounts list: %w", err)
	}

	// 3. Transformation et aiguillage selon tes contraintes métier
	var results []ExternalIdentity

	// Map temporaire pour inverser la relation d'appartenance aux groupes.
	// Stalwart stocke "User -> fait partie de ces groupes".
	// Ton besoin est d'avoir "Groupe -> contient ces utilisateurs".
	// Map qui associe l'ID d'un groupe (string) à la liste des EMAILS de ses membres ([]string)
	groupMembersMap := make(map[string][]string)

	// ==========================================
	// Premier passage : Utilisateurs et leurs Alias
	// ==========================================
	for _, acc := range accounts {
		if acc.Type == "User" {
			// 1. Ajouter l'utilisateur principal
			results = append(results, ExternalIdentity{
				Email:    acc.EmailAddress,
				IsGroup:  false,
				IsActive: true,
			})

			// 2. Ajouter ses alias
			for _, alias := range acc.Aliases {
				if alias.Enabled {
					aliasEmail := fmt.Sprintf("%s@%s", alias.Name, alias.DomainID)
					results = append(results, ExternalIdentity{
						Email:       aliasEmail,
						IsGroup:     false,
						IsActive:    true,
						ParentEmail: acc.EmailAddress,
					})
				}
			}

			// 3. Collecter l'appartenance aux groupes
			// acc.MemberGroupIds contient les IDs des groupes (ex: "grp_abc")
			// On y associe l'email de l'utilisateur actuel
			for _, groupID := range acc.MemberGroupIds {
				groupMembersMap[groupID] = append(groupMembersMap[groupID], acc.EmailAddress)
			}
		}
	}

	// ==========================================
	// Second passage : Groupes autonomes
	// ==========================================
	for _, acc := range accounts {
		if acc.Type == "Group" {
			// On extrait la liste des EMAILS des membres en utilisant l'ID unique du groupe (acc.ID)
			members := groupMembersMap[acc.ID]
			if members == nil {
				members = []string{} // Évite de renvoyer un slice nil
			}

			results = append(results, ExternalIdentity{
				Email:    acc.EmailAddress, // L'adresse du groupe (ex: contact@aurion.local)
				IsGroup:  true,
				IsActive: true,
				Members:  members, // Contient maintenant bien les ADRESSES EMAILS des membres, pas leurs IDs
			})
		}
	}

	return results, nil
}
