package whatsappinstance

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
	whatsAppRepository Repository
	userRepository     user.Repository
}

func NewService(whatsappRepository Repository, userRepository user.Repository) Service {
	return &service{
		whatsAppRepository: whatsappRepository,
		userRepository:     userRepository,
	}
}

var (
	HasAlreadyInstance = customerror.Make("Instância já existe", http.StatusConflict, errors.New("HasAlreadyInstance"))
	UserNotFound       = customerror.Make("Usuário não encontrado", http.StatusNotFound, errors.New("userNotFound"))
)

func (s *service) CreateInstance(userId, phone string) (instance *Instance, qrCode string, err error) {
	hasInstance, err := s.whatsAppRepository.FindByPhoneAndUserId(phone, userId)
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

	err = s.whatsAppRepository.Create(phone, userId, instanceId)
	if err != nil {
		return nil, "", err
	}

	instance, err = s.whatsAppRepository.FindByPhoneAndUserId(phone, userId)
	if err != nil {
		return nil, "", err
	}
	return instance, newQrCode, nil
}

func (s *service) GetInstances(userID string) ([]*Instance, error) {
	instances, err := s.whatsAppRepository.FindAllByUserId(userID)
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		return []*Instance{}, nil
	}
	return instances, nil
}
