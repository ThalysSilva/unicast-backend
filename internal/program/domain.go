package program

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Program struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
	CampusID    string    `json:"-"`
	Active      bool      `json:"active"`
}

type ProgramWithUserOwnerID struct {
	Program
	UserOwnerID string `json:"owner_id"`
}

type Repository interface {
	database.Transactional
	Create(ctx context.Context, name, description, campusID string, active bool) error
	FindByID(ctx context.Context, id string) (*Program, error)
	// Campos disponíveis para atualização
	//
	// - name string
	//
	// - description string
	//
	// - active bool
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, id string) error
	FindByNameAndCampusID(ctx context.Context, name, campusID string) (*Program, error)
	FindByIDWithUserOwnerID(ctx context.Context, id string) (*ProgramWithUserOwnerID, error)
	FindByCampusID(ctx context.Context, campusID string) ([]*Program, error)
}

func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
