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
// @Summary Send a message
// @Description Send a message via email and WhatsApp
// @OperationId sendMessage
// @Tags message
// @Accept json
// @Produce json
// @Param message body MessageInput true "Message data"
// @Success 200 {object} services.SendResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /message/send [post]
// Send handles the sending of messages via email and WhatsApp
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
