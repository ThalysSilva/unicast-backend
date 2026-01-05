package invite

import (
	"io"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/api"
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

// @Summary Cria um convite para um curso
// @Tags invite
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param courseId path string true "Course ID"
// @Param body body createInviteInput false "Expiração opcional"
// @Success 200 {object} api.DefaultResponse[map[string]string]
// @Router /invite/{courseId} [post]
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

		data := map[string]string{"code": invite.Code}
		c.JSON(200, api.DefaultResponse[map[string]string]{Message: "Convite criado com sucesso", Data: data})
	}
}

// @Summary Auto-registro de aluno via convite
// @Tags invite
// @Accept json
// @Produce json
// @Param code path string true "Código do convite"
// @Param body body selfRegisterInput true "Dados do aluno"
// @Success 200 {object} api.MessageResponse
// @Router /invite/self-register/{code} [post]
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

		c.JSON(200, api.MessageResponse{Message: "Cadastro concluído com sucesso"})
	}
}
