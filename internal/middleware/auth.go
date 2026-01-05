package middleware

import (
	"github.com/ThalysSilva/unicast-backend/internal/auth"
	"github.com/ThalysSilva/unicast-backend/pkg/api"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func UseAuthentication(accessTokenSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "Cabeçalho de autorização nao encontrado"})
			c.Abort()
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "Cabeçalho de autorização inválido"})
			c.Abort()
			return
		}
		claims, err := auth.ValidateToken(parts[1], accessTokenSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "Token inválido"})
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
