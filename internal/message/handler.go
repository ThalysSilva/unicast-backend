package message

import (
	"net/http"

	"github.com/ThalysSilva/unicast-backend/pkg/api"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

type Handler interface {
	Send() gin.HandlerFunc
}

func NewHandler(service Service) Handler {
	return &handler{service: service}
}

// @Summary Envia uma mensagem
// @Description Envia uma mensagem via email e WhatsApp
// @OperationId sendMessage
// @Tags message
// @Accept json
// @Produce json
// @Param message body MessageInput true "Message data"
// @Success 200 {object} api.DefaultResponse[MessageDataResponse]
// @Failure 400 {object} api.ErrorResponse
// @Router /message/send [post]
// Send handles the sending of messages via email and WhatsApp
func (h *handler) Send() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input MessageInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}

		emailsFailed, whatsappFailed, err := h.service.Send(&Message{
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
			customerror.HandleResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, api.DefaultResponse[MessageDataResponse]{
			Message: "Mensagem enviada com sucesso",
			Data: MessageDataResponse{
				EmailsFailed:   *emailsFailed,
				WhatsappFailed: *whatsappFailed,
			},
		})
	}
}
