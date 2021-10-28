package pubsub

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

type Subscriber struct {
	channel *amqp.Channel
	queue   amqp.Queue
}

type Consumer interface {
	Consume(ctx context.Context, event Event) error
}

func NewAmqpSubscriber(channel *amqp.Channel, queue amqp.Queue) *Subscriber {
	return &Subscriber{
		channel: channel,
		queue:   queue,
	}
}

func (l Subscriber) Subscribe(ctx context.Context, consumer Consumer) error {
	delivery, err := listen(l.channel, l.queue)
	if err != nil {
		return err
	}

	log.Printf("listening for %s", l.queue.Name)

	for message := range delivery {
		log.Printf(fmt.Sprintf("got a message from %s", l.queue.Name))
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

func listen(channel *amqp.Channel, queue amqp.Queue) (<-chan amqp.Delivery, error) {
	return channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}
