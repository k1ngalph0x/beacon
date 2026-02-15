package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/k1ngalph0x/beacon/services/query-service/config"
	"github.com/k1ngalph0x/beacon/services/query-service/db"
	"github.com/k1ngalph0x/beacon/services/query-service/handler"
)

func main() {
	_, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	conn, err := db.ConnectDB()

	router := gin.Default()

	router.GET("/projects/:id/events", handler.GetEvents(conn))
	router.GET("/projects/:id/errors/count", handler.GetErrorCount(conn))
	router.GET("/projects/:id/error-rate", handler.GetErrorRate(conn))


	router.Run(":8093")
	
}