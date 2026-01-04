package invite

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

type createInviteInput struct {
	ExpiresAt *time.Time `json:"expiresAt"`
}

type selfRegisterInput struct {
	StudentID string `json:"studentId" binding:"required"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

type Handler interface {
	Create() gin.HandlerFunc
	SelfRegister() gin.HandlerFunc
}

func NewHandler(service Service) Handler {
	return &handler{
		service: service,
	}
}

func (h *handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.Param("courseId")
		userID := c.GetString("userID")

		var input createInviteInput
		if c.Request.ContentLength > 0 {
			if err := c.ShouldBindJSON(&input); err != nil && err != io.EOF {
				c.Error(err)
				return
			}
		}

		invite, err := h.service.Create(c.Request.Context(), courseID, userID, input.ExpiresAt)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(200, gin.H{
			"message": "Convite criado com sucesso",
			"code":    invite.Code,
		})
	}
}

func (h *handler) SelfRegister() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input selfRegisterInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}

		code := c.Param("code")

		err := h.service.SelfRegister(c.Request.Context(), code, input.StudentID, input.Name, input.Phone, input.Email)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(200, gin.H{"message": "Cadastro concluído com sucesso"})
	}
}
