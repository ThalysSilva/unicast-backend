package auth

import (
	"github.com/ThalysSilva/unicast-backend/internal/user"
	"github.com/dgrijalva/jwt-go"
)

type LoginResponse struct {
	User         *user.User `json:"user"`
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken"`
	JWE          string     `json:"jwe"`
}

type RefreshResponse struct {
	User         *user.User `json:"user"`
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken"`
	JWE          string     `json:"-"`
}
type JwePayload struct {
	SmtpKey string `json:"smtpKey"`
}

type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	jwt.StandardClaims
}
