package services

import (
	"unicast-api/internal/models"
	"unicast-api/internal/repositories"
	"unicast-api/pkg/auth"
	"unicast-api/pkg/utils"

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

var customError = utils.CustomError{}
var makeError = customError.MakeError
var trace = utils.TraceError

var (
	ErrUserNotFound         = makeError("User not found", 404)
	ErrUserAlreadyExists    = makeError("User already exists", 409)
	ErrInvalidCredentials   = makeError("Invalid credentials", 401)
	ErrUnauthorized         = makeError("Unauthorized", 401)
	ErrInternalServer       = makeError("Internal server error", 500)
	ErrGenerateHash         = makeError("Error generating hash", 500)
	ErrGenerateSalt         = makeError("Error generating salt", 500)
	ErrGenerateAccessToken  = makeError("Error generating access token", 500)
	ErrGenerateRefreshToken = makeError("Error generating refresh token", 500)
	ErrGenerateJWE          = makeError("Error generating JWE", 500)
	ErrSaveRefreshToken     = makeError("Error saving refresh token", 500)
)

func NewAuthService(userRepo repositories.UserRepository, secrets *models.Secrets) AuthService {
	return &authService{userRepo: userRepo, secrets: secrets}
}

func (s *authService) Register(email, password, name string) (userId string, err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", trace("Register", ErrGenerateHash)
	}

	salt, err := auth.GenerateSalt(16)
	if err != nil {
		return "", trace("Register", ErrGenerateSalt)
	}

	user := &models.User{
		Email:    email,
		Password: string(hash),
		Name:     name,
		Salt:     salt,
	}

	userId, err = s.userRepo.Create(user)
	if err != nil {
		return "", trace("Register", err)
	}
	return userId, err
}

func (s *authService) Login(email, password string) (*LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, trace("Login", err)
	}
	if user == nil {
		return nil, trace("Login", ErrUserNotFound)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, trace("Login", ErrInvalidCredentials)
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, s.secrets.AccessToken)
	if err != nil {
		return nil, trace("Login", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, user.Email, s.secrets.RefreshToken)
	if err != nil {
		return nil, trace("Login", err)
	}

	smtpKey, err := auth.GenerateSmtpKey(password, user.Salt)
	if err != nil {
		return nil, trace("Login", err)
	}

	jwe, err := auth.GenerateJWE(JwePayload{SmtpKey: string(smtpKey)}, s.secrets.Jwe)
	if err != nil {
		return nil, trace("Login", err)
	}

	if err := s.userRepo.SaveRefreshToken(user.ID, refreshToken); err != nil {
		return nil, trace("Login", err)
	}

	return &LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		JWE:          jwe,
	}, nil
}

func (s *authService) RefreshToken(refreshToken string) (*RefreshResponse, error) {
	claims, err := auth.ValidateToken(refreshToken, s.secrets.RefreshToken)
	if err != nil {
		return nil, trace("RefreshToken", err)
	}
	user, err := s.userRepo.FindByEmail(claims.Email)
	if err != nil {
		return nil, trace("RefreshToken", err)
	}
	if user == nil || *user.RefreshToken != refreshToken {
		return nil, trace("RefreshToken", auth.ErrRefreshTokenNotValid)
	}
	newAccessToken, err := auth.GenerateAccessToken(user.ID, user.Email, s.secrets.AccessToken)
	if err != nil {
		return nil, trace("RefreshToken", err)
	}
	newRefreshToken, err := auth.GenerateRefreshToken(user.ID, user.Email, s.secrets.RefreshToken)
	if err != nil {
		return nil, trace("RefreshToken", err)
	}
	if err := s.userRepo.SaveRefreshToken(user.ID, newRefreshToken); err != nil {
		return nil, trace("RefreshToken", err)
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
