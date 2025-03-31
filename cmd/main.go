package main

import (
	"encoding/hex"
	"log"
	"os"

	_ "github.com/ThalysSilva/unicast-backend/docs"
	"github.com/ThalysSilva/unicast-backend/internal/handlers"
	"github.com/ThalysSilva/unicast-backend/internal/middleware"
	"github.com/ThalysSilva/unicast-backend/internal/models"
	"github.com/ThalysSilva/unicast-backend/internal/services"
	"github.com/ThalysSilva/unicast-backend/pkg/database"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	if err := database.InitDB(); err != nil {
		log.Fatal(err)
	}

	// Secrets
	jweSecretHex := os.Getenv("JWE_SECRET")
	jweSecret, err := hex.DecodeString(jweSecretHex)
	if err != nil {
		log.Fatalf("Erro ao decodificar JWE_SECRET: %v", err)
	}
	if len(jweSecret) != 32 {
		log.Fatalf("JWE_SECRET tem tamanho inválido: %d bytes, esperado 32", len(jweSecret))
	}

	secrets := &models.Secrets{
		AccessToken:  []byte(os.Getenv("ACCESS_TOKEN_SECRET")),
		RefreshToken: []byte(os.Getenv("REFRESH_TOKEN_SECRET")),
		Jwe:          jweSecret,
	}

	port := os.Getenv("API_PORT")

	// Repositórios
	userRepo := repositories.NewUserRepository(database.DB)

	// Serviços
	authService := services.NewAuthService(userRepo, secrets)

	r := gin.Default()

	r.Use(middleware.ValidationErrorHandler())

	// Rotas de autenticação
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", handlers.Register(authService))
		authGroup.POST("/login", handlers.Login(authService))
		authGroup.POST("/refresh", handlers.Refresh(authService))
		// Com autenticação
		authGroup.Use(middleware.UseAuthentication(secrets.AccessToken))
		authGroup.POST("/logout", handlers.Logout(authService))
	}

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Inicia o servidor
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
