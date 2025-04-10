package whatsapp

import (
	"net/http"

	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
)

type Service interface {
	CreateInstance(userId, Phone, instanceName string) (instance *Instance, qrCode string, err error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

func (s *service) CreateInstance(userId, phone, instanceName string) (instance *Instance, qrCode string, err error) {
	hasInstance, err := s.repository.FindByPhoneAndUserId(phone, instanceName)
	if err != nil {
		return nil, "", err
	}
	if hasInstance != nil {
		return nil, "", customerror.Make("Instância já existe", http.StatusConflict)
	}

	instanceId, newQrCode, err := createEvolutionInstance(phone, instanceName, true)
	if err != nil {
		err := customerror.Make("Falha ao criar instância", http.StatusInternalServerError)
		return nil, "", customerror.Trace("CreateInstance", err)
	}

	err = s.repository.Create(phone, instanceName, userId, instanceId)
	if err != nil {
		return nil, "", err
	}

	instance, err = s.repository.FindByPhoneAndUserId(phone, userId)
	if err != nil {
		return nil, "", err
	}
	return instance, newQrCode, nil
}
