package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"unicast-api/pkg/utils"
)

var customError = &utils.CustomError{}
var makeError = customError.MakeError
var trace = utils.TraceError

var (
	ErrSaltSize = makeError("Salt deve ter pelo menos 8 bytes.", 500)
)

func GenerateSmtpKey(password string, salt []byte) ([]byte, error) {
	if len(salt) < 8 {
		return nil, trace("GenerateSmtpKey", ErrSaltSize)
	}
	encryptedKey, err := pbkdf2.Key(sha256.New, password, salt, 10000, 32)
	if err != nil {
		return nil, trace("GenerateSmtpKey", err)
	}
	return encryptedKey, nil
}

func EncryptSmtpPassword(smtpPassword string, smtpKey []byte) (encryptedSmtpPassword, iv []byte, err error) {

	// Criar cifra AES com o SmtpKey
	block, err := aes.NewCipher(smtpKey)
	if err != nil {
		return nil, nil, trace("EncryptSmtpPassword", err)
	}

	// Usar AES-GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, trace("EncryptSmtpPassword", err)
	}

	iv = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, trace("EncryptSmtpPassword", err)
	}

	ciphertext := gcm.Seal(nil, iv, []byte(smtpPassword), nil)
	return ciphertext, iv, nil
}

func DecryptSmtpPassword(encryptedPassword, smtpKey, iv []byte) (string, error) {

	block, err := aes.NewCipher(smtpKey)
	if err != nil {
			return "", trace("DecryptSmtpPassword", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
			return "", trace("DecryptSmtpPassword", err)
	}

	plaintext, err := gcm.Open(nil, iv, encryptedPassword, nil)
	if err != nil {
			return "", trace("DecryptSmtpPassword", err)
	}

	return string(plaintext), nil
}
