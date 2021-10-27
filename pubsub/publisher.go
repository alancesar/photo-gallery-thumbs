package pubsub

import (
	"github.com/streadway/amqp"
)

type Publisher struct {
	channel  *amqp.Channel
	exchange string
}

func NewAmpqPublisher(channel *amqp.Channel, exchange string) *Publisher {
	return &Publisher{
		channel:  channel,
		exchange: exchange,
	}
}

func (p Publisher) Publish(event Event) error {
	return p.channel.Publish(p.exchange, "", false, false, amqp.Publishing{
		ContentType: event.ContentType,
		Headers:     event.Headers,
		Body:        event.Message,
	})
}
