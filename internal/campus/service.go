package campus

import (
	"context"
)

type campusService struct {
	campusRepository Repository
}

type Service interface {
	Create(ctx context.Context, userID, name, description string) error
	GetCampus(id string) (*Campus, error)
	GetCampuses(ctx context.Context, userID string) ([]*Campus, error)
	Update(ctx context.Context, id string, fields map[string]any) error
}

func NewService(campusRepository Repository) Service {
	return &campusService{campusRepository: campusRepository}
}

func (s *campusService) Create(ctx context.Context, userID, name, description string) error {
	return s.campusRepository.Create(ctx, name, description, userID)
}

func (s *campusService) GetCampus(id string) (*Campus, error) {
	campus, err := s.campusRepository.FindByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if campus == nil {
		return nil, nil
	}
	return campus, nil
}

func (s *campusService) GetCampuses(ctx context.Context, userID string) ([]*Campus, error) {
	return s.campusRepository.FindByUserOwnerId(ctx, userID)
}

func (s *campusService) Update(ctx context.Context, id string, fields map[string]any) error {
	return s.campusRepository.Update(ctx, id, fields)
}
