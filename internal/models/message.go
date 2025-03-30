package models

type Attachment struct {
	FileName string `json:"fileName"`
	Data     []byte `json:"data"`
}

type Message struct {
	To          []string      `json:"to" binding:"required"`
	From        string        `json:"from" binding:"required"`
	Subject     string        `json:"subject" binding:"required"`
	WhatsappId  string        `json:"whatsapp_id"`
	Body        string        `json:"body" binding:"required"`
	Attachments *[]Attachment `json:"attachments"`
	SmtpId      string        `json:"smtp_id" binding:"required"`
}
