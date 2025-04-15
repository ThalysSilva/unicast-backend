package smtp

import (
	"context"
	"encoding/base64"

	"github.com/ThalysSilva/unicast-backend/internal/auth"
	"github.com/ThalysSilva/unicast-backend/internal/encryption"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
)

type smtpService struct {
	smtpRepository Repository
}

type Service interface {
	Create(ctx context.Context, jweSecret []byte, userId, jwe, email, password, host string, port int) error
	GetInstances(ctx context.Context, userID string) ([]*Instance, error)
}

func NewService(smtpRepository Repository) Service {
	return &smtpService{smtpRepository: smtpRepository}
}

func (s *smtpService) Create(ctx context.Context, jweSecret []byte, userId, jwe, email, password, host string, port int) error {
	decryptedJwe, err := auth.DecryptJWE[auth.JwePayload](jwe, jweSecret)
	if err != nil {
		return customerror.Trace("Create", err)
	}
	smtpKey, err := base64.StdEncoding.DecodeString(decryptedJwe.SmtpKeyEncoded)
	if err != nil {
		return customerror.Trace("Create", err)
	}

	encryptedPassword, iv, err := encryption.EncryptSmtpPassword(password, smtpKey)
	if err != nil {
		return customerror.Trace("Create", err)
	}

	if (s.smtpRepository.Create(ctx, userId, email, host, port, encryptedPassword, iv)) != nil {
		return customerror.Trace("Create", err)
	}
	return nil
}

func (s *smtpService) GetInstances(ctx context.Context, userID string) ([]*Instance, error) {
	return s.smtpRepository.GetInstances(ctx, userID)
}
