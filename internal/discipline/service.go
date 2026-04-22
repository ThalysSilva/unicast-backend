package discipline

import (
	"context"
	"errors"
	"net/http"

	"github.com/ThalysSilva/unicast-backend/internal/program"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type disciplineService struct {
	disciplineRepository Repository
	programRepository    program.Repository
}

var (
	ErrDisciplineAlreadyExists  = customerror.Make("o nome da disciplina já existe neste curso", http.StatusConflict, errors.New("ErrDisciplineAlreadyExists"))
	ErrDisciplineNotFound       = customerror.Make("a disciplina não foi encontrada", http.StatusNotFound, errors.New("ErrDisciplineNotFound"))
	ErrProgramNotFound          = customerror.Make("o curso não foi encontrado", http.StatusNotFound, errors.New("ErrProgramNotFound"))
	ErrProgramAccessForbidden   = customerror.Make("você não tem permissão para acessar este curso", http.StatusForbidden, errors.New("ErrProgramAccessForbidden"))
)

type Service interface {
	Create(ctx context.Context, userID, programID, name, description string, year, semester int) error
	GetDiscipline(id string) (*Discipline, error)
	GetDisciplinesByProgramID(ctx context.Context, userID, programID string) ([]*Discipline, error)
	GetDisciplinesByUserID(ctx context.Context, userID string) ([]*Discipline, error)
	isOwner(ctx context.Context, disciplineID, userID string) (bool, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, id string) error
}

func NewService(disciplineRepository Repository, programRepository program.Repository) Service {
	return &disciplineService{
		disciplineRepository: disciplineRepository,
		programRepository:    programRepository,
	}
}

func (s *disciplineService) Create(ctx context.Context, userID, programID, name, description string, year, semester int) error {
	if err := s.ensureProgramOwner(ctx, programID, userID); err != nil {
		return err
	}

	discipline, err := s.disciplineRepository.FindByNameAndProgramID(ctx, name, programID)
	if err != nil {
		return err
	}
	if discipline != nil {
		return ErrDisciplineAlreadyExists
	}

	return s.disciplineRepository.Create(ctx, name, description, programID, year, semester)
}

func (s *disciplineService) GetDiscipline(id string) (*Discipline, error) {
	discipline, err := s.disciplineRepository.FindByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if discipline == nil {
		return nil, nil
	}
	return discipline, nil
}

func (s *disciplineService) GetDisciplinesByProgramID(ctx context.Context, userID, programID string) ([]*Discipline, error) {
	if err := s.ensureProgramOwner(ctx, programID, userID); err != nil {
		return nil, err
	}

	return s.disciplineRepository.FindByProgramID(ctx, programID)
}

func (s *disciplineService) GetDisciplinesByUserID(ctx context.Context, userID string) ([]*Discipline, error) {
	return s.disciplineRepository.FindByUserOwnerID(ctx, userID)
}

func (s *disciplineService) isOwner(ctx context.Context, disciplineID, userID string) (bool, error) {
	discipline, err := s.disciplineRepository.FindByIDWithUserOwnerID(ctx, disciplineID)
	if err != nil {
		return false, err
	}
	if discipline == nil {
		return false, ErrDisciplineNotFound
	}
	if discipline.UserOwnerID != userID {
		return false, nil
	}

	return true, nil
}

func (s *disciplineService) Update(ctx context.Context, id string, fields map[string]any) error {
	discipline, err := s.disciplineRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if discipline == nil {
		return ErrDisciplineNotFound
	}
	if _, ok := fields["name"]; ok {
		existing, err := s.disciplineRepository.FindByNameAndProgramID(ctx, fields["name"].(string), discipline.ProgramID)
		if err != nil {
			return err
		}
		if existing != nil {
			return ErrDisciplineAlreadyExists
		}
	}

	return s.disciplineRepository.Update(ctx, id, fields)
}

func (s *disciplineService) Delete(ctx context.Context, id string) error {
	discipline, err := s.disciplineRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if discipline == nil {
		return ErrDisciplineNotFound
	}

	_, err = database.MakeTransaction(ctx, []database.Transactional{s.disciplineRepository}, func(txRepos []database.Transactional) (any, error) {
		repo := txRepos[0].(Repository)
		err := repo.Delete(ctx, id)
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

func (s *disciplineService) ensureProgramOwner(ctx context.Context, programID, userID string) error {
	programFound, err := s.programRepository.FindByIDWithUserOwnerID(ctx, programID)
	if err != nil {
		return err
	}
	if programFound == nil {
		return ErrProgramNotFound
	}
	if programFound.UserOwnerID != userID {
		return ErrProgramAccessForbidden
	}

	return nil
}
