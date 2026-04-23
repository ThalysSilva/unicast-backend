package message

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
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

type Service interface {
	Send(ctx context.Context, message *Message) (emailsFails, whatsappFails *[]student.Student, err error)
}

type service struct {
	whatsAppRepository whatsapp.Repository
	smtpService        smtp.Service
	smtpRepository     smtp.Repository
	userRepository     user.Repository
	studentRepository  student.Repository
	logRepository      LogRepository
	jweSecret          []byte
	defaultCountryCode string
}

var (
	ErrSmtpNotFound      = customerror.Make("smtp não encontrado.", 404, errors.New("ErrSmtpNotFound"))
	ErrWhatsAppNotFound  = customerror.Make("whatsapp não encontrado.", 404, errors.New("ErrWhatsAppNotFound"))
	ErrStudentsNotFound  = customerror.Make("estudantes não encontrado.", 404, errors.New("ErrStudentsNotFound"))
	ErrNoChannelSelected = customerror.Make("selecione ao menos um canal de envio", 400, errors.New("ErrNoChannelSelected"))
	ErrPhoneMissing      = customerror.Make("estudante sem telefone configurado", 400, errors.New("ErrPhoneMissing"))
	ErrPhoneInvalid      = customerror.Make("telefone inválido para WhatsApp", 400, errors.New("ErrPhoneInvalid"))
	ErrInvalidAttachment = customerror.Make("anexo inválido", 400, errors.New("ErrInvalidAttachment"))
	httpClient           = http.DefaultClient
)

const (
	maxAttachmentCount     = 5
	maxAttachmentBytes     = 10 * 1024 * 1024
	maxEmailTotalBytes     = 25 * 1024 * 1024
	maxWhatsAppTotalBytes  = 15 * 1024 * 1024
)

var blockedAttachmentExtensions = map[string]struct{}{
	".apk": {}, ".app": {}, ".bat": {}, ".cmd": {}, ".com": {}, ".dll": {}, ".dmg": {},
	".exe": {}, ".hta": {}, ".iso": {}, ".jar": {}, ".js": {}, ".msi": {}, ".ps1": {},
	".scr": {}, ".sh": {}, ".vbs": {}, ".wsf": {},
}

var allowedAttachmentExtensions = map[string]struct{}{
	".csv": {}, ".doc": {}, ".docx": {}, ".jpeg": {}, ".jpg": {}, ".mp3": {}, ".mp4": {},
	".ogg": {}, ".pdf": {}, ".png": {}, ".ppt": {}, ".pptx": {}, ".txt": {}, ".webp": {},
	".xls": {}, ".xlsx": {},
}

func NewMessageService(whatsAppRepository whatsapp.Repository, smtpService smtp.Service, smtpRepository smtp.Repository, userRepository user.Repository, studentRepository student.Repository, logRepository LogRepository, jweSecret []byte) Service {
	cfg, _ := env.Load()
	defaultCountry := "55"
	if cfg != nil && cfg.Defaults.CountryCode != "" {
		defaultCountry = cfg.Defaults.CountryCode
	}

	return &service{
		whatsAppRepository: whatsAppRepository,
		smtpService:        smtpService,
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
	message.To = uniqueIDs(message.To)
	students, err := s.studentRepository.FindByIDs(ctx, message.UserID, message.To)
	if err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}
	if len(students) == 0 {
		return nil, nil, customerror.Trace("Send", ErrStudentsNotFound)
	}
	if err := validateAttachmentCount(message.Attachments); err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}

	smtpInstance, waInstance, err := s.loadSenders(ctx, message.UserID, message.SmtpId, message.WhatsappId)
	if err != nil {
		return nil, nil, err
	}

	rawAttachments, attachmentNamesStr, err := buildWhatsAppAttachments(message)
	if err != nil {
		return nil, nil, customerror.Trace("Send", err)
	}

	emailFailedSlice := []student.Student{}
	var emailErr error
	if smtpInstance != nil {
		mailerAttachments, err := buildEmailAttachments(ctx, message)
		if err != nil {
			return nil, nil, customerror.Trace("Send", err)
		}
		emailFailedSlice, emailErr = s.sendEmails(ctx, message, smtpInstance, mailerAttachments, students)
	}

	whatsappFailedSlice := []student.Student{}
	if waInstance != nil {
		whatsappFailedSlice = s.sendWhats(ctx, waInstance, students, formatWhatsAppBody(message.Subject, message.Body), rawAttachments)
	}

	s.logResults(ctx, students, emailFailedSlice, whatsappFailedSlice, message, attachmentNamesStr)

	return &emailFailedSlice, &whatsappFailedSlice, emailErr
}

func uniqueIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	unique := make([]string, 0, len(ids))

	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}

		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	return unique
}

func (s *service) loadSenders(ctx context.Context, userID, smtpID, whatsappID string) (*smtp.Instance, *whatsapp.Instance, error) {
	if smtpID == "" && whatsappID == "" {
		return nil, nil, customerror.Trace("Send", ErrNoChannelSelected)
	}

	var smtpInstance *smtp.Instance
	var waInstance *whatsapp.Instance

	if smtpID != "" {
		instance, err := s.smtpRepository.FindByID(ctx, smtpID)
		if err != nil {
			return nil, nil, customerror.Trace("Send", err)
		}
		if instance == nil {
			return nil, nil, customerror.Trace("Send", ErrSmtpNotFound)
		}
		if instance.UserID != userID {
			return nil, nil, customerror.Trace("Send", ErrSmtpNotFound)
		}
		smtpInstance = instance
	}

	if whatsappID != "" {
		instance, err := s.whatsAppRepository.FindByID(ctx, whatsappID)
		if err != nil {
			return nil, nil, customerror.Trace("Send", err)
		}
		if instance == nil {
			return nil, nil, customerror.Trace("Send", ErrWhatsAppNotFound)
		}
		if instance.UserID != userID {
			return nil, nil, customerror.Trace("Send", ErrWhatsAppNotFound)
		}
		waInstance = instance
	}

	return smtpInstance, waInstance, nil
}

func buildWhatsAppAttachments(message *Message) ([]Attachment, string, error) {
	raw := []Attachment{}
	var names []string
	totalBytes := 0
	if message.Attachments != nil {
		for _, attachment := range *message.Attachments {
			if err := validateAttachmentMetadata(attachment.FileName); err != nil {
				return nil, "", err
			}
			if len(attachment.Data) > 0 {
				if err := validateAttachmentData(attachment.FileName, attachment.Data, maxAttachmentBytes); err != nil {
					return nil, "", err
				}
				totalBytes += len(attachment.Data)
				if totalBytes > maxWhatsAppTotalBytes {
					return nil, "", customerror.Make("anexos excedem o limite total do WhatsApp", http.StatusBadRequest, errors.New("whatsapp attachment total too large"))
				}
			}
			raw = append(raw, attachment)
			names = append(names, attachment.FileName)
		}
	}
	if len(names) == 0 {
		return raw, "", nil
	}
	return raw, strings.Join(names, ","), nil
}

func buildEmailAttachments(ctx context.Context, message *Message) ([]mailer.Attachment, error) {
	attachments := []mailer.Attachment{}
	totalBytes := 0
	if message.Attachments == nil {
		return attachments, nil
	}

	for _, attachment := range *message.Attachments {
		if err := validateAttachmentMetadata(attachment.FileName); err != nil {
			return nil, err
		}
		switch {
		case len(attachment.Data) > 0:
			if err := validateAttachmentData(attachment.FileName, attachment.Data, maxAttachmentBytes); err != nil {
				return nil, err
			}
			totalBytes += len(attachment.Data)
			if totalBytes > maxEmailTotalBytes {
				return nil, customerror.Make("anexos excedem o limite total do email", http.StatusBadRequest, errors.New("email attachment total too large"))
			}
			attachments = append(attachments, mailer.Attachment{
				FileName: attachment.FileName,
				Data:     attachment.Data,
			})
		case attachment.URL != "":
			data, err := fetchAttachmentData(ctx, attachment.URL)
			if err != nil {
				return nil, customerror.Make("falha ao baixar anexo para email", http.StatusBadRequest, err)
			}
			if err := validateAttachmentData(attachment.FileName, data, maxAttachmentBytes); err != nil {
				return nil, err
			}
			totalBytes += len(data)
			if totalBytes > maxEmailTotalBytes {
				return nil, customerror.Make("anexos excedem o limite total do email", http.StatusBadRequest, errors.New("email attachment total too large"))
			}
			attachments = append(attachments, mailer.Attachment{
				FileName: attachment.FileName,
				Data:     data,
			})
		default:
			return nil, customerror.Make("anexo deve conter data ou url", http.StatusBadRequest, errors.New("attachment missing data and url"))
		}
	}

	return attachments, nil
}

