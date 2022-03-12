package listener

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"log"
)

type (
	Consumer[T any] func(ctx context.Context, data T) error

	Listener[T any] struct {
		subscription *pubsub.Subscription
	}
)

func New[T any](subscription *pubsub.Subscription) *Listener[T] {
	return &Listener[T]{
		subscription: subscription,
	}
}

func (l Listener[T]) Listen(ctx context.Context, consumer Consumer[T]) error {
	log.Printf("listening for %s\n", l.subscription.ID())

	err := l.subscription.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
		var data T
		if err := json.Unmarshal(message.Data, &data); err != nil {
			message.Ack()
			return
		}

		if err := consumer(ctx, data); err != nil {
			message.Nack()
			return
		}

		message.Ack()
	})

	if err != nil {
		log.Printf("error when trying to receive messages with subscription %s\n", l.subscription.String())
	}

	return err
}
