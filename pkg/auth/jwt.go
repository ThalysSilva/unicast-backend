package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
	"unicast-api/pkg/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	jwt.StandardClaims
}

var customError = &utils.CustomError{}
var makeError = customError.MakeError
var trace = utils.TraceError

var (
	ErrTokenNotValid = makeError("Token inválido", 401)
	ErrRefreshTokenNotValid = makeError("Refresh token inválido", 401)
)

func GenerateAccessToken(userID string, userEmail string, secret []byte) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		UserID: userID,
		Email:  userEmail,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", trace("GenerateAccessToken", err)
	}

	return signedToken, nil
}

func GenerateRefreshToken(userID string, userEmail string, secret []byte) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		Email:  userEmail,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", trace("GenerateRefreshToken", err)
	}
	return signedToken, nil
}

func GenerateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)

	if _, err := rand.Read(salt); err != nil {
		return nil, trace("GenerateSalt", err)
	}

	return salt, nil
}
func ValidateToken(tokenStr string, secret []byte) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		return secret, nil
	})
	if err != nil {
		return nil, trace("ValidateToken", err)
	}
	if !token.Valid {
		return nil, trace("ValidateToken", ErrTokenNotValid)
	}
	return claims, nil
}

// GenerateJWE criptografa o payload como JWE usando AES-256-GCM.
// O payload é serializado para JSON antes da criptografia.
func GenerateJWE(payload any, secret []byte) (string, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", trace("GenerateJWE", err)
	}

	key, err := jwk.FromRaw(secret)
	if err != nil {
		return "", trace("GenerateJWE", err)
	}

	jwe, err := jwe.Encrypt(payloadBytes, jwe.WithKey(jwa.A256KW, key), jwe.WithContentEncryption(jwa.A256GCM))
	if err != nil {
		return "", trace("GenerateJWE", err)
	}

	return string(jwe), nil
}

func GenerateSmtpKey(password string, salt []byte) ([]byte, error) {
	if len(salt) < 8 {
		err := fmt.Errorf("salt deve ter pelo menos 8 bytes. O atual tem %d bytes", len(salt))
		return nil, trace("GenerateSmtpKey", &utils.CustomError{Err: err, HttpCode: 500})
	}
	return pbkdf2.Key(sha256.New, password, salt, 10000, 32)
}
