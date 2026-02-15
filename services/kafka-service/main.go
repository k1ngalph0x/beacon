package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/k1ngalph0x/beacon/services/kafka-service/config"
	"github.com/k1ngalph0x/beacon/services/kafka-service/db"
	"github.com/k1ngalph0x/beacon/services/kafka-service/models"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type Event struct {
	ProjectID  string    `json:"project_id"`
	Timestamp  time.Time `json:"timestamp"`
	Level      string    `json:"level"`
	Message    string    `json:"message"`
	StackTrace *string    `json:"stack_trace,omitempty"`
}


func main() {
	
	config, err := config.LoadConfig()

	if err != nil{
		log.Fatalf("Error loading config: %v", err)
	}
	conn, err := db.ConnectDB()

	if err != nil{
		log.Fatalf("Error connecting to database: %v", err)
	}

	err = conn.AutoMigrate(&models.Events{})
	if err != nil{
		log.Fatalf("Failed to migrate User table: %v", err)
	}


	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.KAFKA.Brokers,
		Topic:   "beacon-events",
		GroupID: "beacon-consumers",
		MinBytes: 1,
		MaxBytes: 10e6,

	})

	for {
		//msg, err := reader.ReadMessage(context.Background())
		msg, err := reader.FetchMessage(context.Background())
		if err != nil{
			fmt.Println("Error reading message:", err)
			continue
		}
		
		var event Event

		err  = json.Unmarshal(msg.Value, &event)
		if err != nil{
			fmt.Println("Not a valid event:", err)
			continue
		}

		err = insertEvent(conn, event, msg.Partition, msg.Offset)

		if err != nil{
			fmt.Println("Failed to insert:", err)
			continue
		}

		err = reader.CommitMessages(context.Background(), msg)
		if err != nil{
			fmt.Println("Failed to commit:", err)
		}

		fmt.Println("Successfully Inserted")
	}
}



func insertEvent(db *gorm.DB, e Event, partition int, offset int64) error {
	event := models.Events{
		ProjectID:      e.ProjectID,
		Level:          e.Level,
		Message:        e.Message,
		StackTrace:     e.StackTrace,  
		EventTimestamp: e.Timestamp,
		KafkaPartition: &partition,     
		KafkaOffset:    &offset,  
	}

	result := db.Create(&event)
	if result.Error != nil{
		return result.Error
	}

	fmt.Println("Successfully inserted to db")

	return nil
}