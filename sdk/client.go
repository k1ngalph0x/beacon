package beacon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type Config struct {
	ProjectID string
	APIKey    string
	IngestURL string
	Environment string
	Release string
}

type Event struct {
	Timestamp   time.Time         `json:"timestamp"`
	Level       string            `json:"level"`
	Message     string            `json:"message"`
	StackTrace  string            `json:"stack_trace,omitempty"`
	Environment string            `json:"environment"`
	Release     string            `json:"release"`
	Tags        map[string]string `json:"tags,omitempty"`
}


type Client struct{
	config Config
	http *http.Client
	Queue chan *Event
}

func(c *Client) worker(){
	for event := range c.Queue{
		c.send(event)
	}
}


func Init(config Config) *Client {
	client := &Client{
		config: config,
		http: &http.Client{
			Timeout: 3 * time.Second,
		},
		Queue: make(chan *Event, 100),
	}

	go client.worker()
	return client
}


func (c *Client) send(event *Event){
	reqBody, _ := json.Marshal(event)

	req, _ := http.NewRequest("POST", c.config.IngestURL, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp , err := c.http.Do(req)
	if err != nil{
		return
	}

	defer resp.Body.Close()
}


