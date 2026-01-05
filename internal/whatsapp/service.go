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
	var qrCode string

	_, err := database.MakeTransaction(ctx, []database.Transactional{s.whatsappInstanceRepository, s.userRepository}, func(txRepos []database.Transactional) (any, error) {
		waRepo := txRepos[0].(Repository)
		userRepo := txRepos[1].(user.Repository)

		if err := s.ensureNoExistingInstance(ctx, waRepo, phone, userId); err != nil {
			return nil, err
		}

		user, err := s.fetchUser(ctx, userRepo, userId)
		if err != nil {
			return nil, err
		}

		instanceId, newQrCode, err := s.createRemoteInstance(phone, user.Email)
		if err != nil {
			return nil, err
		}

		if err := s.persistInstance(ctx, waRepo, phone, userId, instanceId); err != nil {
			return nil, err
		}

		instance, err = waRepo.FindByPhoneAndUserId(ctx, phone, userId)
		if err != nil {
			return nil, fmt.Errorf("falha ao buscar instância criada: %w", err)
		}

		qrCode = newQrCode
		return nil, nil
	})

	if err != nil {
		return nil, "", err
	}

	// Verifica contexto normal (*sql.DB)
	_, err = s.whatsappInstanceRepository.FindByID(ctx, "non-existent")
	if err != nil {
		fmt.Printf("Pós-transação usa *sql.DB: %v\n", err)
	}

	return instance, qrCode, nil
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
