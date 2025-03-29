package repositories

import (
	"database/sql"
	"unicast-api/internal/models"
	"unicast-api/pkg/utils"

	"github.com/lib/pq"
)

type UserRepository interface {
	Create(user *models.User) (userId string, err error)
	FindByEmail(email string) (*models.User, error)
	SaveRefreshToken(userId string, refreshToken string) error
	Logout(userId string) error
	FindByID(id string) (*models.User, error)
}

type userRepository struct {
	db *sql.DB
}

var customError = &utils.CustomError{}
var makeError = customError.MakeError
var trace = utils.TraceError

var (
	ErrUserAlreadyExists = makeError("Usuário já existe", 409)
)

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// Cria um novo usuário no banco de dados e retorna o ID do usuário criado. Se o usuário já existir, retorna um erro.
func (r *userRepository) Create(user *models.User) (userId string, err error) {
	query := "INSERT INTO users (email, name, password, salt) VALUES ($1, $2, $3 ,$4) RETURNING id"
	err = r.db.QueryRow(query, user.Email, user.Name, user.Password, string(user.Salt)).Scan(&userId)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return "", trace("CreateUser", ErrUserAlreadyExists)
			}
		}
		return "", trace("CreateUser", err)
	}
	return userId, nil
}

// Atualiza um usuário existente
func (r *userRepository) Update(user *models.User) error {
	query := `
			UPDATE users
			SET email = $2, name = $3, password = $5, refresh_token = $6, salt = $7
			WHERE id = $1
	`
	_, err := r.db.Exec(query, user.ID, user.Email, user.Name, user.Password, user.RefreshToken, user.Salt)
	return err
}

// Busca um usuário pelo ID
func (r *userRepository) FindByID(id string) (*models.User, error) {
	query := `
			SELECT id, email, name, created_at, updated_at, password, refresh_token, salt
			FROM users
			WHERE id = $1
	`
	row := r.db.QueryRow(query, id)

	user := &models.User{}
	var refreshToken sql.NullString
	err := row.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt,  &user.Password, &refreshToken, &user.Salt)
	if err != nil {
			if err == sql.ErrNoRows {
					return nil, nil
			}
			return nil, err
	}

	if refreshToken.Valid {
			user.RefreshToken = &refreshToken.String
	}
	return user, nil
}

// Encontra um usuário no banco de dados pelo email. Se o usuário não existir, retorna nil. Se ocorrer um erro, retorna o erro.
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := "SELECT id, email, password, name, refresh_token, salt FROM users WHERE email = $1"
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.RefreshToken, &user.Salt) // TO-CHECK: confirmar se o refresh está ok

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, trace("GetUserByEmail", err)
	}
	return user, nil
}

// Salva o token de atualização do usuário no banco de dados.
func (r *userRepository) SaveRefreshToken(userId string, refreshToken string) error {
	query := "UPDATE users SET refresh_token = $1 WHERE id = $2"

	if _, err := r.db.Exec(query, refreshToken, userId); err != nil {
		return trace("SaveRefreshToken", err)

	}
	return nil
}

// Remove o token de atualização do usuário no banco de dados.
func (r *userRepository) Logout(userId string) error {
	query := "UPDATE users SET refresh_token = NULL WHERE id = $1"

	if _, err := r.db.Exec(query, userId); err != nil {
		return trace("Logout", err)
	}
	return nil
}
