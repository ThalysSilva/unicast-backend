package message

import "github.com/ThalysSilva/unicast-backend/internal/student"

type Attachment struct {
	FileName string `json:"fileName"`
	Data     []byte `json:"data,omitempty"`
	URL      string `json:"url,omitempty"`
}

type Message struct {
	Jwe         string        `json:"jwe"`
	To          []string      `json:"to" binding:"required"`
	From        string        `json:"from" binding:"required"`
	Subject     string        `json:"subject" binding:"required"`
	WhatsappId  string        `json:"whatsapp_id"`
	Body        string        `json:"body" binding:"required"`
	Attachments *[]Attachment `json:"attachments"`
	SmtpId      string        `json:"smtp_id" binding:"required"`
}

type MessageInput struct {
	Jwe         string        `json:"jwe" binding:"required"`
	SmtpId      string        `json:"smtp_id" binding:"required"`
	WhatsappId  string        `json:"whatsapp_id" binding:"required"`
	Subject     string        `json:"subject" binding:"required"`
	Body        string        `json:"body" binding:"required"`
	To          []string      `json:"to" binding:"required"`
	From        string        `json:"from" binding:"required"`
	Attachments *[]Attachment `json:"attachments"`
}

type MessageDataResponse struct {
	EmailsFailed   []student.Student `json:"emailsFailed"`
	WhatsappFailed []student.Student `json:"whatsappFailed"`
}
