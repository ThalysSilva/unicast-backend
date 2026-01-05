package message

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ThalysSilva/unicast-backend/internal/auth"
	"github.com/ThalysSilva/unicast-backend/internal/config/env"
	"github.com/ThalysSilva/unicast-backend/internal/encryption"
	"github.com/ThalysSilva/unicast-backend/internal/smtp"
	"github.com/ThalysSilva/unicast-backend/internal/student"
	"github.com/ThalysSilva/unicast-backend/internal/user"
	"github.com/ThalysSilva/unicast-backend/internal/whatsapp"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/ThalysSilva/unicast-backend/pkg/mailer"
)

type SendResponse struct {
	EmailsFailed   []student.Student `json:"emailsFailed"`
	WhatsappFailed []student.Student `json:"whatsappFailed"`
}

type Service interface {
	Send(ctx context.Context, message *Message) (emailsFails, whatsappFails *[]student.Student, err error)
}

type service struct {
	whatsAppRepository whatsapp.Repository
	smtpRepository     smtp.Repository
	userRepository     user.Repository
	studentRepository  student.Repository
	logRepository      LogRepository
	jweSecret          []byte
	defaultCountryCode string
}

var (
	ErrSmtpNotFound     = customerror.Make("smtp não encontrado.", 404, errors.New("ErrSmtpNotFound"))
	ErrWhatsAppNotFound = customerror.Make("whatsapp não encontrado.", 404, errors.New("ErrWhatsAppNotFound"))
	ErrStudentsNotFound = customerror.Make("estudantes não encontrado.", 404, errors.New("ErrStudentsNotFound"))
	ErrPhoneMissing     = customerror.Make("estudante sem telefone configurado", 400, errors.New("ErrPhoneMissing"))
	ErrPhoneInvalid     = customerror.Make("telefone inválido para WhatsApp", 400, errors.New("ErrPhoneInvalid"))
)

func NewMessageService(whatsAppRepository whatsapp.Repository, smtpRepository smtp.Repository, userRepository user.Repository, studentRepository student.Repository, logRepository LogRepository, jweSecret []byte) Service {
	cfg, _ := env.Load()
	defaultCountry := "55"
	if cfg != nil && cfg.Defaults.CountryCode != "" {
		defaultCountry = cfg.Defaults.CountryCode
	}

	return &service{
		whatsAppRepository: whatsAppRepository,
		smtpRepository:     smtpRepository,
		userRepository:     userRepository,
		studentRepository:  studentRepository,
		logRepository:      logRepository,
		jweSecret:          jweSecret,
		defaultCountryCode: defaultCountry,
	}
}

func extractEmailFailedStudents(err error, students []*student.Student) ([]student.Student, error) {
	var emailErr *mailer.EmailSentError
	if errors.As(err, &emailErr) {
		emailToStudent := make(map[string]*student.Student, len(students))
		for _, student := range students {
			if student.Email != nil {
				emailToStudent[*student.Email] = student
			}
		}

		var failedStudents []student.Student
		for _, failedEmail := range emailErr.To {
			if student, exists := emailToStudent[failedEmail]; exists {
				failedStudents = append(failedStudents, *student)
			}
		}
		return failedStudents, err
	}
	return nil, err
}

func (s *service) Send(ctx context.Context, message *Message) (emailsFails, whatsappFails *[]student.Student, err error) {
	students, err := s.studentRepository.FindByIDs(ctx, message.To)
	if err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}
	if len(students) == 0 {
		return nil, nil, customerror.Trace("Send", ErrStudentsNotFound)
	}

	smtpInstance, waInstance, err := s.loadSenders(ctx, message.SmtpId, message.WhatsappId)
	if err != nil {
		return nil, nil, err
	}

	mailerAttachments, rawAttachments, attachmentNamesStr := buildAttachments(message)

	emailFailedSlice, emailErr := s.sendEmails(ctx, message, smtpInstance, mailerAttachments, students)
	emailFailedStudents := &emailFailedSlice
	whatsappFailedSlice := s.sendWhats(ctx, waInstance, students, message.Body, rawAttachments)
	whatsappFailedStudents := &whatsappFailedSlice

	s.logResults(ctx, students, emailFailedSlice, whatsappFailedSlice, message, attachmentNamesStr)

	return emailFailedStudents, whatsappFailedStudents, emailErr
}

func (s *service) loadSenders(ctx context.Context, smtpID, whatsappID string) (*smtp.Instance, *whatsapp.Instance, error) {
	smtpInstance, err := s.smtpRepository.FindByID(ctx, smtpID)
	if err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}
	if smtpInstance == nil {
		return nil, nil, customerror.Trace("Send", ErrSmtpNotFound)
	}

	waInstance, err := s.whatsAppRepository.FindByID(ctx, whatsappID)
	if err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}
	if waInstance == nil {
		return nil, nil, customerror.Trace("Send", ErrWhatsAppNotFound)
	}
	return smtpInstance, waInstance, nil
}

func buildAttachments(message *Message) ([]mailer.Attachment, []Attachment, string) {
	attachments := []mailer.Attachment{}
	raw := []Attachment{}
	var names []string
	if message.Attachments != nil {
		for _, attachment := range *message.Attachments {
			if len(attachment.Data) > 0 {
				attachments = append(attachments, mailer.Attachment{
					FileName: attachment.FileName,
					Data:     attachment.Data,
				})
			}
			raw = append(raw, attachment)
			names = append(names, attachment.FileName)
		}
	}
	if len(names) == 0 {
		return attachments, raw, ""
	}
	return attachments, raw, strings.Join(names, ",")
}

