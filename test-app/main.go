package main

import (
	"time"

	beacon "github.com/k1ngalph0x/beacon/sdk"
)

func main() {
	client := beacon.Init(beacon.Config{
		ProjectID:   "your-project-id",
		APIKey:      "sk_live_your_secret_key",
		IngestURL:   "http://localhost:8092/events",
		Environment: "production",
		Release:     "v1.0.0",
	})

	
	event := &beacon.Event{
		Timestamp:   time.Now(),
		Level:       "error",
		Message:     "Database connection failed",
		StackTrace:  "at connectDB (db.go:42)\nat main (main.go:15)",
		Environment: "production",
		Release:     "v1.0.0",
		Tags: map[string]string{
			"database": "postgres",
			"host":     "localhost:5432",
		},
	}

	client.Queue <- event


	client.Queue <- &beacon.Event{
		Timestamp: time.Now(),
		Level:     "warning",
		Message:   "High memory usage detected",
	}

	time.Sleep(2 * time.Second)
}