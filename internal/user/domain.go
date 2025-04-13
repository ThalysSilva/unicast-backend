package user

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name" validate:"required"`
	Email        string    `json:"email" validate:"required,email"`
	Password     string    `json:"-"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
	Salt         []byte    `json:"-"`
	RefreshToken *string   `json:"-"`
}

type Repository interface {
	database.Transactional
	Create(ctx context.Context, user *User) (userId string, err error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	SaveRefreshToken(ctx context.Context, userId string, refreshToken string) error
	Logout(ctx context.Context, userId string) error
	FindByID(ctx context.Context, id string) (*User, error)
}


func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
