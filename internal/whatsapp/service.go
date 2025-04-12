package whatsapp

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ThalysSilva/unicast-backend/internal/user"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
)

type Service interface {
	CreateInstance(userId, Phone string) (instance *Instance, qrCode string, err error)
	GetInstances(userID string) ([]*Instance, error)
}

type service struct {
	whatsAppInstanceRepository Repository
	userRepository             user.Repository
}

func NewService(whatsappInstanceRepository Repository, userRepository user.Repository) Service {
	return &service{
		whatsAppInstanceRepository: whatsappInstanceRepository,
		userRepository:             userRepository,
	}
}

var (
	HasAlreadyInstance = customerror.Make("Instância já existe", http.StatusConflict, errors.New("HasAlreadyInstance"))
	UserNotFound       = customerror.Make("Usuário não encontrado", http.StatusNotFound, errors.New("userNotFound"))
)

func (s *service) CreateInstance(userId, phone string) (instance *Instance, qrCode string, err error) {
	hasInstance, err := s.whatsAppInstanceRepository.FindByPhoneAndUserId(phone, userId)
	if err != nil {
		return nil, "", err
	}
	if hasInstance != nil {
		return nil, "", customerror.Trace("CreateInstance", HasAlreadyInstance)
	}
	user, err := s.userRepository.FindByID(userId)
	if err != nil {
		return nil, "", err
	}
	if user == nil {

		return nil, "", customerror.Trace("CreateInstance", UserNotFound)
	}

	instanceName := user.Email + ":" + phone

	instanceId, newQrCode, err := createEvolutionInstance(phone, instanceName, true)
	if err != nil {
		return nil, "", customerror.Trace("CreateInstance", err)
	}
	fmt.Println("instanceId", instanceId)

	err = s.whatsAppInstanceRepository.Create(phone, userId, instanceId)
	if err != nil {
		return nil, "", err
	}

	instance, err = s.whatsAppInstanceRepository.FindByPhoneAndUserId(phone, userId)
	if err != nil {
		return nil, "", err
	}
	return instance, newQrCode, nil
}

func (s *service) GetInstances(userID string) ([]*Instance, error) {
	instances, err := s.whatsAppInstanceRepository.FindAllByUserId(userID)
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		return []*Instance{}, nil
	}
	return instances, nil
}
