package publisher

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"log"
)

type Publisher[T any] struct {
	topic *pubsub.Topic
}

func New[T any](topic *pubsub.Topic) *Publisher[T] {
	return &Publisher[T]{
		topic: topic,
	}
}

func (p Publisher[T]) Publish(ctx context.Context, data T, attributes map[string]string) {
	bytes, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return
	}

	p.topic.Publish(ctx, &pubsub.Message{
		Data:       bytes,
		Attributes: attributes,
	})
}
