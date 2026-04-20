package discipline

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Discipline struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	Year        int       `json:"year" validate:"required"`
	Semester    int       `json:"semester" validate:"required"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
	ProgramID   string    `json:"-"`
}

type DisciplineWithOwnerID struct {
	Discipline
	UserOwnerID string `json:"-"`
}

type Repository interface {
	database.Transactional
	Create(ctx context.Context, name, description, programID string, year, semester int) error
	FindByID(ctx context.Context, id string) (*Discipline, error)
	FindByIDWithUserOwnerID(ctx context.Context, id string) (*DisciplineWithOwnerID, error)
	FindByProgramID(ctx context.Context, programID string) ([]*Discipline, error)
	FindByUserOwnerID(ctx context.Context, userOwnerID string) ([]*Discipline, error)
	// Campos disponíveis para atualização
	//
	// - name string
	//
	// - description string
	//
	// - year int
	//
	// - semester int
	//
	// - program_id string
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, id string) error
	FindByNameAndProgramID(ctx context.Context, name, programID string) (*Discipline, error)
}

func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