func fetchAttachmentData(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("attachment download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func validateAttachmentCount(attachments *[]Attachment) error {
	if attachments == nil {
		return nil
	}
	if len(*attachments) > maxAttachmentCount {
		return customerror.Make("quantidade de anexos excede o limite permitido", http.StatusBadRequest, errors.New("too many attachments"))
	}
	return nil
}

func validateAttachmentMetadata(fileName string) error {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return customerror.Make("anexo deve ter um nome de arquivo", http.StatusBadRequest, errors.New("attachment filename required"))
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		return customerror.Make("anexo deve ter uma extensão permitida", http.StatusBadRequest, errors.New("attachment extension required"))
	}
	if _, blocked := blockedAttachmentExtensions[ext]; blocked {
		return customerror.Make("tipo de arquivo não permitido", http.StatusBadRequest, errors.New("blocked attachment extension"))
	}
	if _, allowed := allowedAttachmentExtensions[ext]; !allowed {
		return customerror.Make("tipo de arquivo não permitido", http.StatusBadRequest, errors.New("attachment extension not allowed"))
	}

	return nil
}

func validateAttachmentData(fileName string, data []byte, maxBytes int) error {
	if len(data) == 0 {
		return customerror.Make("anexo sem conteúdo", http.StatusBadRequest, errors.New("empty attachment data"))
	}
	if len(data) > maxBytes {
		return customerror.Make("anexo excede o tamanho máximo permitido", http.StatusBadRequest, errors.New("attachment too large"))
	}
	if err := validateAttachmentDetectedType(fileName, data); err != nil {
		return err
	}
	return nil
}

func validateAttachmentDetectedType(fileName string, data []byte) error {
	ext := strings.ToLower(filepath.Ext(fileName))
	detected := http.DetectContentType(data)
	if idx := strings.Index(detected, ";"); idx >= 0 {
		detected = detected[:idx]
	}

	expected := mime.TypeByExtension(ext)
	if idx := strings.Index(expected, ";"); idx >= 0 {
		expected = expected[:idx]
	}

	if detected == "application/octet-stream" || detected == "text/plain" {
		return nil
	}
	if expected != "" && detected != expected {
		return customerror.Make("conteúdo do anexo não corresponde ao tipo permitido", http.StatusBadRequest, errors.New("attachment mime mismatch"))
	}

	return nil
}

