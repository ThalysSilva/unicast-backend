package mailer

import (
	"errors"
	"net/smtp"
	"sync"
	"testing"
	"time"

	"github.com/jordan-wright/email"
	"github.com/stretchr/testify/assert"
)

type mockEmailPool struct {
	sendErr error
}

func (m *mockEmailPool) Send(e *email.Email, timeout time.Duration) error {
	return m.sendErr
}

func (m *mockEmailPool) Close() {}

func mockNewPoolFunc(host string, pools int, auth smtp.Auth) (emailPool, error) {
	return &mockEmailPool{}, nil
}

func mockNewPoolFuncWithError(host string, pools int, auth smtp.Auth) (emailPool, error) {
	return nil, errors.New("failed to create pool")
}

var originalNewPoolFunc = newPoolFunc

func TestSendEmails_Success(t *testing.T) {
	newPoolFunc = mockNewPoolFunc
	t.Cleanup(func() {
		newPoolFunc = originalNewPoolFunc
	})

	mailer := &MailerData{
		From:        "sender@example.com",
		To:          []string{"recipient@example.com"},
		Subject:     "Test Subject",
		Body:        "Test Body",
		ContentType: TextPlain,
		SmtpAuthentication: SmtpAuthentication{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user",
			Password: "pass",
		},
	}

	err := mailer.SendEmails(2, 2, 1, 5*time.Second)
	assert.Nil(t, err, "Expected no error on email send")
}

func TestSendEmails_InvalidContentType(t *testing.T) {
	mailer := &MailerData{
		From:        "sender@example.com",
		To:          []string{"recipient@example.com"},
		Subject:     "Test Subject",
		Body:        "Test Body",
		ContentType: "invalid", // ContentType inválido
	}

	err := mailer.SendEmails(2, 2, 1, 5*time.Second)
	assert.NotNil(t, err)
	assert.Equal(t, "conteúdo do email inválido", err.Error())
}

func TestSendEmails_FailureOnPoolCreation(t *testing.T) {
	newPoolFunc = mockNewPoolFuncWithError

	mailer := &MailerData{
		From:        "sender@example.com",
		To:          []string{"recipient@example.com"},
		Subject:     "Test Subject",
		Body:        "Test Body",
		ContentType: TextPlain,
		SmtpAuthentication: SmtpAuthentication{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user",
			Password: "pass",
		},
	}

	err := mailer.SendEmails(2, 2, 1, 5*time.Second)
	assert.NotNil(t, err)
	assert.Equal(t, "failed to create pool", err.Error())
}

func TestSendEmails_FailureOnSend(t *testing.T) {
	newPoolFunc = func(host string, pools int, auth smtp.Auth) (emailPool, error) {
		return &mockEmailPool{sendErr: errors.New("failed to send email")}, nil
	}

	mailer := &MailerData{
		From:        "sender@example.com",
		To:          []string{"recipient@example.com"},
		Subject:     "Test Subject",
		Body:        "Test Body",
		ContentType: TextPlain,
		SmtpAuthentication: SmtpAuthentication{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user",
			Password: "pass",
		},
	}

	err := mailer.SendEmails(2, 2, 1, 5*time.Second)
	assert.NotNil(t, err)
	assert.IsType(t, &EmailSentError{}, err)
}

func TestSendEmails_CaptureEmails(t *testing.T) {
	var wgTest sync.WaitGroup
	var wgInitTest sync.WaitGroup
	var muTest sync.Mutex
	newPoolFunc = mockNewPoolFunc
	capturedEmails := make([]*email.Email, 0)
	capturedRetries := make([]*EmailToRetry, 0)

	newEmailChannelsFunc = func(bufferSize int) (chan *email.Email, chan *EmailToRetry) {
		emailsChan := make(chan *email.Email, bufferSize)
		retriesChan := make(chan *EmailToRetry, bufferSize)

		// Captura os e-mails enviados
		wgTest.Add(1)
		wgInitTest.Add(2)
		go func() {
			defer wgTest.Done()
			wgInitTest.Done()
			for e := range emailsChan {
				muTest.Lock()
				capturedEmails = append(capturedEmails, e)
				muTest.Unlock()
			}
		}()
		// Captura os retries
		wgTest.Add(1)
		go func() {
			defer wgTest.Done()
			wgInitTest.Done()
			for r := range retriesChan {
				muTest.Lock()
				capturedRetries = append(capturedRetries, r)
				muTest.Unlock()
			}
		}()
		wgInitTest.Wait()

		return emailsChan, retriesChan
	}

	mailer := &MailerData{
		From:        "sender@example.com",
		To:          []string{"recipient1@example.com", "recipient2@example.com"},
		Subject:     "Test Email",
		Body:        "This is a test email.",
		ContentType: TextPlain,
		SmtpAuthentication: SmtpAuthentication{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user",
			Password: "pass",
		},
	}
	// Chamando a função SendEmails
	err := mailer.SendEmails(1, 1, 2, 5*time.Second)
	assert.Nil(t, err)

	// Aguarda as goroutines terminarem antes de realizar as asserções
	wgTest.Wait()

	muTest.Lock()
	defer muTest.Unlock()

	// Verifica se o e-mail foi capturado corretamente
	assert.Equal(t, 1, len(capturedEmails), "Deveria ter enviado um e-mail com dois destinatários")
	assert.Equal(t, "sender@example.com", capturedEmails[0].From, "Remetente incorreto")
	assert.Equal(t, "Test Email", capturedEmails[0].Subject, "Assunto incorreto")
	assert.Equal(t, "This is a test email.", string(capturedEmails[0].Text), "Corpo incorreto")
	assert.ElementsMatch(t, []string{"recipient1@example.com", "recipient2@example.com"}, capturedEmails[0].To, "Destinatários incorretos")
}
