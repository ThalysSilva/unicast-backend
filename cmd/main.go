package main

import (
	"log"
	"os"
	_ "todo-list-api/docs"
	"todo-list-api/internal/handlers"
	"todo-list-api/internal/models"
	"todo-list-api/internal/repositories"
	"todo-list-api/internal/services"
	"todo-list-api/pkg/database"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	if err := database.InitDB(); err != nil {
		log.Fatal(err)
	}

	// Secrets
	secrets := &models.Secrets{
		AccessToken:  []byte(os.Getenv("ACCESS_TOKEN_SECRET")),
		RefreshToken: []byte(os.Getenv("REFRESH_TOKEN_SECRET")),
		Jwe:          []byte(os.Getenv("JWE_SECRET")),
	}

	port := os.Getenv("API_PORT")

	// Repositórios
	userRepo := repositories.NewUserRepository(database.DB)

	authService := services.NewAuthService(userRepo, secrets)

	r := gin.Default()

	// Rotas de autenticação
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", handlers.Register(authService))
		authGroup.POST("/login", handlers.Login(authService))
		authGroup.POST("/refresh", handlers.Refresh(authService))
	}

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Inicia o servidor
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
