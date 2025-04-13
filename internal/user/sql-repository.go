package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/database"
	"github.com/lib/pq"
)

type sqlRepository struct {
	db    database.DB
	sqlDB *sql.DB // Para TransactionBackend
}

var (
	ErrUserAlreadyExists = customerror.Make("Usuário já existe", 409, errors.New("ErrUserAlreadyExists"))
)

func newSQLRepository(db *sql.DB) Repository {
	return &sqlRepository{
		db:    database.NewSQLTx(db).DB,
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

// Cria um novo usuário no banco de dados e retorna o ID do usuário criado. Se o usuário já existir, retorna um erro.
func (r *sqlRepository) Create(ctx context.Context, user *User) (userId string, err error) {
	query := "INSERT INTO users (email, name, password, salt) VALUES ($1, $2, $3 ,$4) RETURNING id"
	err = r.db.QueryRowContext(ctx, query, user.Email, user.Name, user.Password, string(user.Salt)).Scan(&userId)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return "", customerror.Trace("CreateUser", ErrUserAlreadyExists)
			}
		}
		return "", customerror.Trace("CreateUser", err)
	}
	return userId, nil
}

// Atualiza um usuário existente
func (r *sqlRepository) Update(ctx context.Context, user *User) error {
	query := `
			UPDATE users
			SET email = $2, name = $3, password = $5, refresh_token = $6, salt = $7
			WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.Name, user.Password, user.RefreshToken, user.Salt)
	return err
}

// Busca um usuário pelo ID
func (r *sqlRepository) FindByID(ctx context.Context, id string) (*User, error) {
	query := `
			SELECT id, email, name, created_at, updated_at, password, refresh_token, salt
			FROM users
			WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	user := &User{}
	var refreshToken sql.NullString
	err := row.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt, &user.Password, &refreshToken, &user.Salt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		err = customerror.Trace("GetUserByID: ", err)
		return nil, err
	}

	if refreshToken.Valid {
		user.RefreshToken = &refreshToken.String
	}
	return user, nil
}

// Encontra um usuário no banco de dados pelo email. Se o usuário não existir, retorna nil. Se ocorrer um erro, retorna o erro.
func (r *sqlRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	query := "SELECT id, email, password, name, refresh_token, salt FROM users WHERE email = $1"
	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.RefreshToken, &user.Salt) // TO-CHECK: confirmar se o refresh está ok

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, customerror.Trace("GetUserByEmail", err)
	}
	return user, nil
}

// Salva o token de atualização do usuário no banco de dados.
func (r *sqlRepository) SaveRefreshToken(ctx context.Context, userId string, refreshToken string) error {
	query := "UPDATE users SET refresh_token = $1 WHERE id = $2"

	if _, err := r.db.ExecContext(ctx, query, refreshToken, userId); err != nil {
		return customerror.Trace("SaveRefreshToken", err)

	}
	return nil
}

// Remove o token de atualização do usuário no banco de dados.
func (r *sqlRepository) Logout(ctx context.Context, userId string) error {
	query := "UPDATE users SET refresh_token = NULL WHERE id = $1"

	if _, err := r.db.ExecContext(ctx, query, userId); err != nil {
		return customerror.Trace("Logout", err)
	}
	return nil
}
