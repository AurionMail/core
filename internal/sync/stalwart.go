package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

// JMAPRequest correspond aux structures de requêtage JMAP standard (RFC 8620)
type JMAPRequest struct {
	Using       []string        `json:"using"`
	MethodCalls [][]interface{} `json:"methodCalls"`
}

type JMAPResponse struct {
	MethodResponses [][]interface{} `json:"methodResponses"`
}

// StalwartAlias correspond aux spécifications réelles des alias renvoyés par Stalwart
type StalwartAlias struct {
	Enabled     bool   `json:"enabled"`
	Name        string `json:"name"`
	DomainID    string `json:"domainId"` // Attention : souvent l'ID interne (ex: "b")
	Description string `json:"description"`
}

// StalwartAccountObj modélise un compte (User ou Group) tel que retourné dans ton payload
type StalwartAccountObj struct {
	ID           string `json:"id"`
	Type         string `json:"@type"` // "User" ou "Group"
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`

	// Utilisation de maps pour éviter le crash sur les objets JSON {}
	Aliases        map[string]StalwartAlias `json:"aliases"`
	MemberGroupIds map[string]interface{}   `json:"memberGroupIds"`
}

func (s *StalwartJMAPConnector) FetchIdentities(ctx context.Context) ([]ExternalIdentity, error) {
	// 1. Préparation du payload JMAP combinant Query + Get
	payload := JMAPRequest{
		Using: []string{
			"urn:ietf:params:jmap:core",
			"urn:stalwart:jmap",
		},
		MethodCalls: [][]interface{}{
			{
				"x:Account/query",
				map[string]interface{}{
					"filter": map[string]interface{}{},
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

	// 2. Extraction du résultat de x:Account/get
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

	// 3. Traitement et pivot des données
	var results []ExternalIdentity
	groupMembersMap := make(map[string][]string)

	// ==========================================
	// Premier passage : Utilisateurs, Groupes d'appartenance et Alias
	// ==========================================
	for _, acc := range accounts {
		if acc.Type == "User" {
			// A. Enregistrement de l'utilisateur principal
			results = append(results, ExternalIdentity{
				Email:    acc.EmailAddress,
				IsGroup:  false,
				IsActive: true,
			})

			// Extraction du nom de domaine public (ex: "aurionmail.org") à partir de l'adresse principale
			domain := ""
			if parts := strings.Split(acc.EmailAddress, "@"); len(parts) == 2 {
				domain = parts[1]
			}

			// B. Enregistrement de ses alias (si présents dans la map)
			for _, alias := range acc.Aliases {
				if alias.Enabled && domain != "" {
					aliasEmail := fmt.Sprintf("%s@%s", alias.Name, domain)
					results = append(results, ExternalIdentity{
						Email:       aliasEmail,
						IsGroup:     false,
						IsActive:    true,
						ParentEmail: acc.EmailAddress,
					})
				}
			}

			// C. Collecte de l'appartenance aux groupes (Inversion de l'index)
			for groupID := range acc.MemberGroupIds {
				groupMembersMap[groupID] = append(groupMembersMap[groupID], acc.EmailAddress)
			}
		}
	}

	// ==========================================
	// Second passage : Résolution des Groupes
	// ==========================================
	for _, acc := range accounts {
		if acc.Type == "Group" {
			members := groupMembersMap[acc.ID]
			if members == nil {
				members = []string{} // Évite un slice nil dans le résultat final
			}

			results = append(results, ExternalIdentity{
				Email:    acc.EmailAddress,
				IsGroup:  true,
				IsActive: true,
				Members:  members, // Contient les adresses email des membres collectés au 1er passage
			})
		}
	}

	return results, nil
}
