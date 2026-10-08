package rabbitmq

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (a *Adapter) Receive(ctx context.Context, queueName string, handler func(ctx context.Context, d amqp.Delivery) error) error {
	ch, err := a.conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	if err := ch.Qos(a.config.PrefetchCount, 0, false); err != nil {
		return fmt.Errorf("qos: %w", err)
	}

	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("delivery channel closed")
			}
			// handler decides to say ack/nack
			if err := handler(ctx, d); err != nil {
				// requeue of the handler itself didn't ack/nack.
				_ = d.Nack(false, true)
			}
		}
	}
}
