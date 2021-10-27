package image

import (
	"encoding/json"
	"github.com/alancesar/photo-gallery/thumbs/pubsub"
)

const (
	eventType    = "WORKER"
	contentType  = "application/json"
	eventTypeKey = "event-type"
)

type message struct {
	Filename string  `json:"filename"`
	Images   []Image `json:"images"`
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

func (p Producer) Produce(filename string, images []Image) error {
	bytes, err := json.Marshal(&message{
		Filename: filename,
		Images:   images,
	})
	if err != nil {
		return err
	}

	event := pubsub.Event{
		Headers:     map[string]interface{}{eventTypeKey: eventType},
		ContentType: contentType,
		Message:     bytes,
	}

	return p.publisher.Publish(event)
}
