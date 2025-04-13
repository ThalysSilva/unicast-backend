package smtp

import "github.com/gin-gonic/gin"

type handler struct {
	service Service
}

type Handler interface {
	Create() gin.HandlerFunc
	GetInstances() gin.HandlerFunc
}

func NewHandler(service Service) Handler {
	return &handler{
		service: service,
	}
}

func (h *handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for creating an SMTP instance
	}
}

func (h *handler) GetInstances() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for getting SMTP instances
	}
}
