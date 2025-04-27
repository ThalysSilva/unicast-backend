package campus

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Campus struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
	UserOwnerID string    `json:"-"`
}

type Repository interface {
	database.Transactional
	Create(ctx context.Context, name, description, userOwnerID string) error
	FindByID(ctx context.Context, id string) (*Campus, error)
	FindByUserOwnerId(ctx context.Context, userOwnerID string) ([]*Campus, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	Delete(ctx context.Context, id string) error
}

func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
