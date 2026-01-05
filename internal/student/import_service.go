package student

import (
	"context"
	"fmt"

	"github.com/ThalysSilva/unicast-backend/internal/enrollment"
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

type ImportService interface {
	ImportForCourse(ctx context.Context, courseID string, mode ImportMode, records []ImportRecord) (*ImportResult, error)
}

type importService struct {
	studentsRepo   Repository
	enrollmentRepo enrollment.Repository
}

func NewImportService(studentsRepo Repository, enrollmentRepo enrollment.Repository) ImportService {
	return &importService{studentsRepo: studentsRepo, enrollmentRepo: enrollmentRepo}
}

func (s *importService) ImportForCourse(ctx context.Context, courseID string, mode ImportMode, records []ImportRecord) (*ImportResult, error) {
	result := &ImportResult{}

	if mode == ImportModeClean {
		if err := s.enrollmentRepo.DeleteByCourseID(ctx, courseID); err != nil {
			return nil, fmt.Errorf("falha ao limpar matrículas do curso: %w", err)
		}
	}

	for idx, rec := range records {
		if err := s.processRecord(ctx, courseID, idx, rec, result); err != nil {
			result.Errors = append(result.Errors, err.Error())
			continue
		}
	}

	return result, nil
}

func (s *importService) processRecord(ctx context.Context, courseID string, idx int, rec ImportRecord, result *ImportResult) error {
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
		result.Inserted++
	} else {
		if err := s.updateExisting(ctx, rec, existing, idx, result); err != nil {
			return err
		}
	}

	if err := s.ensureEnrollment(ctx, courseID, rec.StudentID, idx, result); err != nil {
		return err
	}

	return nil
}

func (s *importService) updateExisting(ctx context.Context, rec ImportRecord, existing *Student, idx int, result *ImportResult) error {
	canUpdate := existing.Phone != nil || existing.Email != nil

	fields := map[string]any{
		"status": rec.Status,
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

func (s *importService) ensureEnrollment(ctx context.Context, courseID, studentID string, idx int, result *ImportResult) error {
	enroll, err := s.enrollmentRepo.FindByCourseAndStudent(ctx, courseID, studentID)
	if err != nil {
		return fmt.Errorf("linha %d: erro ao verificar enrollment: %v", idx+1, err)
	}
	if enroll != nil {
		return nil
	}

	if err := s.enrollmentRepo.Create(ctx, courseID, studentID); err != nil {
		return fmt.Errorf("linha %d: erro ao criar enrollment: %v", idx+1, err)
	}
	result.EnrollmentsAdded++
	return nil
}
