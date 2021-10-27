package consumer

import (
	"context"
	"encoding/json"
	"github.com/alancesar/photo-gallery/thumbs/image"
	"github.com/alancesar/photo-gallery/thumbs/pubsub"
	"github.com/alancesar/photo-gallery/thumbs/worker"
)

type Message struct {
	Filename string `json:"filename"`
	image.Metadata
}

type Worker interface {
	CreateThumbnails(context.Context, string, image.Metadata) error
}

type Consumer struct {
	worker Worker
}

func NewConsumer(worker *worker.Thumbs) *Consumer {
	return &Consumer{
		worker: worker,
	}
}

func (c Consumer) Consume(ctx context.Context, event pubsub.Event) error {
	var message Message
	if err := json.Unmarshal(event.Message, &message); err != nil {
		return err
	}
	return c.worker.CreateThumbnails(ctx, message.Filename, message.Metadata)
}
