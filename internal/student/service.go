package student

import (
	"context"
	"errors"
	"net/http"

	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
)

type Service interface {
	Create(ctx context.Context, studentID string, name, phone, email, annotation *string, status StudentStatus) error
	GetStudent(ctx context.Context, id string) (*Student, error)
	GetStudents(ctx context.Context, filters map[string]string) ([]*Student, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, id string) error
}

type studentService struct {
	studentRepository Repository
}

var (
	ErrStudentNotFound = customerror.Make("o campus não foi encontrado", http.StatusNotFound, errors.New("ErrCampusNotFound"))
)

func NewService(studentRepository Repository) Service {
	return &studentService{
		studentRepository: studentRepository,
	}
}

func (s *studentService) Create(ctx context.Context, studentID string, name, phone, email, annotation *string, status StudentStatus) error {

	err := s.studentRepository.Create(ctx, studentID, name, phone, email, annotation, status)
	if err != nil {
		return err
	}

	return nil
}

func (s *studentService) GetStudent(ctx context.Context, id string) (*Student, error) {
	student, err := s.studentRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, nil
	}
	return student, nil
}

func (s *studentService) GetStudents(ctx context.Context, filters map[string]string) ([]*Student, error) {
	if filters == nil {
		filters = make(map[string]string)
	}
	students, err := s.studentRepository.FindByFilters(ctx, filters)
	if err != nil {
		return nil, err
	}
	if students == nil {
		return nil, nil
	}
	return students, nil
}

func (s *studentService) Update(ctx context.Context, id string, fields map[string]any) error {
	student, err := s.studentRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if student == nil {
		return ErrStudentNotFound
	}

	err = s.studentRepository.Update(ctx, id, fields)
	if err != nil {
		return err
	}

	return nil
}

func (s *studentService) Delete(ctx context.Context, id string) error {
	student, err := s.studentRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if student == nil {
		return ErrStudentNotFound
	}

	err = s.studentRepository.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
