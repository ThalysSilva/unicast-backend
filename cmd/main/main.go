package main

import (
	"encoding/hex"
	"log"
	"os"

	_ "github.com/ThalysSilva/unicast-backend/docs"
	"github.com/ThalysSilva/unicast-backend/internal/auth"
	"github.com/ThalysSilva/unicast-backend/internal/campus"
	"github.com/ThalysSilva/unicast-backend/internal/config"
	"github.com/ThalysSilva/unicast-backend/internal/course"
	"github.com/ThalysSilva/unicast-backend/internal/middleware"
	"github.com/ThalysSilva/unicast-backend/internal/program"
	"github.com/ThalysSilva/unicast-backend/internal/repository"
	"github.com/ThalysSilva/unicast-backend/internal/smtp"
	"github.com/ThalysSilva/unicast-backend/internal/whatsapp"
	"github.com/ThalysSilva/unicast-backend/pkg/database"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

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
	smtpService := smtp.NewService(repos.SmtpInstance)
	campusService := campus.NewService(repos.Campus)
	courseService := course.NewService(repos.Course)
	programService := program.NewService(repos.Program)

	// Handlers
	authHandler := auth.NewHandler(authService)
	whatsappHandler := whatsapp.NewHandler(whatsappService)
	smtpHandler := smtp.NewHandler(smtpService)
	campusHandler := campus.NewHandler(campusService)
	courseHandler := course.NewHandler(courseService)
	programHandler := program.NewHandler(programService)

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
	// Rotas de campus
	campusGroup := r.Group("/campus")
	{
		campusGroup.Use(middleware.UseAuthentication(secrets.AccessToken))
		campusGroup.POST("", campusHandler.Create())
		campusGroup.GET("", campusHandler.GetCampuses())
		campusGroup.PUT(":id", campusHandler.Update())
	}

	// Rotas de disciplinas
	courseGroup := r.Group("/course")
	{
		courseGroup.Use(middleware.UseAuthentication(secrets.AccessToken))
		courseGroup.POST("", courseHandler.Create())
		courseGroup.GET("/:programId", courseHandler.GetCoursesByProgramID())
		courseGroup.PUT("/:id", courseHandler.Update())
		courseGroup.DELETE("/:id", courseHandler.Delete())
	}

	// Rotas de cursos
	programGroup := r.Group("/program")
	{
		programGroup.Use(middleware.UseAuthentication(secrets.AccessToken))
		programGroup.POST("", programHandler.Create())
		programGroup.GET("/:campusID", programHandler.GetProgramsByCampusID())
		programGroup.PUT("/:id", programHandler.Update())
		programGroup.DELETE("/:id", programHandler.Delete())
	}

	// Rotas do WhatsApp
	whatsappGroup := r.Group("/whatsapp")
	{
		whatsappGroup.Use(middleware.UseAuthentication(secrets.AccessToken))
		whatsappGroup.POST("/instance", whatsappHandler.CreateInstance())
		whatsappGroup.GET("/instance", whatsappHandler.GetInstances())
	}

	smtpGroup := r.Group("/smtp")
	{
		smtpGroup.Use(middleware.UseAuthentication(secrets.AccessToken))
		smtpGroup.POST("/instance", smtpHandler.Create(secrets.Jwe))
		smtpGroup.GET("/instance", smtpHandler.GetInstances())
	}

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Inicia o servidor
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
