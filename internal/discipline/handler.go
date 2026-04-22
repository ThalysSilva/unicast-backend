package discipline

import (
	"errors"
	"net/http"

	"github.com/ThalysSilva/unicast-backend/pkg/api"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

type createDisciplineInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	ProgramID   string `json:"program_id" binding:"required"`
	Year        int    `json:"year" binding:"required"`
	Semester    int    `json:"semester" binding:"required"`
}

type updateDisciplineInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Year        int    `json:"year"`
	Semester    int    `json:"semester"`
}

type Handler interface {
	Create() gin.HandlerFunc
	GetDisciplines() gin.HandlerFunc
	GetDisciplinesByProgramID() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
}

func NewHandler(service Service) Handler {
	return &handler{
		service: service,
	}
}

// @Summary Cria uma disciplina
// @Tags discipline
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body createDisciplineInput true "Dados da disciplina"
// @Success 200 {object} api.MessageResponse
// @Router /discipline [post]
func (h *handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		var input createDisciplineInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		err := h.service.Create(c.Request.Context(), userID, input.ProgramID, input.Name, input.Description, input.Year, input.Semester)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, api.MessageResponse{Message: "Disciplina criada com sucesso"})

	}
}

// @Summary Lista todas as disciplinas do usuário
// @Tags discipline
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} api.DefaultResponse[[]Discipline]
// @Router /discipline [get]
func (h *handler) GetDisciplines() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		instances, err := h.service.GetDisciplinesByUserID(c.Request.Context(), userID)
		if err != nil {
			c.Error(err)
			return
		}
		items := make([]Discipline, 0, len(instances))
		for _, discipline := range instances {
			if discipline != nil {
				items = append(items, *discipline)
			}
		}
		c.JSON(200, api.DefaultResponse[[]Discipline]{Message: "Disciplinas listadas com sucesso", Data: items})
	}
}

// @Summary Lista disciplinas por curso
// @Tags discipline
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param programId path string true "Program ID"
// @Success 200 {object} api.DefaultResponse[[]Discipline]
// @Router /discipline/{programId} [get]
func (h *handler) GetDisciplinesByProgramID() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		programID := c.Param("programId")
		instances, err := h.service.GetDisciplinesByProgramID(c.Request.Context(), userID, programID)
		if err != nil {
			c.Error(err)
			return
		}
		items := make([]Discipline, 0, len(instances))
		for _, discipline := range instances {
			if discipline != nil {
				items = append(items, *discipline)
			}
		}
		c.JSON(200, api.DefaultResponse[[]Discipline]{Message: "Disciplinas listadas com sucesso", Data: items})
	}
}

// @Summary Atualiza uma disciplina
// @Tags discipline
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Discipline ID"
// @Param body body updateDisciplineInput true "Campos para atualizar"
// @Success 200 {object} api.MessageResponse
// @Router /discipline/{id} [put]
func (h *handler) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input updateDisciplineInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		userID := c.GetString("userID")
		disciplineID := c.Param("id")

		isOwner, err := h.service.isOwner(c.Request.Context(), disciplineID, userID)
		if err != nil {
			c.Error(err)
			return
		}
		if !isOwner {
			c.JSON(http.StatusForbidden, api.ErrorResponse{Error: "você não tem permissão para atualizar esta disciplina"})
			return
		}

		fields := make(map[string]any)

		if input.Name != "" {
			fields["name"] = input.Name
		}
		if input.Description != "" {
			fields["description"] = input.Description
		}
		if input.Year != 0 {
			fields["year"] = input.Year
		}
		if input.Semester != 0 {
			fields["semester"] = input.Semester
		}

		if len(fields) == 0 {
			c.Error(errors.New("nenhum campo para atualizar"))
			return
		}

		err = h.service.Update(c.Request.Context(), disciplineID, fields)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, api.MessageResponse{Message: "Disciplina atualizada com sucesso"})
	}
}

// @Summary Deleta uma disciplina
// @Tags discipline
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Discipline ID"
// @Success 200 {object} api.MessageResponse
// @Router /discipline/{id} [delete]
func (h *handler) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		disciplineID := c.Param("id")

		isOwner, err := h.service.isOwner(c.Request.Context(), disciplineID, userID)
		if err != nil {
			c.Error(err)
			return
		}
		if !isOwner {
			c.JSON(http.StatusForbidden, api.ErrorResponse{Error: "você não tem permissão para deletar esta disciplina"})
			return
		}

		err = h.service.Delete(c.Request.Context(), disciplineID)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, api.MessageResponse{Message: "Disciplina deletada com sucesso"})
	}
}
