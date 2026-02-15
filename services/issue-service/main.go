package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/k1ngalph0x/beacon/services/issue-service/config"
	"github.com/k1ngalph0x/beacon/services/issue-service/db"
	"github.com/k1ngalph0x/beacon/services/issue-service/handler"
	"github.com/k1ngalph0x/beacon/services/issue-service/middleware"
	"github.com/k1ngalph0x/beacon/services/issue-service/models"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)


func main() {

	config, err := config.LoadConfig()
	if err != nil{
		log.Fatalf("Error loading config: %v", err)
	}
	
	conn, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}

	err = conn.AutoMigrate(&models.Issue{})
	if err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	authMiddleware := middleware.NewAuthMiddleware(config.TOKEN.JwtKey)

	go startKafkaConsumer(conn)
	startHTTPServer(conn, authMiddleware)
}


func startKafkaConsumer(conn *gorm.DB) {

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "beacon-events",
		GroupID: "issue-consumers",
	})

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Println("Kafka read error:", err)
			continue
		}

		var event models.Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Println("Invalid event:", err)
			continue
		}

		handler.ProcessEvent(conn, event)
	}

}

func startHTTPServer(db *gorm.DB, authMiddleware  *middleware.AuthMiddleware) {
	router := gin.Default()

	router.Use(authMiddleware.RequireAuth())
	router.GET("/projects/:project_id/issues", handler.GetProjectIssue(db))
	router.GET("/issues/:id", handler.GetIssue(db))
	router.PATCH("/issues/:id/resolve", handler.ResolveIssue(db))

	log.Println("Issue API running on :8094")
	router.Run(":8094")
}
