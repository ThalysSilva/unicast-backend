package repositories

import (
	"database/sql"
	"unicast-api/internal/models"
	"unicast-api/pkg/utils"
)

type UserRepository interface {
	CreateUser(user *models.User) (userId string, err error)
	GetUserByEmail(username string) (*models.User, error)
	SaveRefreshToken(userId string, refreshToken string) error
	Logout(userId string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *models.User) (userId string, err error) {
	trace := utils.TraceError("CreateUser")
	query := "INSERT INTO users (email, name, password, salt) VALUES ($1, $2) RETURNING id"
	err = r.db.QueryRow(query, user.Email, user.Name, user.Password, user.Salt).Scan(&userId)
	if err != nil {
		return "", trace(err)
	}
	return userId, nil
}

func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	trace := utils.TraceError("GetUserByEmail")
	user := &models.User{}
	query := "SELECT id, email, password, refresh_token FROM users WHERE email = $1"
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password, user.RefreshToken) // TO-CHECK: confirmar se o refresh está ok
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, trace(err)
	}
	return user, nil
}

func (r *userRepository) SaveRefreshToken(userId string, refreshToken string) error {
	trace := utils.TraceError("SaveRefreshToken")
	query := "UPDATE users SET refresh_token = $1 WHERE id = $2"
	_, err := r.db.Exec(query, refreshToken, userId)
	return trace(err)
}

func (r *userRepository) Logout(userId string) error {
	trace := utils.TraceError("Logout")
	query := "UPDATE users SET refresh_token = NULL WHERE id = $1"
	_, err := r.db.Exec(query, userId)
	return trace(err)
}
