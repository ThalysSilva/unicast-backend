package services

import (
	"errors"
	"fmt"
	"todo-list-api/internal/models"
	"todo-list-api/internal/repositories"
	"todo-list-api/pkg/auth"
	"todo-list-api/pkg/utils"

	"golang.org/x/crypto/bcrypt"
)

type LoginResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	JWE          string       `json:"jwe"`
}

type RefreshResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	JWE          string       `json:"-"`
}
type JwePayload struct {
	SmtpKey string `json:"smtpKey"`
}
type AuthService interface {
	Register(email, password, name string) (userId string, err error)
	Login(email, password string) (*LoginResponse, error)
	Logout(userId string) error
	RefreshToken(refreshToken string) (*RefreshResponse, error)
}

type authService struct {
	userRepo repositories.UserRepository
	secrets  *models.Secrets
}

func NewAuthService(userRepo repositories.UserRepository, secrets *models.Secrets) AuthService {
	return &authService{userRepo: userRepo, secrets: secrets}
}

func (s *authService) Register(email, password, name string) (userId string, err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("Register: %w", err)
	}

	salt, err := auth.GenerateSalt(16)
	if err != nil {
		return "", err
	}

	user := &models.User{
		Email:        email,
		Password: string(hash),
		Name:         name,
		Salt:         salt,
	}

	userId, err = s.userRepo.CreateUser(user)
	return userId, err
}

func (s *authService) Login(email, password string) (*LoginResponse, error) {
	trace := utils.TraceError("Login")
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, trace(err)
	}
	if user == nil {
		return nil, trace(errors.New("usuário não encontrado"))
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, trace(errors.New("senha inválida"))
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, s.secrets.AccessToken)
	if err != nil {
		return nil, trace(err)
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, user.Email, s.secrets.RefreshToken)
	if err != nil {
		return nil, trace(err)
	}

	smtpKey, err := auth.GenerateSmtpKey(password, user.Salt)
	if err != nil {
		return nil, trace(err)
	}

	jwe, err := auth.GenerateJWE(JwePayload{SmtpKey: string(smtpKey)}, s.secrets.Jwe)
	if err != nil {
		return nil, trace(err)
	}

	if err := s.userRepo.SaveRefreshToken(user.ID, refreshToken); err != nil {
		return nil, trace(err)
	}

	return &LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		JWE:          jwe,
	}, nil
}

func (s *authService) RefreshToken(refreshToken string) (*RefreshResponse, error) {
	trace := utils.TraceError("Login")
	claims, err := auth.ValidateToken(refreshToken, s.secrets.RefreshToken)
	if err != nil {
		return nil, trace(err)
	}
	user, err := s.userRepo.GetUserByEmail(claims.Email)
	if err != nil {
		return nil, trace(err)
	}
	if user == nil || *user.RefreshToken != refreshToken {
		return nil, trace(errors.New("refresh token inválido"))
	}
	newAccessToken, err := auth.GenerateAccessToken(user.ID, user.Email, s.secrets.AccessToken)
	if err != nil {
		return nil, trace(err)
	}
	newRefreshToken, err := auth.GenerateRefreshToken(user.ID, user.Email, s.secrets.RefreshToken)
	if err != nil {
		return nil, trace(err)
	}
	if err := s.userRepo.SaveRefreshToken(user.ID, newRefreshToken); err != nil {
		return nil, trace(err)
	}
	return &RefreshResponse{
		User:         user,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		JWE:          "",
	}, nil
}

func (s *authService) Logout(userId string) error {
	return s.userRepo.Logout(userId)
}
