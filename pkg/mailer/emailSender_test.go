package mailer

import (
	"errors"
	"fmt"
	"math"
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

const qtyMillisecondsToSleep = 5
const sleepTime = qtyMillisecondsToSleep * time.Millisecond

func (m *mockEmailPool) Send(e *email.Email, timeout time.Duration) error {
	time.Sleep(sleepTime)
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

	sender := NewEmailSender(SmtpAuthentication{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
	})

	// Configurar os dados
	senderImpl := sender.(*emailSenderImpl)
	senderImpl.data.From = "sender@example.com"
	senderImpl.data.To = []string{"test123@test.com"}
	senderImpl.data.Subject = "Test Subject"
	senderImpl.data.Body = "This is a  test email."
	senderImpl.data.ContentType = TextPlain
	err := sender.SendEmails(2, 2, 1, 5*time.Second)
	assert.Nil(t, err, "Expected no error on email send")
}

func TestSendEmails_InvalidContentType(t *testing.T) {
	newPoolFunc = mockNewPoolFunc
	t.Cleanup(func() {
		newPoolFunc = originalNewPoolFunc
	})

	sender := NewEmailSender(SmtpAuthentication{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
	})

	senderImpl := sender.(*emailSenderImpl)
	senderImpl.data.From = "sender@example.com"
	senderImpl.data.To = []string{"recipient@example.com"}
	senderImpl.data.Subject = "Test Subject"
	senderImpl.data.Body = "Test Body"
	senderImpl.data.ContentType = "invalid" // ContentType inválido

	err := sender.SendEmails(2, 2, 1, 5*time.Second)
	assert.NotNil(t, err)
	assert.Equal(t, "conteúdo do email inválido", err.Error())
}

func TestSendEmails_FailureOnPoolCreation(t *testing.T) {
	newPoolFunc = mockNewPoolFuncWithError

	sender := NewEmailSender(SmtpAuthentication{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
	})

	// Configurar os dados
	senderImpl := sender.(*emailSenderImpl)
	senderImpl.data.From = "sender@example.com"
	senderImpl.data.To = []string{"test123@test.com"}
	senderImpl.data.Subject = "Test Subject"
	senderImpl.data.Body = "This is a  test email."
	senderImpl.data.ContentType = TextPlain

	err := sender.SendEmails(2, 2, 1, 5*time.Second)
	assert.NotNil(t, err)
	assert.Equal(t, "failed to create pool", err.Error())
}

func TestSendEmails_FailureOnSend(t *testing.T) {
	newPoolFunc = func(host string, pools int, auth smtp.Auth) (emailPool, error) {
		return &mockEmailPool{sendErr: errors.New("failed to send email")}, nil
	}

	sender := NewEmailSender(SmtpAuthentication{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
	})

	// Configurar os dados
	senderImpl := sender.(*emailSenderImpl)
	senderImpl.data.From = "sender@example.com"
	senderImpl.data.To = []string{"test123@test.com"}
	senderImpl.data.Subject = "Test Subject"
	senderImpl.data.Body = "This is a  test email."
	senderImpl.data.ContentType = TextPlain

	err := sender.SendEmails(2, 2, 1, 5*time.Second)
	assert.NotNil(t, err)
	assert.IsType(t, &EmailSentError{}, err)
}

func TestSendEmails_MassiveWithInterception(t *testing.T) {
	newPoolFunc = mockNewPoolFunc
	t.Cleanup(func() {
		newPoolFunc = originalNewPoolFunc
	})

	numRecipients := 1000
	groupSize := 50
	expectedGroups := int(math.Floor(float64(numRecipients) / float64(groupSize)))
	poolsForSend := 4
	poolsForRetry := 4
	timeout := 5 * time.Second

	recipients := make([]string, numRecipients)
	for i := range numRecipients {
		recipients[i] = fmt.Sprintf("recipient%d@example.com", i)
	}

	interceptChan := make(chan *email.Email, expectedGroups)
	sender := NewEmailSender(SmtpAuthentication{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
	}, WithInterceptChan(interceptChan))

	senderImpl := sender.(*emailSenderImpl)
	senderImpl.data.From = "sender@example.com"
	senderImpl.data.To = recipients
	senderImpl.data.Subject = "Massive Test Subject"
	senderImpl.data.Body = "This is a massive test email."
	senderImpl.data.ContentType = TextPlain

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := sender.SendEmails(poolsForSend, poolsForRetry, groupSize, timeout)
		assert.Nil(t, err, "Expected no error on massive email send")
	}()

	wg.Wait()

	capturedEmails := make([]*email.Email, 0, expectedGroups)
	for email := range interceptChan {
		capturedEmails = append(capturedEmails, email)
	}

	assert.Equal(t, expectedGroups, len(capturedEmails), fmt.Sprintf("Deveria ter capturado %d grupos de e-mails", expectedGroups))

	for i, email := range capturedEmails {
		startIdx := i * groupSize
		endIdx := min(startIdx+groupSize, numRecipients)
		expectedTo := recipients[startIdx:endIdx]
		assert.ElementsMatch(t, expectedTo, email.To, fmt.Sprintf("Grupo %d deveria ter os destinatários corretos", i))
		assert.Equal(t, "sender@example.com", email.From, "Remetente incorreto")
		assert.Equal(t, "Massive Test Subject", email.Subject, "Assunto incorreto")
		assert.Equal(t, "This is a massive test email.", string(email.Text), "Corpo incorreto")
	}

	senderBench := NewEmailSender(SmtpAuthentication{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
	})

	senderImpl2 := senderBench.(*emailSenderImpl)
	senderImpl2.data.From = "sender@example.com"
	senderImpl2.data.To = recipients
	senderImpl2.data.Subject = "Massive Test Subject"
	senderImpl2.data.Body = "This is a massive test email."
	senderImpl2.data.ContentType = TextPlain

	start := time.Now()
	err := senderBench.SendEmails(poolsForSend, poolsForRetry, groupSize, timeout)
	duration := time.Since(start)
	assert.Nil(t, err)
	expectedMinDuration := time.Duration(expectedGroups*qtyMillisecondsToSleep/poolsForSend) * time.Millisecond
	assert.GreaterOrEqual(t, duration, expectedMinDuration, "Tempo de execução deveria ser pelo menos %v", expectedMinDuration)
	t.Logf("Tempo de execução para %d destinatários: %v", numRecipients, duration)

}
