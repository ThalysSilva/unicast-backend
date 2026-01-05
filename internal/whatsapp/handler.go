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
	InstanceID  string `json:"instanceId"`
	QrCode      string `json:"qrCode"`
	PairingCode string `json:"pairingCode,omitempty"`
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
	ConnectInstance() gin.HandlerFunc
	ConnectionState() gin.HandlerFunc
	LogoutInstance() gin.HandlerFunc
	RestartInstance() gin.HandlerFunc
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
		instance, connectResp, err := h.service.CreateInstance(c.Request.Context(), userID, input.Phone)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, api.DefaultResponse[CreateInstanceResponse]{
			Message: "Instância criada com sucesso !.",
			Data: CreateInstanceResponse{
				InstanceID:  instance.InstanceName,
				QrCode:      connectResp.Qrcode.Code,
				PairingCode: connectResp.PairingCode,
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

// @OperationId connectInstance
// @Summary Conecta/pareia uma instância na Evolution
// @Description Dispara a conexão usando number (telefone com DDI/DD).
// @Tags whatsapp
// @Accept json
// @Produce json
// @Param id path string true "Instance ID"
// @Success 200 {object} api.DefaultResponse[any]
// @Failure 400 {object} api.ErrorResponse
// @Router /whatsapp/instance/{id}/connect [post]
// @Security Bearer
func (h *handler) ConnectInstance() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		instanceID := c.Param("id")

		resp, err := h.service.ConnectInstance(c.Request.Context(), userID, instanceID)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, api.DefaultResponse[connectResponse]{
			Message: "Instância conectando...",
			Data:    *resp,
		})
	}
}

// @OperationId connectionState
// @Summary Consulta status de conexão da instância
// @Tags whatsapp
// @Produce json
// @Param id path string true "Instance ID"
// @Success 200 {object} api.DefaultResponse[map[string]string]
// @Failure 400 {object} api.ErrorResponse
// @Router /whatsapp/instance/{id}/status [get]
// @Security Bearer
func (h *handler) ConnectionState() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		instanceID := c.Param("id")

		state, err := h.service.ConnectionState(c.Request.Context(), userID, instanceID)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, api.DefaultResponse[map[string]string]{
			Message: "Status consultado com sucesso.",
			Data:    map[string]string{"status": state},
		})
	}
}

// @OperationId logoutInstance
// @Summary Desconecta/logout de uma instância na Evolution
// @Tags whatsapp
// @Produce json
// @Param id path string true "Instance ID"
// @Success 200 {object} api.DefaultResponse[any]
// @Failure 400 {object} api.ErrorResponse
// @Router /whatsapp/instance/{id}/logout [delete]
// @Security Bearer
func (h *handler) LogoutInstance() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		instanceID := c.Param("id")

		if err := h.service.LogoutInstance(c.Request.Context(), userID, instanceID); err != nil {
			customerror.HandleResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, api.DefaultResponse[any]{Message: "Instância desconectada com sucesso."})
	}
}

// @OperationId restartInstance
// @Summary Reinicia uma instância na Evolution
// @Tags whatsapp
// @Produce json
// @Param id path string true "Instance ID"
// @Success 200 {object} api.DefaultResponse[any]
// @Failure 400 {object} api.ErrorResponse
// @Router /whatsapp/instance/{id}/restart [post]
// @Security Bearer
func (h *handler) RestartInstance() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		instanceID := c.Param("id")

		if err := h.service.RestartInstance(c.Request.Context(), userID, instanceID); err != nil {
			customerror.HandleResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, api.DefaultResponse[any]{Message: "Instância reiniciada com sucesso."})
	}
}
