package invite

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

// Gerencia operações de banco para Invite
type sqlRepository struct {
	db    database.DB
	sqlDB *sql.DB
}

func newSQLRepository(db *sql.DB) Repository {
	newDb := database.NewSQLTx(db)
	return &sqlRepository{
		db:    newDb.DB,
		sqlDB: db,
	}
}

func (r *sqlRepository) WithTransaction(tx any) any {
	return &sqlRepository{
		db:    database.NewSQLTx(nil).WithSQLTransaction(tx).DB,
		sqlDB: r.sqlDB,
	}
}

func (r *sqlRepository) TransactionBackend() any {
	return r.sqlDB
}

func (r *sqlRepository) Create(ctx context.Context, courseID, code string, expiresAt *time.Time) error {
	query := `
        INSERT INTO invites (course_id, code, expires_at)
        VALUES ($1, $2, $3)
    `
	_, err := r.db.ExecContext(ctx, query, courseID, code, expiresAt)
	if err != nil {
		return customerror.Trace("inviteRepository: create", err)
	}
	return nil
}

func (r *sqlRepository) FindByCode(ctx context.Context, code string) (*Invite, error) {
	query := `
        SELECT id, course_id, code, expires_at, active, created_at, updated_at
        FROM invites
        WHERE code = $1
    `
	row := r.db.QueryRowContext(ctx, query, code)

	invite := &Invite{}
	var expiresAt sql.NullTime
	err := row.Scan(&invite.ID, &invite.CourseID, &invite.Code, &expiresAt, &invite.Active, &invite.CreatedAt, &invite.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, customerror.Trace("inviteRepository: findByCode", err)
	}
	if expiresAt.Valid {
		invite.ExpiresAt = &expiresAt.Time
	}
	return invite, nil
}
