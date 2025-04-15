package smtp

import "github.com/gin-gonic/gin"

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
		c.JSON(200, gin.H{"message": "SMTP instance created successfully"})

	}
}

func (h *handler) GetInstances() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for getting SMTP instances
	}
}
