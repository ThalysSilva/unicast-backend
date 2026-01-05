package course

import (
	"errors"

	"github.com/ThalysSilva/unicast-backend/pkg/api"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

type createCourseInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Year        int    `json:"year" binding:"required"`
	Semester    int    `json:"semester" binding:"required"`
}

type updateCourseInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Year        int    `json:"year"`
	Semester    int    `json:"semester"`
}

type Handler interface {
	Create() gin.HandlerFunc
	GetCoursesByProgramID() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
}

func NewHandler(service Service) Handler {
	return &handler{
		service: service,
	}
}

// @Summary Cria uma disciplina
// @Tags course
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body createCourseInput true "Dados da disciplina"
// @Success 200 {object} api.DefaultResponse[map[string]string]
// @Router /course [post]
func (h *handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input createCourseInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		userID := c.GetString("userID")
		err := h.service.Create(c.Request.Context(), userID, input.Name, input.Description, input.Year, input.Semester)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, api.DefaultResponse[map[string]string]{Message: "Disciplina criada com sucesso", Data: map[string]string{}})

	}
}

// @Summary Lista disciplinas do usuário
// @Tags course
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} api.DefaultResponse[[]Course]
// @Router /course/{programId} [get]
func (h *handler) GetCoursesByProgramID() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		instances, err := h.service.GetCoursesByProgramID(c.Request.Context(), userID)
		if err != nil {
			c.Error(err)
			return
		}
		items := make([]Course, 0, len(instances))
		for _, course := range instances {
			if course != nil {
				items = append(items, *course)
			}
		}
		c.JSON(200, api.DefaultResponse[[]Course]{Message: "Disciplinas listadas com sucesso", Data: items})
	}
}

// @Summary Atualiza uma disciplina
// @Tags course
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Course ID"
// @Param body body updateCourseInput true "Campos para atualizar"
// @Success 200 {object} api.DefaultResponse[map[string]string]
// @Router /course/{id} [put]
func (h *handler) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input updateCourseInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		userID := c.GetString("userID")
		courseID := c.Param("id")

		isOwner, err := h.service.isOwner(c.Request.Context(), courseID, userID)
		if err != nil {
			c.Error(err)
			return
		}
		if !isOwner {
			c.Error(errors.New("você não tem permissão para atualizar este curso"))
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

		err = h.service.Update(c.Request.Context(), courseID, fields)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, api.DefaultResponse[map[string]string]{Message: "Disciplina atualizada com sucesso", Data: map[string]string{}})
	}
}

// @Summary Deleta uma disciplina
// @Tags course
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Course ID"
// @Success 200 {object} api.DefaultResponse[map[string]string]
// @Router /course/{id} [delete]
func (h *handler) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		courseID := c.Param("id")

		isOwner, err := h.service.isOwner(c.Request.Context(), courseID, userID)
		if err != nil {
			c.Error(err)
			return
		}
		if !isOwner {
			c.Error(errors.New("você não tem permissão para deletar este curso"))
			return
		}

		err = h.service.Delete(c.Request.Context(), courseID)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, api.DefaultResponse[map[string]string]{Message: "Disciplina deletada com sucesso", Data: map[string]string{}})
	}
}
