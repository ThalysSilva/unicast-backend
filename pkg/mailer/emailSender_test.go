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
	// Erro a ser retornado no envio
	sendErr error
	// Falhar apenas no primeiro envio
	failFirstSend bool
	// Contador de envios para controlar retries
	sendCount int
	// Mutex para controle de concorrência
	mu sync.Mutex
	// Emails enviados para inspeção
	sentEmails []*email.Email
}

const qtyMillisecondsToSleep = 5
const sleepTime = qtyMillisecondsToSleep * time.Millisecond

func (m *mockEmailPool) Send(e *email.Email, timeout time.Duration) error {
	time.Sleep(sleepTime)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendCount++
	m.sentEmails = append(m.sentEmails, e)
	if m.failFirstSend && m.sendCount == 1 {
		return errors.New("failed to send email on first attempt")
	}
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

// Teste de Envio de Email com sucesso
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
	sender.SetData(&MailerData{
		From:        "sender@example.com",
		To:          []string{"test123@test.com"},
		Subject:     "Teste de Envio",
		Body:        "Esse é um teste de envio.",
		ContentType: TextPlain,
	})

	err := sender.SendEmails(2, 2, 1, 5*time.Second)
	assert.Nil(t, err, "Expected no error on email send")
}

// Verificar Erro de Envio com contentType inválido
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

	sender.SetData(&MailerData{
		From:        "sender@example.com",
		To:          []string{"test123@test.com"},
		Subject:     "Teste de Envio",
		Body:        "Esse é um teste de envio.",
		ContentType: "invalid",
	})

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

	sender.SetData(&MailerData{
		From:        "sender@example.com",
		To:          []string{"test123@test.com"},
		Subject:     "Teste de Envio",
		Body:        "Esse é um teste de envio.",
		ContentType: TextPlain,
	})

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

	sender.SetData(&MailerData{
		From:        "sender@example.com",
		To:          []string{"test123@test.com"},
		Subject:     "Teste de Envio",
		Body:        "Esse é um teste de envio.",
		ContentType: TextPlain,
	})

	err := sender.SendEmails(2, 2, 1, 5*time.Second)
	assert.NotNil(t, err)
}

// / Teste de envio em massa com interceptação
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

	err := sender.SetData(&MailerData{
		From:        "sender@example.com",
		To:          recipients,
		Subject:     "Teste de Envio em Massa",
		Body:        "Esse é um teste de envio em massa.",
		ContentType: TextPlain,
	})
	assert.Nil(t, err, "Esperado sucesso ao definir dados")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := sender.SendEmails(poolsForSend, poolsForRetry, groupSize, timeout)
		assert.Nil(t, err, "Esperado sucesso ao enviar emails")
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
		assert.Equal(t, "Teste de Envio em Massa", email.Subject, "Assunto incorreto")
		assert.Equal(t, "Esse é um teste de envio em massa.", string(email.Text), "Corpo incorreto")
	}
}

// Verificar Envio em Massa sem Interceptação  verificando o tempo de execução
func TestSendEmails_Massive(t *testing.T) {
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

	senderBench := NewEmailSender(SmtpAuthentication{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
	})

	if err := senderBench.SetData(&MailerData{
		From:        "sender@example.com",
		To:          recipients,
		Subject:     "Teste de Envio em Massa",
		Body:        "Esse é um teste de envio em massa.",
		ContentType: TextPlain,
	}); err != nil {
		assert.Nil(t, err)
	}

	start := time.Now()
	err := senderBench.SendEmails(poolsForSend, poolsForRetry, groupSize, timeout)
	assert.Nil(t, err)
	duration := time.Since(start)
	expectedMinDuration := time.Duration(expectedGroups*qtyMillisecondsToSleep/poolsForSend) * time.Millisecond
	assert.GreaterOrEqual(t, duration, expectedMinDuration, "Tempo de execução deveria ser pelo menos %v", expectedMinDuration)
	t.Logf("Tempo de execução para %d destinatários: %v", numRecipients, duration)
}

