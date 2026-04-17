package student

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ThalysSilva/unicast-backend/internal/enrollment"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
)

type ImportMode string

const (
	ImportModeClean  ImportMode = "clean"
	ImportModeUpsert ImportMode = "upsert"
)

type ImportRecord struct {
	StudentID string
	Name      *string
	Phone     *string
	Email     *string
	Status    StudentStatus
}

type ImportResult struct {
	Inserted         int
	Updated          int
	EnrollmentsAdded int
	Errors           []string
}

var ErrEnrollmentNotFound = customerror.Make("vínculo com disciplina não encontrado", http.StatusNotFound, errors.New("ErrEnrollmentNotFound"))

type ImportService interface {
	ImportForDiscipline(ctx context.Context, disciplineID string, mode ImportMode, records []ImportRecord) (*ImportResult, error)
	AddStudentToDiscipline(ctx context.Context, disciplineID, studentID string) error
	RemoveStudentFromDiscipline(ctx context.Context, disciplineID, studentUUID string) error
}

type importService struct {
	studentsRepo   Repository
	enrollmentRepo enrollment.Repository
}

func NewImportService(studentsRepo Repository, enrollmentRepo enrollment.Repository) ImportService {
	return &importService{studentsRepo: studentsRepo, enrollmentRepo: enrollmentRepo}
}

func (s *importService) ImportForDiscipline(ctx context.Context, disciplineID string, mode ImportMode, records []ImportRecord) (*ImportResult, error) {
	result := &ImportResult{}

	if mode == ImportModeClean {
		if err := s.enrollmentRepo.DeleteByDisciplineID(ctx, disciplineID); err != nil {
			return nil, fmt.Errorf("falha ao limpar matrículas da disciplina: %w", err)
		}
	}

	for idx, rec := range records {
		if err := s.processRecord(ctx, disciplineID, idx, rec, result); err != nil {
			result.Errors = append(result.Errors, err.Error())
			continue
		}
	}

	return result, nil
}

func (s *importService) AddStudentToDiscipline(ctx context.Context, disciplineID, studentID string) error {
	result := &ImportResult{}
	return s.processRecord(ctx, disciplineID, 0, ImportRecord{
		StudentID: studentID,
		Status:    StudentStatusPending,
	}, result)
}

func (s *importService) RemoveStudentFromDiscipline(ctx context.Context, disciplineID, studentUUID string) error {
	enroll, err := s.enrollmentRepo.FindByDisciplineAndStudent(ctx, disciplineID, studentUUID)
	if err != nil {
		return fmt.Errorf("erro ao verificar vínculo: %w", err)
	}
	if enroll == nil {
		return ErrEnrollmentNotFound
	}

	if err := s.enrollmentRepo.Delete(ctx, enroll.ID); err != nil {
		return fmt.Errorf("erro ao desvincular matrícula: %w", err)
	}
	return nil
}

func (s *importService) processRecord(ctx context.Context, disciplineID string, idx int, rec ImportRecord, result *ImportResult) error {
	if rec.StudentID == "" {
		return fmt.Errorf("linha %d: studentId vazio", idx+1)
	}

	existing, err := s.studentsRepo.FindByStudentID(ctx, rec.StudentID)
	if err != nil {
		return fmt.Errorf("linha %d: erro ao buscar student: %v", idx+1, err)
	}

	if existing == nil {
		if err := s.studentsRepo.Create(ctx, rec.StudentID, nil, nil, nil, nil, StudentStatusPending); err != nil {
			return fmt.Errorf("linha %d: erro ao criar student: %v", idx+1, err)
		}
		existing, err = s.studentsRepo.FindByStudentID(ctx, rec.StudentID)
		if err != nil {
			return fmt.Errorf("linha %d: erro ao buscar student criado: %v", idx+1, err)
		}
		if existing == nil {
			return fmt.Errorf("linha %d: student criado nao foi encontrado", idx+1)
		}
		result.Inserted++
	} else {
		if err := s.updateExisting(ctx, rec, existing, idx, result); err != nil {
			return err
		}
	}

	if err := s.ensureEnrollment(ctx, disciplineID, existing.ID, idx, result); err != nil {
		return err
	}

	return nil
}

func (s *importService) updateExisting(ctx context.Context, rec ImportRecord, existing *Student, idx int, result *ImportResult) error {
	canUpdate := existing.Phone != nil || existing.Email != nil
	status := rec.Status
	if status == StudentStatusPending && hasCompletedContact(existing) {
		status = StudentStatusActive
	}

	fields := map[string]any{
		"status": status,
	}

	if canUpdate {
		if rec.Name != nil {
			fields["name"] = *rec.Name
		}
		if rec.Phone != nil {
			fields["phone"] = *rec.Phone
		}
		if rec.Email != nil {
			fields["email"] = *rec.Email
		}
	} else if rec.Phone != nil || rec.Email != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("linha %d: aluno %s ainda não completou cadastro; ignorei contato, mas atualizei status", idx+1, rec.StudentID))
	}

	if err := s.studentsRepo.Update(ctx, existing.ID, fields); err != nil {
		return fmt.Errorf("linha %d: erro ao atualizar student: %v", idx+1, err)
	}
	result.Updated++
	return nil
}

func hasCompletedContact(student *Student) bool {
	if student == nil {
		return false
	}

	return student.Name != nil && student.Email != nil && student.Phone != nil
}

func (s *importService) ensureEnrollment(ctx context.Context, disciplineID, studentUUID string, idx int, result *ImportResult) error {
	enroll, err := s.enrollmentRepo.FindByDisciplineAndStudent(ctx, disciplineID, studentUUID)
	if err != nil {
		return fmt.Errorf("linha %d: erro ao verificar enrollment: %v", idx+1, err)
	}
	if enroll != nil {
		return nil
	}

	if err := s.enrollmentRepo.Create(ctx, disciplineID, studentUUID); err != nil {
		return fmt.Errorf("linha %d: erro ao criar enrollment: %v", idx+1, err)
	}
	result.EnrollmentsAdded++
	return nil
}
