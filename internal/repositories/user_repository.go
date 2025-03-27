package repositories

import (
	"database/sql"
	"github.com/lib/pq"
	"unicast-api/internal/models"
	"unicast-api/pkg/utils"
)

type UserRepository interface {
	CreateUser(user *models.User) (userId string, err error)
	GetUserByEmail(email string) (*models.User, error)
	SaveRefreshToken(userId string, refreshToken string) error
	Logout(userId string) error
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

func (r *userRepository) CreateUser(user *models.User) (userId string, err error) {
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

func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	var salt string
	query := "SELECT id, email, password, refresh_token, salt FROM users WHERE email = $1"
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password, user.RefreshToken, &salt ) // TO-CHECK: confirmar se o refresh está ok
	user.Salt = []byte(salt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, trace("CreateUser", err)
	}
	return user, nil
}

func (r *userRepository) SaveRefreshToken(userId string, refreshToken string) error {
	query := "UPDATE users SET refresh_token = $1 WHERE id = $2"

	if _, err := r.db.Exec(query, refreshToken, userId); err != nil {
		return trace("SaveRefreshToken", err)

	}
	return nil
}

func (r *userRepository) Logout(userId string) error {
	query := "UPDATE users SET refresh_token = NULL WHERE id = $1"

	if _, err := r.db.Exec(query, userId); err != nil {
		return trace("Logout", err)
	}
	return nil
}
