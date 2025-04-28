package course

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
	GetCourses() gin.HandlerFunc
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
		userID := c.GetString("userID")
		err := h.service.Create(c.Request.Context(), userID, input.Name, input.Description, input.Year, input.Semester)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, gin.H{"message": "Campus criado com sucesso"})

	}
}

func (h *handler) GetCourses() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		instances, err := h.service.GetCourses(c.Request.Context(), userID)
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
		c.JSON(200, gin.H{"message": "Campus Atualizado com sucesso"})
	}
}

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
		c.JSON(200, gin.H{"message": "Campus Deletado com sucesso"})
	}
}
