package smtp

import "context"

type smtpService struct {
	repository Repository
}

type Service interface {
	Create(userId, email, password, host string, port int, iv []byte) error
	GetInstances(userID string) ([]*Instance, error)
}

func NewService(repository Repository) Service {
	return &smtpService{repository: repository}
}

func (s *smtpService) Create(userId, email, password, host string, port int, iv []byte) error {
	ctx := context.Background()
	return s.repository.Create(ctx, userId, email, password, host, port, iv)
}

func (s *smtpService) GetInstances(userID string) ([]*Instance, error) {
	ctx := context.Background()
	return s.repository.GetInstances(ctx,  userID)
}