func (s *service) sendEmails(ctx context.Context, message *Message, smtpInstance *smtp.Instance, attachments []mailer.Attachment, students []*student.Student) ([]student.Student, error) {
	from := smtpInstance.Email
	if message.From != "" {
		from = message.From
	}

	recipients := make([]string, 0, len(students))
	emailFailedStudents := make([]student.Student, 0)
	for _, stud := range students {
		if stud.Email == nil || *stud.Email == "" {
			emailFailedStudents = append(emailFailedStudents, *stud)
			continue
		}
		recipients = append(recipients, *stud.Email)
	}

	if len(recipients) == 0 {
		return emailFailedStudents, nil
	}

	mailData := &mailer.MailerData{
		From:        from,
		To:          recipients,
		Subject:     message.Subject,
		Body:        message.Body,
		Attachments: &attachments,
		ContentType: mailer.TextPlain,
	}

	if smtpInstance.AuthMode == smtp.AuthModeOAuth {
		if err := s.sendOAuthEmail(ctx, smtpInstance, mailData); err != nil {
			emailFailedStudents = studentsToValues(students)
			return emailFailedStudents, err
		}
		return emailFailedStudents, nil
	}

	decryptedJwe, err := auth.DecryptJWE[auth.JwePayload](message.Jwe, s.jweSecret)
	if err != nil {
		return nil, customerror.Trace("Send", err)
	}
	smtpKey, err := base64.StdEncoding.DecodeString(decryptedJwe.SmtpKeyEncoded)
	if err != nil {
		return nil, customerror.Trace("Send", err)
	}

	decryptedSmtpPassword, err := encryption.DecryptSmtpPassword(smtpInstance.Password, smtpKey, smtpInstance.IV)
	if err != nil {
		return nil, customerror.Trace("Send", err)
	}
	sender := mailer.NewEmailSender(mailer.SmtpAuthentication{
		Host:     smtpInstance.Host,
		Port:     smtpInstance.Port,
		Username: smtpInstance.Email,
		Password: decryptedSmtpPassword,
	})

	if err := sender.SetData(mailData); err != nil {
		return nil, customerror.Trace("Send", err)
	}

	emailSendErr := sender.SendEmails(4, 4, 10, 5*time.Second)
	if emailSendErr != nil {
		failedStudents, e := extractEmailFailedStudents(emailSendErr, students)
		if e != nil {
			return nil, customerror.Trace("Send", e)
		}
		if len(failedStudents) > 0 {
			emailFailedStudents = failedStudents
		}
		return emailFailedStudents, emailSendErr
	}
	return emailFailedStudents, emailSendErr
}

func (s *service) sendOAuthEmail(ctx context.Context, smtpInstance *smtp.Instance, data *mailer.MailerData) error {
	accessToken, err := s.smtpService.RefreshOAuthAccessToken(ctx, smtpInstance)
	if err != nil {
		return customerror.Trace("Send", err)
	}

	switch smtpInstance.Provider {
	case smtp.ProviderGoogle:
		if err := mailer.SendWithGmailAPI(accessToken, data); err != nil {
			return customerror.Trace("Send", err)
		}
		return nil
	default:
		return customerror.Trace("Send", customerror.Make("provedor OAuth de email inválido", 400, errors.New("invalidOAuthProvider")))
	}
}

func studentsToValues(students []*student.Student) []student.Student {
	out := make([]student.Student, 0, len(students))
	for _, stud := range students {
		if stud != nil {
			out = append(out, *stud)
		}
	}
	return out
}

func formatWhatsAppBody(subject, body string) string {
	subject = strings.TrimSpace(strings.ReplaceAll(subject, "\n", " "))
	if subject == "" {
		return body
	}
	return fmt.Sprintf("*%s*\n\n%s", subject, body)
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
			log.Printf("falha ao enviar whatsapp para %s: %v", *stud.Phone, err)
			failed = append(failed, *stud)
			continue
		}

		for _, att := range attachments {
			if len(att.Data) > 0 {
				if _, err := whatsapp.SendMedia(waInstance.InstanceName, normalized, att.FileName, att.Data, ""); err != nil {
					log.Printf("falha ao enviar anexo via whatsapp para %s: %v", *stud.Phone, err)
					failed = append(failed, *stud)
					break
				}
				continue
			}
			if att.URL != "" {
				if _, err := whatsapp.SendMediaURL(waInstance.InstanceName, normalized, att.URL, att.FileName, ""); err != nil {
					log.Printf("falha ao enviar anexo via url whatsapp para %s: %v", *stud.Phone, err)
					failed = append(failed, *stud)
					break
				}
				continue
			}
			// Se não há data nem URL, ignora o attachment
			log.Printf("falha ao enviar anexo via whatsapp para %s: anexo sem data e sem URL", *stud.Phone)
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

	if message.SmtpId != "" {
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
				log.Printf("falha ao salvar log email student %s: %v", stud.ID, err)
			}
		}
	}

	if message.WhatsappId != "" {
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
				log.Printf("falha ao salvar log whatsapp student %s: %v", stud.ID, err)
			}
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
