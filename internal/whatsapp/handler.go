package whatsapp

import (
	"net/http"

	"github.com/ThalysSilva/unicast-backend/pkg/api"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/gin-gonic/gin"
)

type createInstanceInput struct {
	Phone string `json:"phone" binding:"required"`
}

type CreateInstanceResponse struct {
	InstanceID string `json:"instanceId"`
	QrCode     string `json:"qrCode"`
}

type GetInstancesResponse struct {
	Instances []*Instance `json:"instances"`
	UserEmail string      `json:"userEmail"`
}

type handler struct {
	service Service
}

type Handler interface {
	CreateInstance() gin.HandlerFunc
	GetInstances() gin.HandlerFunc
	DeleteInstance() gin.HandlerFunc
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
// @Success 200 {object} api.DefaultResponse[CreateInstanceResponse]
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
		userID := c.GetString("userID")
		instance, qrCode, err := h.service.CreateInstance(c.Request.Context(), userID, input.Phone)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, api.DefaultResponse[CreateInstanceResponse]{
			Message: "Instância criada com sucesso !.",
			Data: CreateInstanceResponse{
				InstanceID: instance.InstanceName,
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
// @Success 200 {object} api.DefaultResponse[GetInstancesResponse]
// @Failure 400 {object} api.ErrorResponse
// @Router /whatsapp/instance [get]
// @Security Bearer
func (h *handler) GetInstances() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		instances, err := h.service.GetInstances(c.Request.Context(), userID)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}
		if instances == nil {
			instances = []*Instance{}
		}
		response := GetInstancesResponse{Instances: instances, UserEmail: c.GetString("email")}
		c.JSON(http.StatusOK, api.DefaultResponse[GetInstancesResponse]{
			Message: "Instância encontrada com sucesso.",
			Data:    response,
		})
	}
}

// @OperationId deleteInstance
// @Summary Deleta uma instância do WhatsApp
// @Description Remove a instância do usuário e permite criar/parear novamente.
// @Tags whatsapp
// @Accept json
// @Produce json
// @Param id path string true "Instance ID"
// @Success 200 {object} api.DefaultResponse[any]
// @Failure 400 {object} api.ErrorResponse
// @Router /whatsapp/instance/{id} [delete]
// @Security Bearer
func (h *handler) DeleteInstance() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		instanceID := c.Param("id")

		if err := h.service.DeleteInstance(c.Request.Context(), userID, instanceID); err != nil {
			customerror.HandleResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, api.DefaultResponse[any]{
			Message: "Instância deletada com sucesso.",
			Data:    nil,
		})
	}
}
