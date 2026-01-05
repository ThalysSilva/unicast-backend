package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ThalysSilva/unicast-backend/internal/user"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Service interface {
	CreateInstance(ctx context.Context, userId, phone string) (*Instance, string, error)
	GetInstances(ctx context.Context, userID string) ([]*Instance, error)
	DeleteInstance(ctx context.Context, userID, instanceID string) error
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
)

func (s *service) CreateInstance(ctx context.Context, userId, phone string) (*Instance, string, error) {
	var instance *Instance

	// Fase 1: checa existência e usuário dentro da transação.
	var instanceId string
	err := s.withTransaction(ctx, func(waRepo Repository, userRepo user.Repository) error {
		if err := s.ensureNoExistingInstance(ctx, waRepo, phone, userId); err != nil {
			return err
		}
		user, err := s.fetchUser(ctx, userRepo, userId)
		if err != nil {
			return err
		}
		instanceId = s.buildInstanceName(user.Email, phone)
		return nil
	})
	if err != nil {
		return nil, "", err
	}

	// Fase 2: cria instância na Evolution (fora da transação para evitar órfãos).
	remoteInstanceId, qr, err := s.createRemoteInstance(phone, instanceId)
	if err != nil {
		return nil, "", err
	}

	// Fase 3: persiste no banco.
	err = s.withTransaction(ctx, func(waRepo Repository, userRepo user.Repository) error {
		if err := s.ensureNoExistingInstance(ctx, waRepo, phone, userId); err != nil {
			return err
		}
		if err := s.persistInstance(ctx, waRepo, phone, userId, remoteInstanceId); err != nil {
			// Em caso de falha, idealmente deletar a instância remota.
			_ = deleteEvolutionInstance(remoteInstanceId)
			return err
		}
		var errFetch error
		instance, errFetch = waRepo.FindByPhoneAndUserId(ctx, phone, userId)
		if errFetch != nil {
			return fmt.Errorf("falha ao buscar instância criada: %w", errFetch)
		}
		return nil
	})
	if err != nil {
		return nil, "", err
	}

	return instance, qr, nil
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

	return s.whatsappInstanceRepository.Delete(ctx, instanceID)
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

func (s *service) createRemoteInstance(phone, userEmail string) (string, string, error) {
	instanceName := userEmail + ":" + phone
	instanceId, newQrCode, err := createEvolutionInstance(phone, instanceName, true)
	if err != nil {
		return "", "", customerror.Trace("CreateInstance", err)
	}
	return instanceId, newQrCode, nil
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
