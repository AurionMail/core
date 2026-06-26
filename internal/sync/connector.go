package sync

import "context"

// ExternalIdentity représente un utilisateur, un alias ou un groupe renvoyé par le serveur mail
type ExternalIdentity struct {
	Email       string   // L'adresse (ex: user@, alias@ ou groupe@)
	IsGroup     bool     // true si c'est un groupe
	IsActive    bool     // true si le compte est actif
	ParentEmail string   //  Contient l'email de l'User principal si c'est un alias
	Members     []string // Rempli uniquement si IsGroup = true (emails des membres)
}

// MailServerConnector est le contrat que tous les futurs connecteurs devront remplir
type MailServerConnector interface {
	FetchIdentities(ctx context.Context) ([]ExternalIdentity, error)
}
