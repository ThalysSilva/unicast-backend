package invite

import (
	"io"
	"strings"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/api"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
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
	Consent   bool   `json:"consent"`
}

type Handler interface {
	Create() gin.HandlerFunc
	GetCurrent() gin.HandlerFunc
	ListByDiscipline() gin.HandlerFunc
	Delete() gin.HandlerFunc
	SelfRegister() gin.HandlerFunc
}

func NewHandler(service Service) Handler {
	return &handler{
		service: service,
	}
}

// @Summary Cria um convite para uma disciplina
// @Tags invite
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Security BearerAuth
// @Param disciplineId path string true "Discipline ID"
// @Param body body createInviteInput false "Expiração opcional"
// @Success 201 {object} api.DefaultResponse[Invite]
// @Router /invite/{disciplineId} [post]
func (h *handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		disciplineID := c.Param("disciplineId")
		userID := c.GetString("userID")

		var input createInviteInput
		if c.Request.ContentLength > 0 {
			if err := c.ShouldBindJSON(&input); err != nil && err != io.EOF {
				c.Error(err)
				return
			}
		}

		invite, err := h.service.Create(c.Request.Context(), disciplineID, userID, input.ExpiresAt)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}

		c.JSON(201, api.DefaultResponse[*Invite]{Message: "Convite criado com sucesso", Data: invite})
	}
}

// @Summary Busca o convite mais recente de uma disciplina
// @Tags invite
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Security BearerAuth
// @Param disciplineId path string true "Discipline ID"
// @Success 200 {object} api.DefaultResponse[Invite]
// @Router /invite/{disciplineId}/current [get]
func (h *handler) GetCurrent() gin.HandlerFunc {
	return func(c *gin.Context) {
		disciplineID := c.Param("disciplineId")
		userID := c.GetString("userID")

		invite, err := h.service.GetCurrent(c.Request.Context(), disciplineID, userID)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}

		c.JSON(200, api.DefaultResponse[*Invite]{Message: "Convite carregado com sucesso", Data: invite})
	}
}

// @Summary Lista convites de uma disciplina
// @Tags invite
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Security BearerAuth
// @Param disciplineId path string true "Discipline ID"
// @Success 200 {object} api.DefaultResponse[[]Invite]
// @Router /invite/{disciplineId} [get]
func (h *handler) ListByDiscipline() gin.HandlerFunc {
	return func(c *gin.Context) {
		disciplineID := c.Param("disciplineId")
		userID := c.GetString("userID")

		invites, err := h.service.ListByDiscipline(c.Request.Context(), disciplineID, userID)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}

		items := make([]Invite, 0, len(invites))
		for _, invite := range invites {
			if invite != nil {
				items = append(items, *invite)
			}
		}

		c.JSON(200, api.DefaultResponse[[]Invite]{Message: "Convites listados com sucesso", Data: items})
	}
}

// @Summary Remove um convite
// @Tags invite
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Security BearerAuth
// @Param inviteId path string true "Invite ID"
// @Success 200 {object} api.MessageResponse
// @Router /invite/{inviteId} [delete]
func (h *handler) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		inviteID := c.Param("inviteId")
		userID := c.GetString("userID")

		err := h.service.Delete(c.Request.Context(), inviteID, userID)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}

		c.JSON(200, api.MessageResponse{Message: "Convite removido com sucesso"})
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

		code := strings.ToUpper(strings.TrimSpace(c.Param("code")))

		err := h.service.SelfRegister(c.Request.Context(), code, input.StudentID, input.Name, input.Phone, input.Email, input.Consent)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}

		c.JSON(200, api.MessageResponse{Message: "Cadastro concluído com sucesso"})
	}
}
