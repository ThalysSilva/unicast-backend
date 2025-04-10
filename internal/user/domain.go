package user

import (
	"database/sql"
	"time"
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
	Create(user *User) (userId string, err error)
	FindByEmail(email string) (*User, error)
	SaveRefreshToken(userId string, refreshToken string) error
	Logout(userId string) error
	FindByID(id string) (*User, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return newNativeRepository(db)
}
