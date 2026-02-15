package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

var Writer *kafka.Writer

func InitKafka() {
	Writer = &kafka.Writer{
		Addr: kafka.TCP("localhost:9092"),
		Topic: "beacon-events",
		Balancer: &kafka.LeastBytes{},
	}
}	


func Publish(projectID string, message []byte)error{
	return Writer.WriteMessages(context.Background(),
	kafka.Message{
		//Key: []byte(time.Now().String()),
		Key: []byte(projectID),
		Value: message,
	},
)
}