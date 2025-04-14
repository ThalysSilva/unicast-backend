package course

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Course struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	Year        int       `json:"year" validate:"required"`
	Semester    int       `json:"semester" validate:"required"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
	ProgramID   string    `json:"-"`
}

type Repository interface {
	database.Transactional
	Create(ctx context.Context, name, description, programID string, year, semester int) error
	FindByID(ctx context.Context, id string) (*Course, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, id string) error
}

func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
