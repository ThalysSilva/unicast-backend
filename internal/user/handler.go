package user

import (
	"github.com/ThalysSilva/unicast-backend/pkg/api"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

type createUserInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type Handler interface {
	Create() gin.HandlerFunc
}

func NewHandler(service Service) Handler {
	return &handler{
		service: service,
	}
}

// @Summary Cria um usuário
// @Tags user
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body createUserInput true "Dados do usuário"
// @Success 200 {object} api.DefaultResponse[map[string]string]
// @Router /user/create [post]
func (h *handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input createUserInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}

		userId, err := h.service.Create(c.Request.Context(), input.Name, input.Email, input.Password)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, api.DefaultResponse[map[string]string]{Message: "Usuário criado com sucesso", Data: map[string]string{"userId": userId}})
	}
}
