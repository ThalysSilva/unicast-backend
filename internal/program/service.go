package program

import (
	"context"
	"errors"
	"net/http"

	"github.com/ThalysSilva/unicast-backend/internal/campus"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type programService struct {
	programRepository Repository
	campusRepository  campus.Repository
}

var (
	ErrProgramAlreadyExists  = customerror.Make("o nome do curso já existe neste campus", http.StatusConflict, errors.New("ErrProgramAlreadyExists"))
	ErrProgramNotFound       = customerror.Make("o curso não foi encontrado", http.StatusNotFound, errors.New("ErrProgramNotFound"))
	ErrCampusNotFound        = customerror.Make("o campus não foi encontrado", http.StatusNotFound, errors.New("ErrCampusNotFound"))
	ErrCampusAccessForbidden = customerror.Make("você não tem permissão para acessar este campus", http.StatusForbidden, errors.New("ErrCampusAccessForbidden"))
)

type Service interface {
	Create(ctx context.Context, userID, campusID, name, description string, active bool) error
	GetProgram(id string) (*Program, error)
	GetProgramsByCampusID(ctx context.Context, userID, campusID string) ([]*Program, error)
	isOwner(ctx context.Context, programID, userID string) (bool, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, id string) error
}

func NewService(programRepository Repository, campusRepository campus.Repository) Service {
	return &programService{
		programRepository: programRepository,
		campusRepository:  campusRepository,
	}
}

func (s *programService) Create(ctx context.Context, userID, campusID, name, description string, active bool) error {
	if err := s.ensureCampusOwner(ctx, campusID, userID); err != nil {
		return err
	}

	program, err := s.programRepository.FindByNameAndCampusID(ctx, name, campusID)
	if err != nil {
		return err
	}
	if program != nil {
		return ErrProgramAlreadyExists
	}

	return s.programRepository.Create(ctx, name, description, campusID, active)
}

func (s *programService) GetProgram(id string) (*Program, error) {
	program, err := s.programRepository.FindByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if program == nil {
		return nil, nil
	}
	return program, nil
}

func (s *programService) GetProgramsByCampusID(ctx context.Context, userID, campusID string) ([]*Program, error) {
	if err := s.ensureCampusOwner(ctx, campusID, userID); err != nil {
		return nil, err
	}

	return s.programRepository.FindByCampusID(ctx, campusID)
}

func (s *programService) isOwner(ctx context.Context, programID, userID string) (bool, error) {
	program, err := s.programRepository.FindByIDWithUserOwnerID(ctx, programID)
	if err != nil {
		return false, err
	}
	if program == nil {
		return false, ErrProgramNotFound
	}
	if program.UserOwnerID != userID {
		return false, nil
	}

	return true, nil
}

func (s *programService) Update(ctx context.Context, id string, fields map[string]any) error {
	program, err := s.programRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if program == nil {
		return ErrProgramNotFound
	}
	if _, ok := fields["name"]; ok {
		existing, err := s.programRepository.FindByNameAndCampusID(ctx, fields["name"].(string), program.CampusID)
		if err != nil {
			return err
		}
		if existing != nil && existing.ID != id {
			return ErrProgramAlreadyExists
		}
	}

	return s.programRepository.Update(ctx, id, fields)
}

func (s *programService) Delete(ctx context.Context, id string) error {
	program, err := s.programRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if program == nil {
		return ErrProgramNotFound
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

func (s *programService) ensureCampusOwner(ctx context.Context, campusID, userID string) error {
	campus, err := s.campusRepository.FindByID(ctx, campusID)
	if err != nil {
		return err
	}
	if campus == nil {
		return ErrCampusNotFound
	}
	if campus.UserOwnerID != userID {
		return ErrCampusAccessForbidden
	}

	return nil
}
