package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/k1ngalph0x/beacon/services/alert-service/config"
	"github.com/k1ngalph0x/beacon/services/alert-service/db"
	"github.com/k1ngalph0x/beacon/services/alert-service/models"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)


type IssueUpdate struct {
	IssueID   string `json:"issue_id"`
	ProjectID string `json:"project_id"`
	Count     int    `json:"count"`
	Level     string `json:"level"`
	Status    string `json:"status"`
}


func main() {
	_, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	conn, err := db.ConnectDB()

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	err = conn.AutoMigrate(&models.AlertRule{})
	if err != nil {
		log.Fatalf("Failed to migrate alert table: %v", err)
	}

	go startConsumer(conn)

	select {} 
}

func startConsumer(conn *gorm.DB) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "issue-updates",
		GroupID: "alert-consumers",
	})

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		var update IssueUpdate
		if err := json.Unmarshal(msg.Value, &update); err != nil {
			continue
		}

		checkAlerts(conn, update)
	}
}

func checkAlerts(conn *gorm.DB, update IssueUpdate){
	var rules []models.AlertRule


	conn.Where("project_id = ? AND is_active = ?", update.ProjectID, true).
		Find(&rules)

	for _, rule := range rules {

		if update.Level == rule.Level && update.Count >= rule.Threshold {

			log.Printf("ALERT TRIGGERED: Issue %s reached %d",
				update.IssueID, update.Count)

		}
	}

}