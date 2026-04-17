package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ThalysSilva/unicast-backend/internal/user"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Service interface {
	CreateInstance(ctx context.Context, userId, phone string) (*Instance, *connectResponse, error)
	GetInstances(ctx context.Context, userID string) ([]*Instance, error)
	DeleteInstance(ctx context.Context, userID, instanceID string) error
	ConnectInstance(ctx context.Context, userID, instanceID string) (*connectResponse, error)
	ConnectionState(ctx context.Context, userID, instanceID string) (string, error)
	LogoutInstance(ctx context.Context, userID, instanceID string) error
	RestartInstance(ctx context.Context, userID, instanceID string) error
}

type service struct {
	whatsappInstanceRepository Repository
	userRepository             user.Repository
}

func NewService(whatsappRepo Repository, userRepo user.Repository) Service {
	return &service{
		whatsappInstanceRepository: whatsappRepo,
		userRepository:             userRepo,
	}
}

var (
	HasAlreadyInstance = customerror.Make("Instância já existe", http.StatusConflict, errors.New("HasAlreadyInstance"))
	UserNotFound       = customerror.Make("Usuário não encontrado", http.StatusNotFound, errors.New("userNotFound"))
	InstanceNotFound   = customerror.Make("Instância não encontrada", http.StatusNotFound, errors.New("instanceNotFound"))
	InstanceForbidden  = customerror.Make("Você não tem permissão para esta instância", http.StatusForbidden, errors.New("instanceForbidden"))
	InvalidPhone       = customerror.Make("telefone deve estar em formato internacional, com DDI. Exemplo: +5511999999999", http.StatusBadRequest, errors.New("invalidPhone"))
)

