package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/k1ngalph0x/beacon/services/ingestion-service/handler"
	"github.com/k1ngalph0x/beacon/services/ingestion-service/kafka"
)

func main() {

	kafka.InitKafka()

	router := gin.Default()
	router.POST("/events", handler.Ingest)
	router.Run(":8092")

	fmt.Println("Running ingestion-service")

}