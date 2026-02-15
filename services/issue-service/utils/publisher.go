package publisher

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

func PublishEvent(writer *kafka.Writer, key string, payload interface{}) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Println("Failed to marshal event:", err)
		return
	}

	err = writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(key),
			Value: bytes,
		},
	)

	if err != nil {
		log.Println("Failed to publish event:", err)
	}
}
