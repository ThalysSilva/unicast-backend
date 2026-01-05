package campus

import (
	"context"
	"errors"
	"net/http"

	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type campusService struct {
	campusRepository Repository
}

var (
	ErrCampusAlreadyExists = customerror.Make("o nome do campus já existe", http.StatusConflict, errors.New("ErrCampusAlreadyExists"))
	ErrCampusNotFound      = customerror.Make("o campus não foi encontrado", http.StatusNotFound, errors.New("ErrCampusNotFound"))
)

type Service interface {
	Create(ctx context.Context, userID, name, description string) error
	GetCampus(id string) (*Campus, error)
	GetCampuses(ctx context.Context, userID string) ([]*Campus, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, id string) error
}

func NewService(campusRepository Repository) Service {
	return &campusService{campusRepository: campusRepository}
}

func (s *campusService) Create(ctx context.Context, userID, name, description string) error {
	campus, err := s.campusRepository.FindByNameAndUserOwnerID(ctx, name, userID)
	if err != nil {
		return err
	}
	if campus != nil {
		return ErrCampusAlreadyExists
	}

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
	campus, err := s.campusRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if campus == nil {
		return ErrCampusNotFound
	}
	if _, ok := fields["name"]; ok {
		campus, err := s.campusRepository.FindByNameAndUserOwnerID(ctx, fields["name"].(string), campus.UserOwnerID)
		if err != nil {
			return err
		}
		if campus != nil {
			return ErrCampusAlreadyExists
		}
	}

	return s.campusRepository.Update(ctx, id, fields)
}

func (s *campusService) Delete(ctx context.Context, id string) error {
	campus, err := s.campusRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if campus == nil {
		return ErrCampusNotFound
	}

	_, err = database.MakeTransaction(ctx, []database.Transactional{s.campusRepository}, func(txRepos []database.Transactional) (any, error) {
		repo := txRepos[0].(Repository)
		err := repo.Delete(ctx, id)
		// TODO: Implementar a exclusão de campus e todas as suas dependências
		if err != nil {
			return nil, err
		}
		return nil, nil
	},
	)
	if err != nil {
		return err
	}
	return nil

}
