package smtp

import (
	"github.com/ThalysSilva/unicast-backend/pkg/api"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

type createInstanceInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	Jwe      string `json:"jwe" binding:"required"`
}

type Handler interface {
	Create(jweSecret []byte) gin.HandlerFunc
	GetInstances() gin.HandlerFunc
}

func NewHandler(service Service) Handler {
	return &handler{
		service: service,
	}
}

// @Summary Cria uma instância SMTP
// @Tags smtp
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body createInstanceInput true "Dados SMTP"
// @Success 200 {object} api.MessageResponse
// @Router /smtp/instance [post]
func (h *handler) Create(jweSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input createInstanceInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		userID := c.GetString("userID")
		err := h.service.Create(c.Request.Context(), jweSecret, userID, input.Jwe, input.Email, input.Password, input.Host, input.Port)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, api.MessageResponse{Message: "SMTP instance created successfully"})

	}
}

// @Summary Lista instâncias SMTP do usuário
// @Tags smtp
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} api.DefaultResponse[[]Instance]
// @Router /smtp/instance [get]
func (h *handler) GetInstances() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		instances, err := h.service.GetInstances(c.Request.Context(), userID)
		if err != nil {
			c.Error(err)
			return
		}
		items := make([]Instance, 0, len(instances))
		for _, inst := range instances {
			if inst != nil {
				items = append(items, *inst)
			}
		}
		c.JSON(200, api.DefaultResponse[[]Instance]{Message: "Instâncias listadas com sucesso", Data: items})
	}
}
