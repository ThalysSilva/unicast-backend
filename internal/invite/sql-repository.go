package invite

import (
	"context"
	"database/sql"
	"strings"
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

func (r *sqlRepository) Create(ctx context.Context, disciplineID, code string, expiresAt *time.Time) error {
	query := `
        INSERT INTO invites (discipline_id, code, expires_at)
        VALUES ($1, $2, $3)
    `
	_, err := r.db.ExecContext(ctx, query, disciplineID, code, expiresAt)
	if err != nil {
		return customerror.Trace("inviteRepository: create", err)
	}
	return nil
}

func (r *sqlRepository) FindByID(ctx context.Context, id string) (*Invite, error) {
	query := `
        SELECT id, discipline_id, code, expires_at, active, created_at, updated_at
        FROM invites
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	return scanInvite(row, "inviteRepository: findByID")
}

func (r *sqlRepository) FindByCode(ctx context.Context, code string) (*Invite, error) {
	query := `
        SELECT id, discipline_id, code, expires_at, active, created_at, updated_at
        FROM invites
        WHERE UPPER(code) = UPPER($1)
    `
	row := r.db.QueryRowContext(ctx, query, strings.TrimSpace(code))

	return scanInvite(row, "inviteRepository: findByCode")
}

func (r *sqlRepository) FindLatestByDisciplineID(ctx context.Context, disciplineID string) (*Invite, error) {
	query := `
        SELECT id, discipline_id, code, expires_at, active, created_at, updated_at
        FROM invites
        WHERE discipline_id = $1
        ORDER BY created_at DESC
        LIMIT 1
    `
	row := r.db.QueryRowContext(ctx, query, disciplineID)

	return scanInvite(row, "inviteRepository: findLatestByDisciplineID")
}

func scanInvite(scanner interface {
	Scan(dest ...any) error
}, trace string) (*Invite, error) {
	invite := &Invite{}
	var expiresAt sql.NullTime
	err := scanner.Scan(&invite.ID, &invite.DisciplineID, &invite.Code, &expiresAt, &invite.Active, &invite.CreatedAt, &invite.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, customerror.Trace(trace, err)
	}
	if expiresAt.Valid {
		invite.ExpiresAt = &expiresAt.Time
	}
	return invite, nil
}

func (r *sqlRepository) FindByDisciplineID(ctx context.Context, disciplineID string) ([]*Invite, error) {
	query := `
        SELECT id, discipline_id, code, expires_at, active, created_at, updated_at
        FROM invites
        WHERE discipline_id = $1
        ORDER BY created_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query, disciplineID)
	if err != nil {
		return nil, customerror.Trace("inviteRepository: findByDisciplineID", err)
	}
	defer rows.Close()

	invites := make([]*Invite, 0)
	for rows.Next() {
		invite := &Invite{}
		var expiresAt sql.NullTime
		err := rows.Scan(&invite.ID, &invite.DisciplineID, &invite.Code, &expiresAt, &invite.Active, &invite.CreatedAt, &invite.UpdatedAt)
		if err != nil {
			return nil, customerror.Trace("inviteRepository: findByDisciplineID", err)
		}
		if expiresAt.Valid {
			invite.ExpiresAt = &expiresAt.Time
		}
		invites = append(invites, invite)
	}
	return invites, nil
}

func (r *sqlRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM invites WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return customerror.Trace("inviteRepository: delete", err)
	}
	return nil
}
