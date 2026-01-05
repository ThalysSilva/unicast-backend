package backdoor

import (
	"net/http"

	"github.com/ThalysSilva/unicast-backend/pkg/api"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/gin-gonic/gin"
)

type resetInput struct {
	Secret      string `json:"secret" binding:"required"`
	UserID      string `json:"userId"`
	Email       string `json:"email"`
	NewPassword string `json:"newPassword" binding:"required"`
}

type handler struct {
	service Service
}

type Handler interface {
	ResetPassword() gin.HandlerFunc
}

func NewHandler(service Service) Handler {
	return &handler{service: service}
}

// @Summary Reseta a senha de um usuário (backdoor)
// @Description Uso administrativo com segredo estático; requer userId ou email.
// @Tags backdoor
// @Accept json
// @Produce json
// @Param payload body resetInput true "Dados de reset"
// @Success 200 {object} api.MessageResponse
// @Failure 400 {object} api.ErrorResponse
// @Router /backdoor/reset-password [post]
func (h *handler) ResetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input resetInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		if input.UserID == "" && input.Email == "" {
			c.Error(customerror.Make("informe userId ou email", http.StatusBadRequest, nil))
			return
		}

		if err := h.service.ResetPassword(c.Request.Context(), input.Secret, input.UserID, input.Email, input.NewPassword); err != nil {
			customerror.HandleResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, api.MessageResponse{Message: "Senha atualizada com sucesso."})
	}
}
