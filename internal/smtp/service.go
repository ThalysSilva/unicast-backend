package smtp

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/ThalysSilva/unicast-backend/internal/auth"
	configenv "github.com/ThalysSilva/unicast-backend/internal/config/env"
	"github.com/ThalysSilva/unicast-backend/internal/encryption"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/mailer"
)

type smtpService struct {
	smtpRepository Repository
	jweSecret      []byte
	oauth          configenv.OAuth
}

type Service interface {
	Create(ctx context.Context, jweSecret []byte, userId, jwe, email, password, host string, port int) error
	StartOAuth(ctx context.Context, userID, provider string) (string, error)
	HandleOAuthCallback(ctx context.Context, provider, code, state string) (string, error)
	TestConnection(ctx context.Context, email, password, host string, port int) error
	GetInstances(ctx context.Context, userID string) ([]*Instance, error)
	DeleteInstance(ctx context.Context, userID, instanceID string) error
	RefreshOAuthAccessToken(ctx context.Context, instance *Instance) (string, error)
}

func NewService(smtpRepository Repository, jweSecret []byte, oauth configenv.OAuth) Service {
	return &smtpService{smtpRepository: smtpRepository, jweSecret: jweSecret, oauth: oauth}
}

var (
	InstanceNotFound  = customerror.Make("Instância SMTP não encontrada", http.StatusNotFound, errors.New("smtpInstanceNotFound"))
	InstanceForbidden = customerror.Make("Você não tem permissão para esta instância SMTP", http.StatusForbidden, errors.New("smtpInstanceForbidden"))
)

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

func (s *smtpService) TestConnection(ctx context.Context, email, password, host string, port int) error {
	err := mailer.TestSMTPConnection(mailer.SmtpAuthentication{
		Host:     host,
		Port:     port,
		Username: email,
		Password: password,
	}, 8*time.Second)
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "SmtpClientAuthentication is disabled for the Mailbox") {
		return customerror.Trace("TestConnection", customerror.Make("SMTP AUTH desabilitado para esta mailbox", http.StatusBadRequest, err))
	}

	return customerror.Trace("TestConnection", customerror.Make("Falha ao testar conexao SMTP", http.StatusBadGateway, err))
}

func (s *smtpService) GetInstances(ctx context.Context, userID string) ([]*Instance, error) {
	return s.smtpRepository.GetInstances(ctx, userID)
}

func (s *smtpService) DeleteInstance(ctx context.Context, userID, instanceID string) error {
	instance, err := s.smtpRepository.FindByID(ctx, instanceID)
	if err != nil {
		return err
	}
	if instance == nil {
		return InstanceNotFound
	}
	if instance.UserID != userID {
		return InstanceForbidden
	}
	return s.smtpRepository.Delete(ctx, instanceID)
}
