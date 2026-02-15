package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/k1ngalph0x/beacon/services/auth-service/api"
	"github.com/k1ngalph0x/beacon/services/auth-service/config"
	"github.com/k1ngalph0x/beacon/services/auth-service/db"
	"github.com/k1ngalph0x/beacon/services/auth-service/middleware"
	"github.com/k1ngalph0x/beacon/services/auth-service/models"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil{
		log.Fatalf("Error loading config: %v", err)
	}
	conn, err := db.ConnectDB()

	if err != nil{
		log.Fatalf("Error connecting to database: %v", err)
	}

	err = conn.AutoMigrate(&models.User{})
	if err != nil{
		log.Fatalf("Failed to migrate User table: %v", err)
	}

	err = conn.AutoMigrate(&models.RefreshToken{})
	if err != nil{
		log.Fatalf("Failed to migrate RefreshToken table: %v", err)
	}

	err = conn.AutoMigrate(&models.Project{})
	if err != nil {
		log.Fatalf("Failed to migrate project table: %v", err)
	}

	handler := api.NewHandler(conn, config)
	authMiddleware := middleware.NewAuthMiddleware(config.TOKEN.JwtKey)

	router := gin.Default()
	router.Use(gin.Logger())

	auth := router.Group("/auth")
	{
		auth.POST("/signup", handler.SignUp)
		auth.POST("/signin", handler.SignIn)
		auth.POST("/refresh", handler.Refresh)
	}

	router.Use(authMiddleware.RequireAuth())
	user := router.Group("/user")
	{
		//user.POST("/onboard", handler.Onboard)
		user.POST("/project", handler.CreateProject)
	}

	router.Run(":8080")
}