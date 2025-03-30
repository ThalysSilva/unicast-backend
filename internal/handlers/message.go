package handlers

import (
	"net/http"
	"unicast-api/internal/models"
	"unicast-api/internal/services"
	"unicast-api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type MessageInput struct {
	SmtpId      string               `json:"smtp_id" binding:"required"`
	WhatsappId  string               `json:"whatsapp_id"`
	Subject     string               `json:"subject" binding:"required"`
	Body        string               `json:"body" binding:"required"`
	To          []string             `json:"to" binding:"required"`
	From        string               `json:"from" binding:"required"`
	Attachments *[]models.Attachment `json:"attachment"`
}

func Send(messageService services.MessageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input MessageInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}

		emailsFailed, whatsappFailed, err := messageService.Send(&models.Message{
			To:          input.To,
			From:        input.From,
			Subject:     input.Subject,
			WhatsappId:  input.WhatsappId,
			Body:        input.Body,
			Attachments: input.Attachments,
			SmtpId:      input.SmtpId,
		})
		if err != nil {
			utils.HandleErrorResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":  "Mensagem enviada com sucesso",
			"emailsFailed":   emailsFailed,
			"whatsappFailed": whatsappFailed,
		})
	}
}
