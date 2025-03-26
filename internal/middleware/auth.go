package middleware

import (
	"net/http"
	"strings"
	"unicast-api/pkg/auth"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(accessTokenSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Cabeçalho de autorização nao encontrado"})
			c.Abort()
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Cabeçalho de autorização inválido"})
			c.Abort()
			return
		}
		claims, err := auth.ValidateToken(parts[1], accessTokenSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
