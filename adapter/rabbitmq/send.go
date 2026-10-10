package rabbitmq

import (
	"context"
	"fmt"
	"smsgateway/pkg/logger"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

func (a *Adapter) SendToNormalQueue(ctx context.Context, body []byte) error {
	return a.send(ctx, QueueNormal, body)
}

func (a *Adapter) SendToExpressQueue(ctx context.Context, body []byte) error {
	return a.send(ctx, QueueExpress, body)
}

func (a *Adapter) send(ctx context.Context, queueName string, body []byte) error {
	ch, err := a.conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	// enable confirm mode
	if err := ch.Confirm(false); err != nil {
		return fmt.Errorf("confirm mode: %w", err)
	}

	// confirmations channel
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	// publish
	if err := ch.PublishWithContext(ctx,
		Exchange,
		queueName,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/x-protobuf",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Timestamp:    time.Now(),
		},
	); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	// wait to conferms
	select {
	case c, ok := <-confirms:
		if !ok {
			return fmt.Errorf("confirm channel closed")
		}
		if !c.Ack {
			logger.Logger.Error("broker did not confirm publish",
				zap.String("queue", queueName),
			)

			return fmt.Errorf("publish not confirmed by broker")
		}

		logger.Logger.Debug("message published", zap.String("queue", queueName))

		return nil
	case <-ctx.Done():
		logger.Logger.Warn("publish cancelled by context",
			zap.String("queue", queueName),
			zap.Error(ctx.Err()),
		)

		return ctx.Err()
	}
}

