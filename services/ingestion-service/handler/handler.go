package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/k1ngalph0x/beacon/services/ingestion-service/kafka"
)

type Event struct{
	ProjectID  string    `json:"project_id"`
	Timestamp  time.Time `json:"timestamp"`
	Level      string    `json:"level"`
	Message    string    `json:"message"`
	StackTrace string    `json:"stack_trace"`
}

func Ingest(c *gin.Context){
	var event Event

	err := c.ShouldBind(&event)

	println("event")
	println("Project", event.ProjectID)
	println("Level", event.Level)
	println("Message", event.Message)

	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event payload"})
		return
	}

	payload, err := json.Marshal(event)
	if err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return 
	}

	err  = kafka.Publish(event.ProjectID, payload)
	if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish to kafka"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status": "queued",
	})
	
}