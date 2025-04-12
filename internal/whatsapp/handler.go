package whatsapp

import (
	"net/http"

	"github.com/ThalysSilva/unicast-backend/pkg/api"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/gin-gonic/gin"
)

type createInstanceInput struct {
	Phone  string `json:"phone" binding:"required"`
	UserID string `json:"userId" binding:"required"`
}

type createInstanceResponse struct {
	InstanceID string `json:"instanceId"`
	QrCode     string `json:"qrCode"`
}

type handler struct {
	service Service
}

type Handler interface {
	CreateInstance() gin.HandlerFunc
	GetInstances() gin.HandlerFunc
}

func NewHandler(service Service) Handler {
	return &handler{
		service: service,
	}
}

// @OperationId createInstance
// @Summary Cria uma nova instância do WhatsApp
// @Description Cria uma nova instância do WhatsApp para o usuário
// @Tags whatsapp
// @Accept json
// @Produce json
// @Param user body createInstanceInput true "User data"
// @Success 200 {object} api.DefaultResponse[createInstanceResponse]
// @Failure 400 {object} api.ErrorResponse
// @Router /whatsapp/instance [post]
// @Security Bearer
func (h *handler) CreateInstance() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input createInstanceInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		instance, qrCode, err := h.service.CreateInstance(input.UserID, input.Phone)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, api.DefaultResponse[createInstanceResponse]{
			Message: "Instância criada com sucesso !.",
			Data: createInstanceResponse{
				InstanceID: instance.InstanceID,
				QrCode:     qrCode,
			},
		})
	}
}


// @OperationId getInstances
// @Summary Busca todas as instâncias do WhatsApp
// @Description Busca todas as instâncias do WhatsApp para o usuário
// @Tags whatsapp
// @Accept json
// @Produce json
// @Success 200 {object} api.DefaultResponse[[]*Instance]
// @Failure 400 {object} api.ErrorResponse
// @Router /whatsapp/instances [get]
// @Security Bearer
func (h *handler) GetInstances() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userId")
		instance, err := h.service.GetInstances(userID)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}
		if instance == nil {
			c.JSON(http.StatusOK, api.DefaultResponse[[]*Instance]{
				Message: "Instância encontrada com sucesso.",
				Data:    []*Instance{},
			})
		}
		c.JSON(http.StatusOK, api.DefaultResponse[[]*Instance]{
			Message: "Instância encontrada com sucesso.",
			Data:    instance,
		})
	}
}
