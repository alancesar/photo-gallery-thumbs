package image

import (
	"encoding/json"
	"github.com/alancesar/photo-gallery/thumbs/pubsub"
)

const (
	contentType = "application/json"
)

type Message struct {
	Filename string      `json:"filename"`
	Property string      `json:"property"`
	Payload  interface{} `json:"payload"`
}

type Publisher interface {
	Publish(event pubsub.Event) error
}

type Producer struct {
	publisher Publisher
}

func NewProducer(publisher Publisher) *Producer {
	return &Producer{
		publisher: publisher,
	}
}

func (p Producer) Produce(message Message) error {
	bytes, err := json.Marshal(&message)
	if err != nil {
		return err
	}

	event := pubsub.Event{
		ContentType: contentType,
		Message:     bytes,
	}

	return p.publisher.Publish(event)
}