func (s *service) sendEmails(ctx context.Context, message *Message, smtpInstance *smtp.Instance, attachments []mailer.Attachment, students []*student.Student) ([]student.Student, error) {
	decryptedJwe, err := auth.DecryptJWE[auth.JwePayload](message.Jwe, s.jweSecret)
	if err != nil {
		return nil, customerror.Trace("Send", err)
	}
	decryptedSmtpPassword, err := encryption.DecryptSmtpPassword([]byte(smtpInstance.Password), []byte(decryptedJwe.SmtpKeyEncoded), []byte(smtpInstance.IV))
	if err != nil {
		return nil, customerror.Trace("Send", err)
	}
	sender := mailer.NewEmailSender(mailer.SmtpAuthentication{
		Host:     smtpInstance.Host,
		Port:     smtpInstance.Port,
		Username: smtpInstance.Email,
		Password: decryptedSmtpPassword,
	})

	if err := sender.SetData(&mailer.MailerData{
		From:        message.From,
		To:          message.To,
		Subject:     message.Subject,
		Body:        message.Body,
		Attachments: &attachments,
		ContentType: mailer.TextPlain,
	}); err != nil {
		return nil, customerror.Trace("Send", err)
	}

	emailSendErr := sender.SendEmails(4, 4, 10, 5*time.Second)
	emailFailedStudents := []student.Student{}
	if emailSendErr != nil {
		failedStudents, e := extractEmailFailedStudents(emailSendErr, students)
		if e != nil {
			return nil, customerror.Trace("Send", e)
		}
		if len(failedStudents) > 0 {
			emailFailedStudents = failedStudents
		}
	}
	return emailFailedStudents, emailSendErr
}

func (s *service) sendWhats(ctx context.Context, waInstance *whatsapp.Instance, students []*student.Student, body string, attachments []Attachment) []student.Student {
	var failed []student.Student

	for _, stud := range students {
		if stud.Phone == nil || *stud.Phone == "" {
			failed = append(failed, *stud)
			continue
		}

		normalized, err := whatsapp.NormalizeNumber(*stud.Phone, s.defaultCountryCode)
		if err != nil {
			failed = append(failed, *stud)
			continue
		}

		if err := sendWhatsAppWithRetry(waInstance.InstanceName, normalized, body, 3, 1*time.Second); err != nil {
			fmt.Printf("falha ao enviar whatsapp para %s: %v\n", *stud.Phone, err)
			failed = append(failed, *stud)
			continue
		}

		for _, att := range attachments {
			if len(att.Data) > 0 {
				if err := whatsapp.SendMedia(waInstance.InstanceName, normalized, att.FileName, att.Data, body); err != nil {
					fmt.Printf("falha ao enviar anexo via whatsapp para %s: %v\n", *stud.Phone, err)
					failed = append(failed, *stud)
					break
				}
				continue
			}
			if att.URL != "" {
				if err := whatsapp.SendMediaURL(waInstance.InstanceName, normalized, att.URL, att.FileName, body); err != nil {
					fmt.Printf("falha ao enviar anexo via url whatsapp para %s: %v\n", *stud.Phone, err)
					failed = append(failed, *stud)
					break
				}
				continue
			}
			// Se não há data nem URL, ignora o attachment
			fmt.Printf("falha ao enviar anexo via whatsapp para %s: %v\n", *stud.Phone, err)
			failed = append(failed, *stud)
			break
		}
	}

	return failed
}

func (s *service) logResults(ctx context.Context, students []*student.Student, emailFailed, whatsappFailed []student.Student, message *Message, attachmentNames string) {
	attachmentCount := 0
	if attachmentNames != "" {
		attachmentCount = len(strings.Split(attachmentNames, ","))
	}

	emailFailedSet := make(map[string]string)
	for _, s := range emailFailed {
		emailFailedSet[s.ID] = "failed to send email"
	}
	for _, stud := range students {
		errText, failed := emailFailedSet[stud.ID]
		if err := s.logRepository.Save(ctx, &Log{
			StudentID:       stud.ID,
			Channel:         ChannelEmail,
			Success:         !failed,
			ErrorText:       nullableString(errText, failed),
			Subject:         &message.Subject,
			Body:            &message.Body,
			SMTPID:          &message.SmtpId,
			AttachmentNames: nullableString(attachmentNames, attachmentCount > 0),
			AttachmentCount: attachmentCount,
		}); err != nil {
			fmt.Printf("falha ao salvar log email student %s: %v\n", stud.ID, err)
		}
	}

	whatsFailedSet := make(map[string]string)
	for _, s := range whatsappFailed {
		whatsFailedSet[s.ID] = "failed to send whatsapp"
	}
	for _, stud := range students {
		errText, failed := whatsFailedSet[stud.ID]
		if err := s.logRepository.Save(ctx, &Log{
			StudentID:          stud.ID,
			Channel:            ChannelWhatsApp,
			Success:            !failed,
			ErrorText:          nullableString(errText, failed),
			Subject:            &message.Subject,
			Body:               &message.Body,
			WhatsAppInstanceID: &message.WhatsappId,
			AttachmentNames:    nullableString(attachmentNames, attachmentCount > 0),
			AttachmentCount:    attachmentCount,
		}); err != nil {
			fmt.Printf("falha ao salvar log whatsapp student %s: %v\n", stud.ID, err)
		}
	}
}

// sendWhatsAppWithRetry encapsula retentativa simples para envio de WhatsApp.
func sendWhatsAppWithRetry(instanceID, number, body string, attempts int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		lastErr = whatsapp.SendText(instanceID, number, body)
		if lastErr == nil {
			return nil
		}
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}
	return lastErr
}

func nullableString(val string, set bool) *string {
	if !set {
		return nil
	}
	return &val
}
