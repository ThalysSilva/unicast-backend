package student

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type StudentStatus string

// StudentStatus enum
const (
	StudentStatusActive    StudentStatus = "ACTIVE"
	StudentStatusCanceled  StudentStatus = "CANCELED"
	StudentStatusGraduated StudentStatus = "GRADUATED"
	StudentStatusLocked    StudentStatus = "LOCKED"
	StudentStatusPending   StudentStatus = "PENDING"
)

type Student struct {
	ID         string        `json:"id"`
	StudentID  string        `json:"studentId"`
	Name       *string       `json:"name"`
	Phone      *string       `json:"phone"`
	Email      *string       `json:"email" validate:"email"`
	Annotation *string       `json:"annotation"`
	Consent    bool          `json:"consent"`
	CreatedAt  time.Time     `json:"-"`
	UpdatedAt  time.Time     `json:"-"`
	Status     StudentStatus `json:"status"`
}

type Repository interface {
	database.Transactional
	Create(ctx context.Context, studentID string, name, phone, email, annotation *string, status StudentStatus) error
	FindByID(ctx context.Context, id string) (*Student, error)
	FindByStudentID(ctx context.Context, studentID string) (*Student, error)
	FindByFilters(ctx context.Context, filters map[string]string) ([]*Student, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, id string) error
	FindByIDs(ctx context.Context, ids []string) ([]*Student, error)
}

func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
