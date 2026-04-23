package student

import (
	"context"
	"database/sql"
	"strings"
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
	ID          string        `json:"id"`
	StudentID   string        `json:"studentId"`
	Name        *string       `json:"name"`
	Phone       *string       `json:"phone"`
	NoPhone     bool          `json:"noPhone"`
	Email       *string       `json:"email" validate:"email"`
	Annotation  *string       `json:"annotation"`
	Consent     bool          `json:"consent"`
	EmailDeliveryIssue    bool      `json:"emailDeliveryIssue"`
	WhatsAppDeliveryIssue bool      `json:"whatsappDeliveryIssue"`
	CreatedAt   time.Time     `json:"-"`
	UpdatedAt   time.Time     `json:"-"`
	Status      StudentStatus `json:"status"`
	UserOwnerID string       `json:"-"`
}

func hasText(value *string) bool {
	return value != nil && strings.TrimSpace(*value) != ""
}

func HasCompletedContact(student *Student) bool {
	if student == nil {
		return false
	}

	return HasCompletedContactFields(student.Name, student.Phone, student.Email, student.NoPhone)
}

func HasCompletedContactFields(name, phone, email *string, noPhone bool) bool {
	return hasText(name) && hasText(email) && (noPhone || hasText(phone))
}

func DeriveContactAwareStatus(current, requested StudentStatus, statusProvided bool, name, phone, email *string, noPhone bool) StudentStatus {
	hasCompleteContact := HasCompletedContactFields(name, phone, email, noPhone)

	if statusProvided {
		if requested == StudentStatusActive && !hasCompleteContact {
			return StudentStatusPending
		}
		return requested
	}

	switch current {
	case StudentStatusLocked, StudentStatusGraduated, StudentStatusCanceled:
		return current
	}

	if hasCompleteContact {
		return StudentStatusActive
	}
	return StudentStatusPending
}

type Repository interface {
	database.Transactional
	Create(ctx context.Context, userOwnerID, studentID string, name, phone, email, annotation *string, noPhone bool, status StudentStatus) error
	FindByID(ctx context.Context, id, userOwnerID string) (*Student, error)
	FindByStudentID(ctx context.Context, studentID, userOwnerID string) (*Student, error)
	FindByFilters(ctx context.Context, filters map[string]string) ([]*Student, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, id string) error
	FindByIDs(ctx context.Context, userOwnerID string, ids []string) ([]*Student, error)
}

func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