func (s *service) CreateInstance(ctx context.Context, userId, phone string) (*Instance, *connectResponse, error) {
	var instance *Instance
	if !strings.HasPrefix(strings.TrimSpace(phone), "+") {
		return nil, nil, customerror.Trace("CreateInstance", InvalidPhone)
	}
	normalizedPhone, err := NormalizeNumber(phone, "")
	if err != nil {
		return nil, nil, customerror.Trace("CreateInstance", err)
	}
	phone = normalizedPhone

	// Fase 1: checa existência e usuário dentro da transação.
	var instanceName string
	err = s.withTransaction(ctx, func(waRepo Repository, userRepo user.Repository) error {
		if err := s.ensureNoExistingInstance(ctx, waRepo, phone, userId); err != nil {
			return err
		}
		user, err := s.fetchUser(ctx, userRepo, userId)
		if err != nil {
			return err
		}
		instanceName = s.buildInstanceName(user.Email, phone)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// Fase 2: cria instância na Evolution (fora da transação para evitar órfãos).
	remoteInstanceName, qr, err := s.createRemoteInstance(instanceName, phone)
	if err != nil {
		return nil, nil, err
	}
	if remoteInstanceName != "" {
		instanceName = remoteInstanceName
	}

	// Fase 3: persiste no banco.
	err = s.withTransaction(ctx, func(waRepo Repository, userRepo user.Repository) error {
		if err := s.ensureNoExistingInstance(ctx, waRepo, phone, userId); err != nil {
			return err
		}
		if err := s.persistInstance(ctx, waRepo, phone, userId, instanceName); err != nil {
			// Em caso de falha, idealmente deletar a instância remota.
			_ = deleteEvolutionInstance(remoteInstanceName)
			return err
		}
		var errFetch error
		instance, errFetch = waRepo.FindByPhoneAndUserId(ctx, phone, userId)
		if errFetch != nil {
			return fmt.Errorf("falha ao buscar instância criada: %w", errFetch)
		}
		if err := waRepo.Update(ctx, instance.ID, map[string]any{"connection_status": "connecting"}); err != nil {
			return fmt.Errorf("falha ao atualizar status da instância criada: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// Fase 4: conecta/gera QR atualizado (pairing/code) na Evolution
	connectResp, err := connectEvolutionInstance(instanceName, phone)
	if err != nil {
		// Se falhar, retornamos o QR da criação mesmo assim.
		return instance, &connectResponse{
			Code: qr,
			Qrcode: struct {
				Code   string `json:"code"`
				Base64 string `json:"base64"`
			}{Code: qr},
		}, nil
	}

	// Prioriza o QR retornado pelo connect (parsing code/count).
	return instance, connectResp, nil
}

func (s *service) GetInstances(ctx context.Context, userID string) ([]*Instance, error) {
	instances, err := s.whatsappInstanceRepository.FindAllByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		return []*Instance{}, nil
	}
	return instances, nil
}

func (s *service) DeleteInstance(ctx context.Context, userID, instanceID string) error {
	instance, err := s.whatsappInstanceRepository.FindByID(ctx, instanceID)
	if err != nil {
		return err
	}
	if instance == nil {
		return InstanceNotFound
	}
	if instance.UserID != userID {
		return InstanceForbidden
	}

	// Deleta na Evolution antes de remover localmente.
	if err := retryEvolution(ctx, 3, 500*time.Millisecond, func() error {
		return deleteEvolutionInstance(instance.InstanceName)
	}); err != nil {
		return fmt.Errorf("falha ao deletar instância na Evolution: %w", err)
	}

	return s.whatsappInstanceRepository.Delete(ctx, instanceID)
}

func (s *service) ConnectInstance(ctx context.Context, userID, instanceID string) (*connectResponse, error) {
	instance, err := s.ensureOwnership(ctx, userID, instanceID)
	if err != nil {
		return nil, err
	}
	resp, err := connectEvolutionInstance(instance.InstanceName, instance.Phone)
	return resp, err
}

func (s *service) ConnectionState(ctx context.Context, userID, instanceID string) (string, error) {
	instance, err := s.ensureOwnership(ctx, userID, instanceID)
	if err != nil {
		return "", err
	}
	state, err := connectionStateEvolution(instance.InstanceName)
	if err != nil {
		return "", err
	}
	if err := s.whatsappInstanceRepository.Update(ctx, instance.ID, map[string]any{"connection_status": state}); err != nil {
		return "", fmt.Errorf("falha ao atualizar status da instância: %w", err)
	}
	return state, nil
}

func (s *service) LogoutInstance(ctx context.Context, userID, instanceID string) error {
	instance, err := s.ensureOwnership(ctx, userID, instanceID)
	if err != nil {
		return err
	}
	return logoutEvolutionInstance(instance.InstanceName)
}

func (s *service) RestartInstance(ctx context.Context, userID, instanceID string) error {
	instance, err := s.ensureOwnership(ctx, userID, instanceID)
	if err != nil {
		return err
	}
	return restartEvolutionInstance(instance.InstanceName)
}

func (s *service) ensureNoExistingInstance(ctx context.Context, repo Repository, phone, userID string) error {
	hasInstance, err := repo.FindByPhoneAndUserId(ctx, phone, userID)
	if err != nil {
		return fmt.Errorf("falha ao verificar instância existente: %w", err)
	}
	if hasInstance != nil {
		return customerror.Trace("CreateInstance", HasAlreadyInstance)
	}
	return nil
}

func (s *service) fetchUser(ctx context.Context, repo user.Repository, userID string) (*user.User, error) {
	u, err := repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar usuário: %w", err)
	}
	if u == nil {
		return nil, customerror.Trace("CreateInstance", UserNotFound)
	}
	return u, nil
}

func (s *service) createRemoteInstance(instanceName, phone string) (string, string, error) {
	createdName, newQrCode, err := createEvolutionInstance(phone, instanceName, true)
	if err != nil {
		return "", "", customerror.Trace("CreateInstance", err)
	}
	return createdName, newQrCode, nil
}

func (s *service) persistInstance(ctx context.Context, repo Repository, phone, userID, instanceID string) error {
	if err := repo.Create(ctx, phone, userID, instanceID); err != nil {
		return fmt.Errorf("falha ao criar instância: %w", err)
	}
	return nil
}

func (s *service) withTransaction(ctx context.Context, fn func(repo Repository, userRepo user.Repository) error) error {
	_, err := database.MakeTransaction(ctx, []database.Transactional{s.whatsappInstanceRepository, s.userRepository}, func(txRepos []database.Transactional) (any, error) {
		waRepo := txRepos[0].(Repository)
		userRepo := txRepos[1].(user.Repository)
		return nil, fn(waRepo, userRepo)
	})
	return err
}

func (s *service) buildInstanceName(userEmail, phone string) string {
	return userEmail + ":" + phone
}

func retryEvolution(ctx context.Context, attempts int, delay time.Duration, op func() error) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := op(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if i == attempts-1 {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return lastErr
}

func (s *service) ensureOwnership(ctx context.Context, userID, instanceID string) (*Instance, error) {
	instance, err := s.whatsappInstanceRepository.FindByID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, InstanceNotFound
	}
	if instance.UserID != userID {
		return nil, InstanceForbidden
	}
	return instance, nil
}
