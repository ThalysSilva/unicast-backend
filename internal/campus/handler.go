package campus

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

type createCampusInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type updateCampusInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Handler interface {
	Create() gin.HandlerFunc
	GetCampuses() gin.HandlerFunc
	Update() gin.HandlerFunc
}

func NewHandler(service Service) Handler {
	return &handler{
		service: service,
	}
}

func (h *handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input createCampusInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		userID := c.GetString("userID")
		err := h.service.Create(c.Request.Context(), userID, input.Name, input.Description)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, gin.H{"message": "Campus criado com sucesso"})

	}
}

func (h *handler) GetCampuses() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		instances, err := h.service.GetCampuses(c.Request.Context(), userID)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, instances)
	}
}

func (h *handler) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input updateCampusInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		userID := c.GetString("userID")
		campusId := c.Param("id")
		campusSelected, err := h.service.GetCampus(campusId)

		if err != nil {
			c.Error(err)
			return
		}

		if campusSelected == nil {
			c.Error(errors.New("Campus não encontrado"))
			return
		}

		if campusSelected.UserOwnerID != userID {
			c.Error(errors.New("você não tem permissão para atualizar este campus"))
			return
		}

		fields := make(map[string]any)

		if input.Name != "" {
			fields["name"] = input.Name
		}
		if input.Description != "" {
			fields["description"] = input.Description
		}

		if len(fields) == 0 {
			c.Error(errors.New("nenhum campo para atualizar"))
			return
		}

		err = h.service.Update(c.Request.Context(), campusId, fields)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, gin.H{"message": "Campus Atualizado com sucesso"})
	}
}
