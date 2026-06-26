package sync

import (
	"context"
	"log/slog"
	"time"

	"aurion/core/internal/db/repository"
)

type SyncService struct {
	connector  MailServerConnector
	userRepo   *repository.UserRepository
	identRepo  *repository.IdentityRepository
	memberRepo *repository.IdentityMemberRepository // <-- AJOUT
	logger     *slog.Logger
	interval   time.Duration
}

func NewSyncService(
	conn MailServerConnector,
	userRepo *repository.UserRepository,
	identRepo *repository.IdentityRepository,
	memberRepo *repository.IdentityMemberRepository, // <-- AJOUT
	logger *slog.Logger,
	interval time.Duration,
) *SyncService {
	return &SyncService{
		connector:  conn,
		userRepo:   userRepo,
		identRepo:  identRepo,
		memberRepo: memberRepo,
		logger:     logger,
		interval:   interval,
	}
}

func (s *SyncService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.RunSync(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *SyncService) RunSync(ctx context.Context) {
	s.logger.Info("Starting identities synchronization from mail server...")

	extIdentities, err := s.connector.FetchIdentities(ctx)
	if err != nil {
		s.logger.Error("Failed to fetch external identities", "error", err)
		return
	}

	for _, ext := range extIdentities {
		s.logger.Debug("Processing synchronized identity", "email", ext.Email, "is_group", ext.IsGroup)

		// 1. Détermination du type selon ton modèle
		// L'adresse principale est 'primary', les alias et les groupes sont 'shared'
		isShared := ext.IsGroup || ext.ParentEmail != ""
		identityType := "primary"
		if isShared {
			identityType = "shared"
		}

		// 2. Assurer l'existence de l'Identity en base
		dbIdent, err := s.identRepo.GetByEmail(ctx, ext.Email)
		if err != nil {
			dbIdent, err = s.identRepo.CreateIdentity(ctx, ext.Email, identityType)
			if err != nil {
				s.logger.Error("Failed to create identity during sync", "email", ext.Email, "error", err)
				continue
			}
			s.logger.Info("Created new identity", "email", ext.Email, "type", identityType)
		}

		// 3. Aiguillage de la jointure des membres
		if isShared {
			// ---- CAS 1 : ALIAS OU GROUPE ('shared') ----
			// On détermine la liste des emails des membres à lier
			var memberEmails []string
			if ext.IsGroup {
				memberEmails = ext.Members
			} else {
				// C'est un alias, donc un groupe avec un seul membre (le parent)
				memberEmails = []string{ext.ParentEmail}
			}

			// On lie le ou les membres à cette identité partagée
			for _, memberEmail := range memberEmails {
				dbUser, err := s.userRepo.GetUserByEmail(ctx, memberEmail)
				if err != nil {
					s.logger.Warn("Member user not found in database yet, skipping link for now", "identity", ext.Email, "member", memberEmail)
					continue
				}

				err = s.memberRepo.AddMember(ctx, dbIdent.ID, dbUser.ID.String())
				if err != nil {
					s.logger.Debug("Member already linked to this shared identity", "identity", ext.Email, "user_id", dbUser.ID)
				}
			}

		} else {
			// ---- CAS 2 : COMPTE PRINCIPAL ('primary') ----
			// L'utilisateur doit exister pour son adresse principale
			dbUser, err := s.userRepo.GetUserByEmail(ctx, ext.Email)
			if err != nil {
				// Création du Shadow User en attente d'onboarding
				dbUser, err = s.userRepo.CreateUser(ctx, ext.Email, "PENDING_ONBOARDING", "PENDING", "PENDING")
				if err != nil {
					s.logger.Error("Failed to create shadow user for primary identity", "email", ext.Email, "error", err)
					continue
				}
				s.logger.Info("Created shadow user pending activation", "email", ext.Email)
			}

			// Un user est le seul membre de son identité 'primary'
			err = s.memberRepo.AddMember(ctx, dbIdent.ID, dbUser.ID.String())
			if err != nil {
				s.logger.Debug("User already linked to their primary identity", "email", ext.Email)
			}
		}
	}

	s.logger.Info("Identities synchronization finished successfully")
}
