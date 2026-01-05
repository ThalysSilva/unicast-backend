package backdoor

import (
	"context"
	"errors"

	"github.com/ThalysSilva/unicast-backend/internal/auth"
	"github.com/ThalysSilva/unicast-backend/internal/user"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	ResetPassword(ctx context.Context, secret, userID, email, newPassword string) error
}

type service struct {
	userRepo    user.Repository
	adminSecret string
}

var (
	ErrInvalidSecret = customerror.Make("segredo inválido", 403, errors.New("ErrInvalidSecret"))
	ErrUserNotFound  = customerror.Make("usuário não encontrado", 404, errors.New("ErrUserNotFound"))
)

func NewService(userRepo user.Repository, adminSecret string) Service {
	return &service{
		userRepo:    userRepo,
		adminSecret: adminSecret,
	}
}

func (s *service) ResetPassword(ctx context.Context, secret, userID, email, newPassword string) error {
	if secret != s.adminSecret {
		return ErrInvalidSecret
	}

	var u *user.User
	var err error
	if userID != "" {
		u, err = s.userRepo.FindByID(ctx, userID)
	} else if email != "" {
		u, err = s.userRepo.FindByEmail(ctx, email)
	} else {
		return customerror.Make("informe userId ou email", 400, errors.New("ErrMissingUserIdentifier"))
	}
	if err != nil {
		return err
	}
	if u == nil {
		return ErrUserNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return customerror.Trace("ResetPassword", err)
	}
	salt, err := auth.GenerateSalt(16)
	if err != nil {
		return customerror.Trace("ResetPassword", err)
	}

	u.Password = string(hash)
	u.Salt = salt
	u.RefreshToken = nil // invalida sessões

	return s.userRepo.Update(ctx, u)
}
