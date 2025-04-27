package campus

import (
	"context"
)

type campusService struct {
	campusRepository Repository
}

type Service interface {
	Create(ctx context.Context, userID, name, description string) error
	GetCampuses(ctx context.Context, userID string) ([]*Campus, error)
}

func NewService(campusRepository Repository) Service {
	return &campusService{campusRepository: campusRepository}
}

func (s *campusService) Create(ctx context.Context, userID, name, description string) error {
	return s.campusRepository.Create(ctx, name, description, userID)
}

func (s *campusService) GetCampuses(ctx context.Context, userID string) ([]*Campus, error) {
	return s.campusRepository.FindByUserOwnerId(ctx, userID)
}
