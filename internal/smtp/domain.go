package smtp

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Instance struct {
	ID            string     `json:"id"`
	Host          string     `json:"host"`
	Port          int        `json:"port"`
	Email         string     `json:"email" validate:"required"`
	AuthMode      string     `json:"authMode"`
	Provider      string     `json:"provider"`
	Password      []byte     `json:"-"`
	IV            []byte     `json:"-"`
	OAuthPayload  []byte     `json:"-"`
	OAuthIV       []byte     `json:"-"`
	TokenExpiresAt *time.Time `json:"tokenExpiresAt,omitempty"`
	CreatedAt     time.Time  `json:"-"`
	UpdatedAt     time.Time  `json:"-"`
	UserID        string     `json:"-"`
}

type Repository interface {
	database.Transactional
	Create(ctx context.Context, userID, email, host string, port int, password, iv []byte) error
	UpsertOAuth(ctx context.Context, userID, email, provider, host string, port int, oauthPayload, oauthIV []byte, tokenExpiresAt *time.Time) error
	FindByID(ctx context.Context, id string) (*Instance, error)
	UpdateOAuthTokens(ctx context.Context, id string, oauthPayload, oauthIV []byte, tokenExpiresAt *time.Time) error
	Delete(ctx context.Context, id string) error
	GetInstances(ctx context.Context, userID string) ([]*Instance, error)
}

func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
