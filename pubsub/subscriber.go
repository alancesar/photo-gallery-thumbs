package pubsub

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

type Subscriber struct {
	channel *amqp.Channel
}

type Consumer interface {
	Consume(ctx context.Context, event Event) error
}

func NewSubscriber(channel *amqp.Channel) *Subscriber {
	return &Subscriber{
		channel: channel,
	}
}

func (l Subscriber) Subscribe(ctx context.Context, queue string, consumer Consumer) error {
	delivery, err := listen(l.channel, queue)
	if err != nil {
		return err
	}

	log.Printf("listening for %s\n", queue)

	for message := range delivery {
		log.Printf("got a message from %s\n", queue)
		event := Event{
			ID:          message.MessageId,
			Headers:     message.Headers,
			ContentType: message.ContentType,
			Message:     message.Body,
		}

		if err := consumer.Consume(ctx, event); err != nil {
			fmt.Println(err)
			continue
		}

		_ = message.Ack(false)
	}

	return nil
}

func listen(channel *amqp.Channel, queue string) (<-chan amqp.Delivery, error) {
	return channel.Consume(
		queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}
