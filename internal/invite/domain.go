package invite

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type Invite struct {
	ID        string     `json:"id"`
	CourseID  string     `json:"courseId"`
	Code      string     `json:"code"`
	ExpiresAt *time.Time `json:"expiresAt"`
	Active    bool       `json:"active"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
}

type Repository interface {
	database.Transactional
	Create(ctx context.Context, courseID, code string, expiresAt *time.Time) error
	FindByCode(ctx context.Context, code string) (*Invite, error)
}

func NewRepository(db *sql.DB) Repository {
	return newSQLRepository(db)
}