// Verificar Retry Bem-Sucedido
func TestSendEmails_RetrySuccess(t *testing.T) {
	pool := &mockEmailPool{
		failFirstSend: true,
		sendErr:       nil,
	}
	newPoolFunc = func(host string, pools int, auth smtp.Auth) (emailPool, error) {
		return pool, nil
	}
	t.Cleanup(func() {
		newPoolFunc = originalNewPoolFunc
	})

	sender := NewEmailSender(SmtpAuthentication{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
	})

	sender.SetData(&MailerData{
		From:        "sender@example.com",
		To:          []string{"recipient1@example.com", "recipient2@example.com"},
		Subject:     "Assunto de Teste de Retry",
		Body:        "Essa é uma mensagem de teste para retry.",
		ContentType: TextPlain,
	})

	err := sender.SendEmails(1, 1, 1, 5*time.Second)
	assert.Nil(t, err, "Esperado sucesso após retry")
	assert.Len(t, pool.sentEmails, 3)

	// O mock falha no primeiro envio, mas succeeds no retry, então o resultado deve ser nil
}

// Verificar Erro Persistente
func TestSendEmails_PersistentFailure(t *testing.T) {
	newPoolFunc = func(host string, pools int, auth smtp.Auth) (emailPool, error) {
		return &mockEmailPool{
			sendErr: errors.New("falha de erro persistente"),
		}, nil
	}
	t.Cleanup(func() {
		newPoolFunc = originalNewPoolFunc
	})

	sender := NewEmailSender(SmtpAuthentication{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
	})

	sender.SetData(&MailerData{
		From:        "sender@example.com",
		To:          []string{"recipient1@example.com", "recipient2@example.com"},
		Subject:     "Assunto de Teste de Falha Persistente",
		Body:        "Essa é uma mensagem de teste de falha persistente.",
		ContentType: TextPlain,
	})

	err := sender.SendEmails(1, 1, 1, 5*time.Second)
	assert.NotNil(t, err, "Esperado erro após falha persistente")
	assert.Equal(t, "todos os emails falharam", err.Error())
}

func TestSendEmails_InterceptedEmailIncludesAttachments(t *testing.T) {
	newPoolFunc = mockNewPoolFunc
	t.Cleanup(func() {
		newPoolFunc = originalNewPoolFunc
	})

	interceptChan := make(chan *email.Email, 1)
	sender := NewEmailSender(SmtpAuthentication{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
	}, WithInterceptChan(interceptChan))

	err := sender.SetData(&MailerData{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Teste com anexo",
		Body:    "Esse email deve conter anexo.",
		Attachments: &[]Attachment{{
			FileName: "arquivo.txt",
			Data:     []byte("conteudo do anexo"),
		}},
		ContentType: TextPlain,
	})
	assert.Nil(t, err)

	err = sender.SendEmails(1, 1, 1, 5*time.Second)
	assert.Nil(t, err)

	var captured *email.Email
	for msg := range interceptChan {
		captured = msg
	}

	if assert.NotNil(t, captured) {
		raw, bytesErr := captured.Bytes()
		assert.Nil(t, bytesErr)
		assert.Contains(t, string(raw), "filename=\"arquivo.txt\"")
	}
}

func TestSendEmails_RetryPreservesAttachments(t *testing.T) {
	pool := &mockEmailPool{
		failFirstSend: true,
		sendErr:       nil,
	}
	newPoolFunc = func(host string, pools int, auth smtp.Auth) (emailPool, error) {
		return pool, nil
	}
	t.Cleanup(func() {
		newPoolFunc = originalNewPoolFunc
	})

	sender := NewEmailSender(SmtpAuthentication{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
	})

	err := sender.SetData(&MailerData{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Retry com anexo",
		Body:    "Corpo",
		Attachments: &[]Attachment{{
			FileName: "retry.txt",
			Data:     []byte("retry attachment"),
		}},
		ContentType: TextPlain,
	})
	assert.Nil(t, err)

	err = sender.SendEmails(1, 1, 1, 5*time.Second)
	assert.Nil(t, err)

	if assert.GreaterOrEqual(t, len(pool.sentEmails), 2) {
		raw, bytesErr := pool.sentEmails[1].Bytes()
		assert.Nil(t, bytesErr)
		assert.Contains(t, string(raw), "filename=\"retry.txt\"")
	}
}
