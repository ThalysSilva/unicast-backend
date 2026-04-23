package student

import (
	"context"
	"errors"
	"net/http"

	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
)

type Service interface {
	Create(ctx context.Context, userID, studentID string) error
	GetStudent(ctx context.Context, userID, id string) (*Student, error)
	GetStudents(ctx context.Context, userID string, filters map[string]string) ([]*Student, error)
	Update(ctx context.Context, userID, id string, fields map[string]any) error
	Delete(ctx context.Context, userID, id string) error
}

type studentService struct {
	studentRepository Repository
}

var (
	ErrStudentNotFound = customerror.Make("aluno não encontrado", http.StatusNotFound, errors.New("ErrStudentNotFound"))
)

func NewService(studentRepository Repository) Service {
	return &studentService{
		studentRepository: studentRepository,
	}
}

func (s *studentService) Create(ctx context.Context, userID, studentID string) error {

	err := s.studentRepository.Create(ctx, userID, studentID, nil, nil, nil, nil, false, StudentStatusPending)
	if err != nil {
		return err
	}

	return nil
}

func (s *studentService) GetStudent(ctx context.Context, userID, id string) (*Student, error) {
	student, err := s.studentRepository.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, nil
	}
	return student, nil
}

func (s *studentService) GetStudents(ctx context.Context, userID string, filters map[string]string) ([]*Student, error) {
	if filters == nil {
		filters = make(map[string]string)
	}
	filters["user"] = userID
	students, err := s.studentRepository.FindByFilters(ctx, filters)
	if err != nil {
		return nil, err
	}
	if students == nil {
		return nil, nil
	}
	return students, nil
}

func (s *studentService) Update(ctx context.Context, userID, id string, fields map[string]any) error {
	student, err := s.studentRepository.FindByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if student == nil {
		return ErrStudentNotFound
	}

	name := mergeFieldString(student.Name, fields["name"])
	phone := mergeFieldString(student.Phone, fields["phone"])
	email := mergeFieldString(student.Email, fields["email"])
	noPhone := student.NoPhone
	if value, ok := fields["no_phone"].(bool); ok {
		noPhone = value
	}

	if noPhone {
		fields["phone"] = nil
		phone = nil
	}
	if phone != nil && *phone != "" {
		fields["no_phone"] = false
		noPhone = false
	}
	status, statusProvided := fields["status"].(StudentStatus)
	fields["status"] = DeriveContactAwareStatus(student.Status, status, statusProvided, name, phone, email, noPhone)

	err = s.studentRepository.Update(ctx, id, fields)
	if err != nil {
		return err
	}

	return nil
}

func mergeFieldString(current *string, next any) *string {
	switch value := next.(type) {
	case *string:
		if value != nil {
			return value
		}
	case string:
		return &value
	}
	return current
}

func (s *studentService) Delete(ctx context.Context, userID, id string) error {
	student, err := s.studentRepository.FindByID(ctx, id, userID)
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
