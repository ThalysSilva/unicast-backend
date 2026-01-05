package program

import (
	"context"
	"errors"
	"net/http"

	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type programService struct {
	programRepository Repository
}

var (
	ErrCampusAlreadyExists = customerror.Make("o nome do campus já existe", http.StatusConflict, errors.New("ErrCampusAlreadyExists"))
	ErrCampusNotFound      = customerror.Make("o campus não foi encontrado", http.StatusNotFound, errors.New("ErrCampusNotFound"))
)

type Service interface {
	Create(ctx context.Context, CampusID, name, description string, active bool) error
	GetProgram(id string) (*Program, error)
	GetProgramsByCampusID(ctx context.Context, campusID string) ([]*Program, error)
	isOwner(ctx context.Context, ProgramID, userID string) (bool, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, id string) error
}

func NewService(programRepository Repository) Service {
	return &programService{programRepository: programRepository}
}

func (s *programService) Create(ctx context.Context, campusID, name, description string, active bool) error {
	campus, err := s.programRepository.FindByNameAndCampusID(ctx, name, campusID)
	if err != nil {
		return err
	}
	if campus != nil {
		return ErrCampusAlreadyExists
	}

	return s.programRepository.Create(ctx, name, description, campusID, active)
}

func (s *programService) GetProgram(id string) (*Program, error) {
	campus, err := s.programRepository.FindByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if campus == nil {
		return nil, nil
	}
	return campus, nil
}

func (s *programService) GetProgramsByCampusID(ctx context.Context, campusID string) ([]*Program, error) {
	return s.programRepository.FindByCampusID(ctx, campusID)
}

func (s *programService) isOwner(ctx context.Context, ProgramID, userID string) (bool, error) {
	program, err := s.programRepository.FindByIDWithUserOwnerID(ctx, ProgramID)
	if err != nil {
		return false, err
	}
	if program == nil {
		return false, ErrCampusNotFound
	}
	if program.UserOwnerID != userID {
		return false, nil
	}

	return true, nil
}

func (s *programService) Update(ctx context.Context, id string, fields map[string]any) error {
	Program, err := s.programRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if Program == nil {
		return ErrCampusNotFound
	}
	if _, ok := fields["name"]; ok {
		campus, err := s.programRepository.FindByNameAndCampusID(ctx, fields["name"].(string), Program.CampusID)
		if err != nil {
			return err
		}
		if campus != nil {
			return ErrCampusAlreadyExists
		}
	}

	return s.programRepository.Update(ctx, id, fields)
}

func (s *programService) Delete(ctx context.Context, id string) error {
	campus, err := s.programRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if campus == nil {
		return ErrCampusNotFound
	}

	_, err = database.MakeTransaction(ctx, []database.Transactional{s.programRepository}, func(txRepos []database.Transactional) (any, error) {
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
