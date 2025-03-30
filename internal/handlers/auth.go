package handlers

import (
	"net/http"
	"unicast-api/internal/models"
	"unicast-api/internal/services"
	"unicast-api/pkg/utils"

	"github.com/gin-gonic/gin"
)

// RegisterInput Define o input para o registro do usuário
type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

// @Summary Registra um novo usuário
// @Description Registra um novo usuário no sistema
// @Tags auth
// @Accept json
// @Produce json
// @OperationId register
// @Param user body RegisterInput true "User data"
// @Success 201
// @Failure 400 {object} models.ErrorResponse
// @Router /auth/register [post]
func Register(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input RegisterInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		if _, err := authService.Register(input.Email, input.Password, input.Name); err != nil {
			utils.HandleErrorResponse(c, err)
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
// @Success 200 {object} models.DefaultResponse[services.LoginResponse]
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/login [post]
func Login(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input LoginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		loginResponse, err := authService.Login(input.Email, input.Password)
		if err != nil {
			utils.HandleErrorResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, models.DefaultResponse[services.LoginResponse]{
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
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/logout [post]
func Logout(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, _ := c.Get("user_id")
		if err := authService.Logout(userId.(string)); err != nil {
			utils.HandleErrorResponse(c, err)
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
// @Success 200 {object} services.RefreshResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/refresh [post]
func Refresh(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input RefreshInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		response, err := authService.RefreshToken(input.RefreshToken)
		if err != nil {
			utils.HandleErrorResponse(c, err)
		}
		c.JSON(http.StatusOK, response)
	}
}
