package program

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

type createCourseInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Active      bool   `json:"active" binding:"required"`
	CampusID    string `json:"campus_id" binding:"required"`
}

type updateCourseInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

type Handler interface {
	Create() gin.HandlerFunc
	GetProgramsByCampusID() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
}

func NewHandler(service Service) Handler {
	return &handler{
		service: service,
	}
}

func (h *handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input createCourseInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}

		err := h.service.Create(c.Request.Context(), input.CampusID, input.Name, input.Description, input.Active)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, gin.H{"message": "Curso criado com sucesso"})

	}
}

func (h *handler) GetProgramsByCampusID() gin.HandlerFunc {
	return func(c *gin.Context) {
		campusID := c.Param("id")
		instances, err := h.service.GetProgramsByCampusID(c.Request.Context(), campusID)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, instances)
	}
}

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
		if input.Active {
			fields["active"] = input.Active
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
		c.JSON(200, gin.H{"message": "Curso Atualizado com sucesso"})
	}
}

func (h *handler) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		programID := c.Param("id")

		isOwner, err := h.service.isOwner(c.Request.Context(), programID, userID)
		if err != nil {
			c.Error(err)
			return
		}
		if !isOwner {
			c.Error(errors.New("você não tem permissão para deletar este Curso"))
			return
		}

		err = h.service.Delete(c.Request.Context(), programID)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, gin.H{"message": "Curso Deletado com sucesso"})
	}
}
