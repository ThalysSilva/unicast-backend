package handlers

import (
	"net/http"
	"unicast-api/internal/models"
	"unicast-api/internal/models/entities"
	"unicast-api/internal/services"
	"unicast-api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type MessageInput struct {
	Jwe         string               `json:"jwe" binding:"required"`
	SmtpId      string               `json:"smtp_id" binding:"required"`
	WhatsappId  string               `json:"whatsapp_id" binding:"required"`
	Subject     string               `json:"subject" binding:"required"`
	Body        string               `json:"body" binding:"required"`
	To          []string             `json:"to" binding:"required"`
	From        string               `json:"from" binding:"required"`
	Attachments *[]models.Attachment `json:"attachment"`
}

type MessageDataResponse struct {
	EmailsFailed   []entities.Student `json:"emailsFailed"`
	WhatsappFailed []entities.Student `json:"whatsappFailed"`
}

// @Summary Envia uma mensagem
// @Description Envia uma mensagem via email e WhatsApp
// @OperationId sendMessage
// @Tags message
// @Accept json
// @Produce json
// @Param message body MessageInput true "Message data"
// @Success 200 {object} models.DefaultResponse[MessageDataResponse]
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
			Jwe:         input.Jwe,
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
		c.JSON(http.StatusOK, models.DefaultResponse[MessageDataResponse]{
			Message: "Mensagem enviada com sucesso",
			Data: MessageDataResponse{
				EmailsFailed:   *emailsFailed,
				WhatsappFailed: *whatsappFailed,
			},
		})
	}
}
