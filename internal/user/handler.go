package user

import "github.com/gin-gonic/gin"

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
		c.JSON(200, gin.H{"userId": userId})
	}
}
