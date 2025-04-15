package auth

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/ThalysSilva/unicast-backend/internal/config"
	"github.com/ThalysSilva/unicast-backend/internal/encryption"
	"github.com/ThalysSilva/unicast-backend/internal/user"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, email, password, name string) (userID string, err error)
	Login(ctx context.Context, email, password string) (*LoginResponse, error)
	Logout(ctx context.Context, userID string) error
	RefreshToken(ctx context.Context, refreshToken string) (*RefreshResponse, error)
}

type service struct {
	userRepo user.Repository
	secrets  *config.Secrets
}

var (
	ErrUserNotFound         = customerror.Make("User not found", 404, errors.New("ErrUserNotFound"))
	ErrUserAlreadyExists    = customerror.Make("User already exists", 409, errors.New("ErrUserAlreadyExists"))
	ErrInvalidCredentials   = customerror.Make("Invalid credentials", 401, errors.New("ErrInvalidCredentials"))
	ErrUnauthorized         = customerror.Make("Unauthorized", 401, errors.New("ErrUnauthorized"))
	ErrInternalServer       = customerror.Make("Internal server error", 500, errors.New("ErrInternalServer"))
	ErrGenerateHash         = customerror.Make("Error generating hash", 500, errors.New("ErrGenerateHash"))
	ErrGenerateSalt         = customerror.Make("Error generating salt", 500, errors.New("ErrGenerateSalt"))
	ErrGenerateAccessToken  = customerror.Make("Error generating access token", 500, errors.New("ErrGenerateAccessToken"))
	ErrGenerateRefreshToken = customerror.Make("Error generating refresh token", 500, errors.New("ErrGenerateRefreshToken"))
	ErrGenerateJWE          = customerror.Make("Error generating JWE", 500, errors.New("ErrGenerateJWE"))
	ErrSaveRefreshToken     = customerror.Make("Error saving refresh token", 500, errors.New("ErrSaveRefreshToken"))
)

func NewService(userRepo user.Repository, secrets *config.Secrets) Service {
	return &service{userRepo: userRepo, secrets: secrets}
}

func (s *service) Register(ctx context.Context, email, password, name string) (userID string, err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", customerror.Trace("Register", ErrGenerateHash)
	}

	salt, err := GenerateSalt(16)
	if err != nil {
		return "", customerror.Trace("Register", ErrGenerateSalt)
	}

	user := &user.User{
		Email:    email,
		Password: string(hash),
		Name:     name,
		Salt:     salt,
	}

	userID, err = s.userRepo.Create(ctx, user)
	if err != nil {
		return "", customerror.Trace("Register", err)
	}
	return userID, err
}

func (s *service) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, customerror.Trace("Login", err)
	}
	if user == nil {
		return nil, customerror.Trace("Login", ErrUserNotFound)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, customerror.Trace("Login", ErrInvalidCredentials)
	}

	accessToken, err := GenerateAccessToken(user.ID, user.Email, s.secrets.AccessToken)
	if err != nil {
		return nil, customerror.Trace("Login", err)
	}

	refreshToken, err := GenerateRefreshToken(user.ID, user.Email, s.secrets.RefreshToken)
	if err != nil {
		return nil, customerror.Trace("Login", err)
	}

	smtpKey, err := encryption.GenerateSmtpKey(password, user.Salt)
	if err != nil {
		return nil, customerror.Trace("Login", err)
	}

	smtpKeyEncoded := base64.StdEncoding.EncodeToString(smtpKey)

	jwe, err := GenerateJWE(JwePayload{SmtpKeyEncoded: smtpKeyEncoded}, s.secrets.Jwe)
	if err != nil {
		return nil, customerror.Trace("Login", err)
	}

	if err := s.userRepo.SaveRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, customerror.Trace("Login", err)
	}

	return &LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		JWE:          jwe,
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*RefreshResponse, error) {
	claims, err := ValidateToken(refreshToken, s.secrets.RefreshToken)
	if err != nil {
		return nil, customerror.Trace("RefreshToken", err)
	}
	user, err := s.userRepo.FindByEmail(ctx, claims.Email)
	if err != nil {
		return nil, customerror.Trace("RefreshToken", err)
	}
	if user == nil || *user.RefreshToken != refreshToken {
		return nil, customerror.Trace("RefreshToken", ErrRefreshTokenNotValid)
	}
	newAccessToken, err := GenerateAccessToken(user.ID, user.Email, s.secrets.AccessToken)
	if err != nil {
		return nil, customerror.Trace("RefreshToken", err)
	}
	newRefreshToken, err := GenerateRefreshToken(user.ID, user.Email, s.secrets.RefreshToken)
	if err != nil {
		return nil, customerror.Trace("RefreshToken", err)
	}
	if err := s.userRepo.SaveRefreshToken(ctx, user.ID, newRefreshToken); err != nil {
		return nil, customerror.Trace("RefreshToken", err)
	}
	return &RefreshResponse{
		User:         user,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		JWE:          "",
	}, nil
}

func (s *service) Logout(ctx context.Context, userID string) error {
	return s.userRepo.Logout(ctx, userID)
}
