package auth

import (
	"net/http"

	"github.com/ThalysSilva/unicast-backend/pkg/api"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
	"github.com/gin-gonic/gin"
)

// RegisterInput Define o input para o registro do usuário
type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

type handler struct {
	service Service
}

type AuthHandler interface {
	Register() gin.HandlerFunc
	Login() gin.HandlerFunc
	Logout() gin.HandlerFunc
	Refresh() gin.HandlerFunc
}

func NewHandler(service Service) AuthHandler {
	return &handler{service: service}
}

// @Summary Registra um novo usuário
// @Description Registra um novo usuário no sistema
// @Tags auth
// @Accept json
// @Produce json
// @OperationId register
// @Param user body RegisterInput true "User data"
// @Success 201
// @Failure 400 {object} api.ErrorResponse
// @Router /auth/register [post]
func (s *handler) Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input RegisterInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		if _, err := s.service.Register(c.Request.Context(), input.Email, input.Password, input.Name); err != nil {
			customerror.HandleResponse(c, err)
			return
		}
		c.JSON(http.StatusCreated, nil)
	}
}

// LoginInput define o input para o login do usuário
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// @Summary Gera o acesso a um usuário
// @Description Gera o acesso a um usuário no sistema
// @Tags auth
// @Accept json
// @Produce json
// @Param user body LoginInput true "User data"
// @OperationId login
// @Success 200 {object} api.DefaultResponse[LoginResponse]
// @Failure 401 {object} api.ErrorResponse
// @Router /auth/login [post]
func (s *handler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input LoginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		loginResponse, err := s.service.Login(c.Request.Context(), input.Email, input.Password)
		if err != nil {
			customerror.HandleResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, api.DefaultResponse[LoginResponse]{
			Message: "Login realizado com sucesso.",
			Data:    *loginResponse,
		})
	}
}

// @Summary Remove o acesso a um usuário
// @Description Remove o acesso a um usuário do sistema
// @Tags auth
// @Accept json
// @OperationId logout
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Security BearerAuth
// @Param user_id path string true "User ID"
// @Success 200 {object} object{message=string}
// @Failure 401 {object} api.ErrorResponse
// @Router /auth/logout [post]
func (s *handler) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, _ := c.Get("user_id")
		if err := s.service.Logout(c.Request.Context(), userId.(string)); err != nil {
			customerror.HandleResponse(c, err)
		}
		c.JSON(http.StatusOK, gin.H{"message": "Usuário deslogado com sucesso."})
	}
}

// RefreshInput Define o input para o refresh do token
type RefreshInput struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// @Summary Atualiza o Refresh Token do usuário
// @Description Atualiza o Refresh Token do usuário no sistema
// @Tags auth
// @Accept json
// @OperationId refreshToken
// @Produce json
// @Param refreshToken body RefreshInput true "Refresh token"
// @Success 200 {object} api.DefaultResponse[RefreshResponse]
// @Failure 401 {object} api.ErrorResponse
// @Router /auth/refresh [post]
func (s *handler) Refresh() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input RefreshInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		response, err := s.service.RefreshToken(c.Request.Context(), input.RefreshToken)
		if err != nil {
			customerror.HandleResponse(c, err)
		}
		c.JSON(http.StatusOK, response)
	}
}
