package message

type Attachment struct {
	FileName string `json:"fileName"`
	Data     []byte `json:"data,omitempty"`
	URL      string `json:"url,omitempty"`
}

type Message struct {
	Jwe         string        `json:"jwe"`
	To          []string      `json:"to" binding:"required"`
	From        string        `json:"from"`
	Subject     string        `json:"subject" binding:"required"`
	WhatsappId  string        `json:"whatsapp_id"`
	Body        string        `json:"body" binding:"required"`
	Attachments *[]Attachment `json:"attachments"`
	SmtpId      string        `json:"smtp_id"`
}

type MessageInput struct {
	Jwe         string        `json:"jwe"`
	SmtpId      string        `json:"smtp_id"`
	WhatsappId  string        `json:"whatsapp_id"`
	Subject     string        `json:"subject" binding:"required"`
	Body        string        `json:"body" binding:"required"`
	To          []string      `json:"to" binding:"required"`
	From        string        `json:"from"`
	Attachments *[]Attachment `json:"attachments"`
}

type FailedRecipient struct {
	ID        string `json:"id"`
	StudentID string `json:"studentId"`
}

type MessageDataResponse struct {
	EmailsFailed   []FailedRecipient `json:"emailsFailed"`
	WhatsappFailed []FailedRecipient `json:"whatsappFailed"`
}
