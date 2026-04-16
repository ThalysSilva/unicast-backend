package smtp

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/database"
)

type sqlRepository struct {
	db    database.DB
	sqlDB *sql.DB
}

// Cria uma nova instância do repositório
func newSQLRepository(db *sql.DB) Repository {
	newDb := database.NewSQLTx(db)
	return &sqlRepository{
		db: newDb.DB,
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

// Insere uma nova instância SMTP
func (r *sqlRepository) Create(ctx context.Context, userID, email, host string, port int, password, iv []byte) error {
	query := `
        INSERT INTO smtp_instances (host, port, email, password, iv, user_id, auth_mode, provider)
        VALUES ($1, $2, $3, $4, $5, $6, 'password', 'custom_smtp')
    `
	_, err := r.db.ExecContext(ctx, query, host, port, email, password, iv, userID)
	return err
}

func (r *sqlRepository) UpsertOAuth(ctx context.Context, userID, email, provider, host string, port int, oauthPayload, oauthIV []byte, tokenExpiresAt *time.Time) error {
	query := `
		INSERT INTO smtp_instances (
			host, port, email, password, iv, user_id, auth_mode, provider, oauth_payload, oauth_iv, token_expires_at
		)
		VALUES ($1, $2, $3, NULL, NULL, $4, 'oauth', $5, $6, $7, $8)
		ON CONFLICT (user_id, provider, email) WHERE auth_mode = 'oauth'
		DO UPDATE SET
			host = EXCLUDED.host,
			port = EXCLUDED.port,
			auth_mode = 'oauth',
			oauth_payload = EXCLUDED.oauth_payload,
			oauth_iv = EXCLUDED.oauth_iv,
			token_expires_at = EXCLUDED.token_expires_at,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query, host, port, email, userID, provider, oauthPayload, oauthIV, tokenExpiresAt)
	return err
}

// Busca uma instância SMTP pelo ID
func (r *sqlRepository) FindByID(ctx context.Context, id string) (*Instance, error) {
	query := `
        SELECT id, host, port, email, auth_mode, provider, password, iv, oauth_payload, oauth_iv, token_expires_at, created_at, updated_at, user_id
        FROM smtp_instances
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	instance := &Instance{}
	var tokenExpiresAt sql.NullTime
	err := row.Scan(
		&instance.ID,
		&instance.Host,
		&instance.Port,
		&instance.Email,
		&instance.AuthMode,
		&instance.Provider,
		&instance.Password,
		&instance.IV,
		&instance.OAuthPayload,
		&instance.OAuthIV,
		&tokenExpiresAt,
		&instance.CreatedAt,
		&instance.UpdatedAt,
		&instance.UserID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if tokenExpiresAt.Valid {
		instance.TokenExpiresAt = &tokenExpiresAt.Time
	}
	return instance, nil
}

func (r *sqlRepository) GetInstances(ctx context.Context, userID string) ([]*Instance, error) {

	query := `
				SELECT id, host, port, email, auth_mode, provider, password, iv, oauth_payload, oauth_iv, token_expires_at, created_at, updated_at, user_id
				FROM smtp_instances
				WHERE user_id = $1
		`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var instances []*Instance
	for rows.Next() {
		instance := &Instance{}
		var tokenExpiresAt sql.NullTime
		err := rows.Scan(
			&instance.ID,
			&instance.Host,
			&instance.Port,
			&instance.Email,
			&instance.AuthMode,
			&instance.Provider,
			&instance.Password,
			&instance.IV,
			&instance.OAuthPayload,
			&instance.OAuthIV,
			&tokenExpiresAt,
			&instance.CreatedAt,
			&instance.UpdatedAt,
			&instance.UserID,
		)
		if err != nil {
			return nil, err
		}
		if tokenExpiresAt.Valid {
			instance.TokenExpiresAt = &tokenExpiresAt.Time
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

func (r *sqlRepository) UpdateOAuthTokens(ctx context.Context, id string, oauthPayload, oauthIV []byte, tokenExpiresAt *time.Time) error {
	query := `
		UPDATE smtp_instances
		SET oauth_payload = $2,
			oauth_iv = $3,
			token_expires_at = $4,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, oauthPayload, oauthIV, tokenExpiresAt)
	return err
}

// Remove uma instância SMTP pelo ID
func (r *sqlRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM smtp_instances WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
