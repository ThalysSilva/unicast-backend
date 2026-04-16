package smtp

import (
	"net/url"

	"github.com/ThalysSilva/unicast-backend/pkg/api"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

type createInstanceInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	Jwe      string `json:"jwe" binding:"required"`
}

type testConnectionInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
}

type Handler interface {
	Create(jweSecret []byte) gin.HandlerFunc
	StartOAuth() gin.HandlerFunc
	OAuthCallback(provider string) gin.HandlerFunc
	TestConnection() gin.HandlerFunc
	GetInstances() gin.HandlerFunc
	DeleteInstance() gin.HandlerFunc
}

func NewHandler(service Service) Handler {
	return &handler{
		service: service,
	}
}

type oauthStartResponse struct {
	URL string `json:"url"`
}

// @Summary Cria uma instância SMTP
// @Tags smtp
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Security BearerAuth
// @Param body body createInstanceInput true "Dados SMTP"
// @Success 200 {object} api.MessageResponse
// @Router /smtp/instance [post]
func (h *handler) Create(jweSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input createInstanceInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		userID := c.GetString("userID")
		err := h.service.Create(c.Request.Context(), jweSecret, userID, input.Jwe, input.Email, input.Password, input.Host, input.Port)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}
		c.JSON(200, api.MessageResponse{Message: "SMTP instance created successfully"})

	}
}

// @Summary Testa uma conexão SMTP
// @Tags smtp
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Security BearerAuth
// @Param body body createInstanceInput true "Dados SMTP"
// @Success 200 {object} api.MessageResponse
// @Failure 400 {object} api.ErrorResponse
// @Router /smtp/instance/test [post]
func (h *handler) TestConnection() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input testConnectionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		if err := h.service.TestConnection(c.Request.Context(), input.Email, input.Password, input.Host, input.Port); err != nil {
			customerror.HandleResponse(c, err)
			return
		}
		c.JSON(200, api.MessageResponse{Message: "Conexao SMTP validada com sucesso"})
	}
}

func (h *handler) StartOAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		provider := c.Param("provider")
		authURL, err := h.service.StartOAuth(c.Request.Context(), userID, provider)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}
		c.JSON(200, api.DefaultResponse[oauthStartResponse]{
			Message: "URL de autorizacao gerada com sucesso",
			Data:    oauthStartResponse{URL: authURL},
		})
	}
}

func (h *handler) OAuthCallback(provider string) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")
		redirectURL, err := h.service.HandleOAuthCallback(c.Request.Context(), provider, code, state)
		if err != nil {
			msg := url.QueryEscape(err.Error())
			c.Redirect(302, redirectURL+"?oauth_status=error&oauth_provider="+provider+"&oauth_message="+msg)
			return
		}
		c.Redirect(302, redirectURL)
	}
}

// @Summary Lista instâncias SMTP do usuário
// @Tags smtp
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} api.DefaultResponse[[]Instance]
// @Router /smtp/instance [get]
func (h *handler) GetInstances() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		instances, err := h.service.GetInstances(c.Request.Context(), userID)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}
		items := make([]Instance, 0, len(instances))
		for _, inst := range instances {
			if inst != nil {
				items = append(items, *inst)
			}
		}
		c.JSON(200, api.DefaultResponse[[]Instance]{Message: "Instâncias listadas com sucesso", Data: items})
	}
}

// @Summary Remove uma instância SMTP
// @Tags smtp
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Security BearerAuth
// @Param id path string true "Instance ID"
// @Success 200 {object} api.MessageResponse
// @Failure 400 {object} api.ErrorResponse
// @Router /smtp/instance/{id} [delete]
func (h *handler) DeleteInstance() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		instanceID := c.Param("id")
		if err := h.service.DeleteInstance(c.Request.Context(), userID, instanceID); err != nil {
			customerror.HandleResponse(c, err)
			return
		}
		c.JSON(200, api.MessageResponse{Message: "SMTP instance deleted successfully"})
	}
}
