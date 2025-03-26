package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"todo-list-api/internal/services"
	_"todo-list-api/internal/models"
)

// RegisterInput defines the input for user registration
type RegisterInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

// @Summary Register a new user
// @Description Register a new user with username and password
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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if _, err := authService.Register(input.Email, input.Password, input.Name); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error: ": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, nil)
	}
}

// LoginInput defines the input for user login
type LoginInput struct {
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// @Summary Login a user
// @Description Login a user and return access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param user body LoginInput true "User data"
// @OperationId login
// @Success 200 {object} services.LoginResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/login [post]
func Login(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input LoginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		loginResponse, err := authService.Login(input.Email, input.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, loginResponse)
	}
}

// @Summary Logout a user
// @Description Logout a user and invalidate refresh token
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Usuário deslogado com sucesso."})
	}
}

// RefreshInput defines the input for token refresh
type RefreshInput struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @OperationId refreshToken
// @Produce json
// @Param refresh_token body RefreshInput true "Refresh token"
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}
