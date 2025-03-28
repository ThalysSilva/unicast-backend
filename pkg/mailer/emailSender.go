package mailer

import (
	"errors"
	"fmt"
	"github.com/jordan-wright/email"
	"net/smtp"
	"sync"
	"time"
)

type SmtpAuthentication struct {
	Host     string
	Port     int
	Username string
	Password string
}

type ContentType string

const (
	TextPlain ContentType = "text/plain"
	TextHTML  ContentType = "text/html"
)

func (c ContentType) IsValid() bool {
	switch c {
	case TextPlain, TextHTML:
		return true
	default:
		return false
	}
}

type MailerData struct {
	From               string
	To                 []string
	Subject            string
	Body               string
	ContentType        ContentType
	SmtpAuthentication SmtpAuthentication
}

type EmailSentError struct {
	From    string
	To      []string
	Subject string
	Body    string
}

type EmailToRetry struct {
	From    string
	To      string
	Subject string
	Body    string
}

func (e *EmailSentError) Error() string {
	return fmt.Sprintf("Failed to send email from %s to %v with subject %s and body %s", e.From, e.To, e.Subject, e.Body)
}
type emailPool interface {
	Send(e *email.Email, timeout time.Duration) error
	Close()
}
// newPoolFunc é uma variável interna que pode ser sobrescrita nos testes
var newPoolFunc = func(host string, pools int, auth smtp.Auth) (emailPool, error) {
	return email.NewPool(host, pools, auth)
}
// newEmailChannelsFunc é uma variável interna que pode ser sobrescrita nos testes
var newEmailChannelsFunc = func(bufferSize int) (chan *email.Email, chan *EmailToRetry) {
	return make(chan *email.Email, bufferSize), make(chan *EmailToRetry, bufferSize)
}


type EmailSender interface {
	// SendEmails envia os emails para os destinatários especificados em concorrência.
	// O número de pools para envio e retry é especificado pelos parâmetros poolsForSend e poolsForRetry, respectivamente.
	// O parâmetro groupSize especifica o número de destinatários a serem enviados em cada pool.
	// O parâmetro timeout especifica o tempo limite para o envio de cada email.
	// Se o envio falhar, os emails serão enviados novamente usando os pools de retry.
	// Se o envio falhar novamente, os emails com erro serão retornados.
	// Se o envio for bem-sucedido, nil será retornado.
	// Se houver erros, um EmailSentError será retornado com os detalhes do erro.
	SendEmails(poolsForSend int, poolsForRetry int, groupSize int, timeout time.Duration) error
}

func NewEmailSender(config SmtpAuthentication) EmailSender {
	return &MailerData{SmtpAuthentication: config}
}

var wgEmailsDispatch sync.WaitGroup
var wgEmailsRetry sync.WaitGroup
var mu sync.Mutex

func (m *MailerData) SendEmails(poolsForSend int, poolsForRetry int, groupSize int, timeout time.Duration) error {
	if !m.ContentType.IsValid() {
		return errors.New("conteúdo do email inválido")
	}
	if poolsForSend <= 0 {
		return errors.New("pools para envio devem ser maiores que 0")
	}
	if poolsForRetry <= 0 {
		return errors.New("pools para retry devem ser maiores que 0")
	}
	if groupSize <= 0 {
		return errors.New("groupSize deve ser maior que 0")
	}
	if timeout <= 0 {
		return errors.New("timeout deve ser maior que 0")
	}
	if len(m.To) == 0 {
		return errors.New("nenhum destinatário fornecido")
	}
	if len(m.From) == 0 {
		return errors.New("nenhum remetente fornecido")
	}
	if len(m.Subject) == 0 {
		return errors.New("nenhum assunto fornecido")
	}
	if len(m.Body) == 0 {
		return errors.New("nenhum corpo fornecido")
	}

	emailsChan, emailsRetryChan := newEmailChannelsFunc(len(m.To))
	emailsWithErrors := EmailSentError{
		From:    m.From,
		Subject: m.Subject,
		Body:    m.Body,
		To:      []string{},
	}

	totalPools := poolsForSend + poolsForRetry

	const maxTotalPools = 20
	if totalPools > maxTotalPools {
		return errors.New("o número total de pools (send + retry) não pode ser maior que 20")
	}

	smtpPlainAuth := smtp.PlainAuth("", m.SmtpAuthentication.Username, m.SmtpAuthentication.Password, m.SmtpAuthentication.Host)

	pool, err := newPoolFunc(
		m.SmtpAuthentication.Host+":"+fmt.Sprint(m.SmtpAuthentication.Port),
		totalPools,
		smtpPlainAuth,
	)
	if err != nil {
		return err
	}
	defer pool.Close()

	wgEmailsRetry.Add(poolsForRetry)
	for range poolsForRetry {
		go func() {
			defer wgEmailsRetry.Done()
			for retryEmail := range emailsRetryChan {
				email := &email.Email{
					From:    retryEmail.From,
					To:      []string{retryEmail.To},
					Subject: retryEmail.Subject,
				}
				switch m.ContentType {
				case TextPlain:
					email.Text = []byte(retryEmail.Body)
				case TextHTML:
					email.HTML = []byte(retryEmail.Body)
				}

				if err := pool.Send(email, timeout); err != nil {
					mu.Lock()
					emailsWithErrors.To = append(emailsWithErrors.To, retryEmail.To)
					mu.Unlock()
				}
			}
		}()
	}

	wgEmailsDispatch.Add(poolsForSend)
	for range poolsForSend {
		go func() {
			defer wgEmailsDispatch.Done()
			for email := range emailsChan {
				if err := pool.Send(email, timeout); err != nil {
					for _, emailToRetry := range email.To {
						retryEmail := &EmailToRetry{
							From:    m.From,
							To:      emailToRetry,
							Subject: m.Subject,
							Body:    m.Body,
						}
						emailsRetryChan <- retryEmail
					}
				}
			}
		}()
	}

	for i := 0; i < len(m.To); i += groupSize {
		end := min(i+groupSize, len(m.To))
		email := &email.Email{
			From:    m.From,
			To:      m.To[i:end],
			Subject: m.Subject,
		}

		switch m.ContentType {
		case TextPlain:
			email.Text = []byte(m.Body)
		case TextHTML:
			email.HTML = []byte(m.Body)
		}

		emailsChan <- email
	}

	close(emailsChan)
	wgEmailsDispatch.Wait()
	close(emailsRetryChan)
	wgEmailsRetry.Wait()

	if len(emailsWithErrors.To) > 0 {
		return &emailsWithErrors
	}

	return nil
}
