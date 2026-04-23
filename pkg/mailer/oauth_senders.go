package mailer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jordan-wright/email"
)

func SendWithGmailAPI(accessToken string, data *MailerData) error {
	msg, err := buildMessageBytes(data)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]string{
		"raw": base64.RawURLEncoding.EncodeToString(msg),
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "https://gmail.googleapis.com/gmail/v1/users/me/messages/send", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := ioReadAll(resp.Body)
		return fmt.Errorf("gmail send failed: %s", strings.TrimSpace(string(body)))
	}
	return nil
}

func buildMessageBytes(data *MailerData) ([]byte, error) {
	msg := &email.Email{
		From:    data.From,
		To:      data.To,
		Subject: data.Subject,
	}
	switch data.ContentType {
	case TextHTML:
		msg.HTML = []byte(data.Body)
	default:
		msg.Text = []byte(data.Body)
	}
	if err := attachEmailAttachments(msg, data.Attachments); err != nil {
		return nil, err
	}
	return msg.Bytes()
}

func ioReadAll(body io.Reader) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(body)
	return buf.Bytes(), err
}
