package program

import (
	"context"
	"database/sql"
	"time"
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

type Repository interface {
	Create(ctx context.Context, name, description, campusID string, active bool) error
	FindByID(ctx context.Context, id string) (*Program, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
