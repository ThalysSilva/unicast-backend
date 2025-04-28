package course

import (
	"context"
	"errors"
	"net/http"

	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type courseService struct {
	courseRepository Repository
}

var (
	ErrCampusAlreadyExists = customerror.Make("o nome do campus já existe", http.StatusConflict, errors.New("ErrCampusAlreadyExists"))
	ErrCampusNotFound      = customerror.Make("o campus não foi encontrado", http.StatusNotFound, errors.New("ErrCampusNotFound"))
)

type Service interface {
	Create(ctx context.Context, userID, name, description string, year, semester int) error
	GetCourse(id string) (*Course, error)
	GetCourses(ctx context.Context, userID string) ([]*Course, error)
	isOwner(ctx context.Context, courseID, userID string) (bool, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, id string) error
}

func NewService(courseRepository Repository) Service {
	return &courseService{courseRepository: courseRepository}
}

func (s *courseService) Create(ctx context.Context, ProgramID, name, description string, year, semester int) error {
	campus, err := s.courseRepository.FindByNameAndProgramID(ctx, name, ProgramID)
	if err != nil {
		return err
	}
	if campus != nil {
		return ErrCampusAlreadyExists
	}

	return s.courseRepository.Create(ctx, name, description, ProgramID, year, semester)
}

func (s *courseService) GetCourse(id string) (*Course, error) {
	campus, err := s.courseRepository.FindByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if campus == nil {
		return nil, nil
	}
	return campus, nil
}

func (s *courseService) GetCourses(ctx context.Context, ProgramID string) ([]*Course, error) {
	return s.courseRepository.FindByProgramID(ctx, ProgramID)
}

func (s *courseService) isOwner(ctx context.Context, courseID, userID string) (bool, error) {
	campus, err := s.courseRepository.FindByIDWithUserOwnerID(ctx, courseID)
	if err != nil {
		return false, err
	}
	if campus == nil {
		return false, ErrCampusNotFound
	}
	if campus.UserOwnerID != userID {
		return false, nil
	}
	
	return true, nil
}

func (s *courseService) Update(ctx context.Context, id string, fields map[string]any) error {
	course, err := s.courseRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if course == nil {
		return ErrCampusNotFound
	}
	if _, ok := fields["name"]; ok {
		campus, err := s.courseRepository.FindByNameAndProgramID(ctx, fields["name"].(string), course.ProgramID)
		if err != nil {
			return err
		}
		if campus != nil {
			return ErrCampusAlreadyExists
		}
	}

	return s.courseRepository.Update(ctx, id, fields)
}

func (s *courseService) Delete(ctx context.Context, id string) error {
	campus, err := s.courseRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if campus == nil {
		return ErrCampusNotFound
	}

	_, err = database.MakeTransaction(ctx, []database.Transactional{s.courseRepository}, func() (any, error) {
		err := s.courseRepository.Delete(ctx, id)
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
