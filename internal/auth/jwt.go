package auth

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/customerror"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jwk"
)


var (
	ErrTokenNotValid        = customerror.Make("Token inválido", 401,  errors.New("ErrTokenNotValid"))
	ErrRefreshTokenNotValid = customerror.Make("Refresh token inválido", 401, errors.New("ErrRefreshTokenNotValid"))
	ErrInvalidJweSecret     = customerror.Make("JWE secret inválido", 401, errors.New("ErrInvalidJweSecret"))
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
		return "", customerror.Trace("GenerateAccessToken", err)
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
		return "", customerror.Trace("GenerateRefreshToken", err)
	}
	return signedToken, nil
}

func GenerateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)

	if _, err := rand.Read(salt); err != nil {
		return nil, customerror.Trace("GenerateSalt", err)
	}

	return salt, nil
}
func ValidateToken(tokenStr string, secret []byte) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		return secret, nil
	})
	if err != nil {
		return nil, customerror.Trace("ValidateToken", err)
	}
	if !token.Valid {
		return nil, customerror.Trace("ValidateToken", ErrTokenNotValid)
	}
	return claims, nil
}

// GenerateJWE criptografa o payload como JWE usando AES-256-GCM.
// O payload é serializado para JSON antes da criptografia.
func GenerateJWE(payload any, secret []byte) (string, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", customerror.Trace("GenerateJWE", err)
	}

	key, err := jwk.FromRaw(secret)
	if err != nil {
		return "", customerror.Trace("GenerateJWE", err)
	}

	jwe, err := jwe.Encrypt(payloadBytes, jwe.WithKey(jwa.A256KW, key), jwe.WithContentEncryption(jwa.A256GCM))
	if err != nil {
		return "", customerror.Trace("GenerateJWE", err)
	}

	return string(jwe), nil
}



func DecryptJWE[T any](jweToken string, secret []byte) (T, error) {

	if len(secret) != 32 {
		return *new(T), customerror.Trace("DecryptJWE", ErrInvalidJweSecret)
	}
	key, err := jwk.FromRaw(secret)
	if err != nil {
		return *new(T), customerror.Trace("DecryptJWE", err)
	}

	decryptedBytes, err := jwe.Decrypt([]byte(jweToken), jwe.WithKey(jwa.A256KW, key))
	if err != nil {
		return *new(T), customerror.Trace("DecryptJWE", err)
	}

	var decryptedPayload T
	if err := json.Unmarshal(decryptedBytes, &decryptedPayload); err != nil {
		return *new(T), customerror.Trace("DecryptJWE", err)
	}

	return decryptedPayload, nil
}
