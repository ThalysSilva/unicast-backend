package campus

import "github.com/gin-gonic/gin"

type handler struct {
	service Service
}

type createCampusInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type Handler interface {
	Create() gin.HandlerFunc
	GetCampuses() gin.HandlerFunc
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
		c.JSON(200, gin.H{"message": "Campus created successfully"})

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
