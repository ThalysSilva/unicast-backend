package smtp

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Instance struct {
	ID        string    `json:"id"`
	Host      string    `json:"host" validate:"required"`
	Port      int       `json:"port" validate:"required"`
	Email     string    `json:"email" validate:"required"`
	Password  string    `json:"password" validate:"required"`
	IV        string    `json:"iv" validate:"required"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	UserID    string    `json:"-"`
}

type Repository interface {
	database.Transactional
	Create(ctx context.Context, userID, email, host string, port int, password, iv []byte) error
	FindByID(ctx context.Context, id string) (*Instance, error)
	Update(ctx context.Context, id int, fields map[string]any) error
	Delete(ctx context.Context, id string) error
	GetInstances(ctx context.Context, userID string) ([]*Instance, error)
}

func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
