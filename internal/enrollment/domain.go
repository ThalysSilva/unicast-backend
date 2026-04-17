package enrollment

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Enrollment struct {
	ID                          string     `json:"id"`
	DisciplineID                string     `json:"disciplineId"`
	StudentID                   string     `json:"studentId"`
	SelfRegistrationCompletedAt *time.Time `json:"selfRegistrationCompletedAt"`
	SelfRegistrationCount       int        `json:"selfRegistrationCount"`
	CreatedAt                   time.Time  `json:"-"`
	UpdatedAt                   time.Time  `json:"-"`
}

type Repository interface {
	database.Transactional
	Create(ctx context.Context, disciplineID, studentID string) error
	FindByID(ctx context.Context, cid string) (*Enrollment, error)
	FindByDisciplineAndStudent(ctx context.Context, disciplineID, studentID string) (*Enrollment, error)
	DeleteByDisciplineID(ctx context.Context, disciplineID string) error
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, cid string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
