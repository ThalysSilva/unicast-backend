package main

import (
	"database/sql"
	"encoding/hex"
	"log"
	"os"

	_ "github.com/ThalysSilva/unicast-backend/docs"
	"github.com/ThalysSilva/unicast-backend/internal/auth"
	"github.com/ThalysSilva/unicast-backend/internal/config"
	"github.com/ThalysSilva/unicast-backend/internal/middleware"
	"github.com/ThalysSilva/unicast-backend/internal/repository"
	"github.com/ThalysSilva/unicast-backend/internal/whatsapp"
	"github.com/ThalysSilva/unicast-backend/pkg/database"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)
var db *sql.DB
func main() {
	db, err := database.InitDB()
	if err != nil {
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

	secrets := &config.Secrets{
		AccessToken:  []byte(os.Getenv("ACCESS_TOKEN_SECRET")),
		RefreshToken: []byte(os.Getenv("REFRESH_TOKEN_SECRET")),
		Jwe:          jweSecret,
	}

	port := os.Getenv("API_PORT")

	// Repositórios
	repos := repository.NewRepositories(db)

	// Serviços
	authService := auth.NewService(repos.User, secrets)
	whatsappService := whatsapp.NewService(repos.WhatsAppInstance, repos.User)

	// Handlers
	authHandler := auth.NewHandler(authService)
	whatsappHandler := whatsapp.NewHandler(whatsappService)

	r := gin.Default()

	r.Use(middleware.ValidationErrorHandler())

	// Rotas de autenticação
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register())
		authGroup.POST("/login", authHandler.Login())
		authGroup.POST("/refresh", authHandler.Refresh())
		// Com autenticação
		authGroup.Use(middleware.UseAuthentication(secrets.AccessToken))
		authGroup.POST("/logout", authHandler.Logout())
	}

	// Rotas do WhatsApp
	whatsappGroup := r.Group("/whatsapp")
	{
		whatsappGroup.Use(middleware.UseAuthentication(secrets.AccessToken))
		whatsappGroup.POST("/instance", whatsappHandler.CreateInstance())
		whatsappGroup.GET("/instance", whatsappHandler.GetInstances())
	}

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Inicia o servidor
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
