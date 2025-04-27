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
)

func (s *service) CreateInstance(ctx context.Context, userId, phone string) (*Instance, string, error) {
	var instance *Instance
	var qrCode string

	_, err := database.MakeTransaction(ctx, []database.Transactional{s.whatsappInstanceRepository, s.userRepository}, func() (any, error) {
		hasInstance, err := s.whatsappInstanceRepository.FindByPhoneAndUserId(ctx, phone, userId)
		if err != nil {
			return nil, fmt.Errorf("falha ao verificar instância existente: %w", err)
		}
		if hasInstance != nil {
			return nil, customerror.Trace("CreateInstance", HasAlreadyInstance)
		}

		user, err := s.userRepository.FindByID(ctx, userId)
		if err != nil {
			return nil, fmt.Errorf("falha ao buscar usuário: %w", err)
		}
		if user == nil {
			return nil, customerror.Trace("CreateInstance", UserNotFound)
		}

		instanceName := user.Email + ":" + phone
		instanceId, newQrCode, err := createEvolutionInstance(phone, instanceName, true)
		if err != nil {
			return nil, customerror.Trace("CreateInstance", err)
		}

		err = s.whatsappInstanceRepository.Create(ctx, phone, userId, instanceId)
		if err != nil {
			// deleteEvolutionInstance(instanceId)
			return nil, fmt.Errorf("falha ao criar instância: %w", err)
		}

		instance, err = s.whatsappInstanceRepository.FindByPhoneAndUserId(ctx, phone, userId)
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
