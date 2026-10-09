package smsworker

import (
	"context"
	"smsgateway/contract/sms"
	"smsgateway/service"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

const MaxAttempts = 3

type Worker struct {
	msgBroker MessageBrokerAdapter
	smsOp     SmsOperator
	smsSvc    service.SMSService
	queueName	string
}

func New(msgBroker MessageBrokerAdapter, smsOp SmsOperator, smsSvc service.SMSService, queueName string) *Worker {
	return &Worker{msgBroker, smsOp, smsSvc, queueName}
}

type MessageBrokerAdapter interface {
	Receive(ctx context.Context, queueName string, handler func(ctx context.Context, d amqp.Delivery) error) error
}

type SmsOperator interface {
	SendSMS(ctx context.Context, sourceNumber string, receptorNumber string, message string) uint
	IsRetryable(httpStatusCode uint) bool
}

func (w *Worker) Run(ctx context.Context) error {
	// w.logger.Info("worker starting", "queue", w.queue, "provider", w.provider.Name())

	return w.msgBroker.Receive(ctx, w.queueName, func(ctx context.Context, d amqp.Delivery) error {
		return w.handle(ctx, d)
	})
}

func (w *Worker) handle(ctx context.Context, d amqp.Delivery) error {
	var req sms.SmsRequest
	if err := proto.Unmarshal(d.Body, &req); err != nil {
		// TODO: w.logger.Warn("bad message, moving to DLQ", "err", err)
		_ = d.Nack(false, false)
		return nil
	}

	var smsStatusCode uint
	for attempt := 1; attempt <= MaxAttempts; attempt++ {
		smsStatusCode = w.smsOp.SendSMS(ctx, req.SourceNumber, req.ReceptorNumber, req.Content)
		if smsStatusCode == 200 {
			_ = d.Ack(false)
			return nil
		}

		if !w.smsOp.IsRetryable(smsStatusCode) {
			break
		}

		delay := backoff(attempt)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			_ = d.Nack(false, true)
			return nil
		}
	}

	// TODO: w.logger.Warn("sms failed", "id", req.Id, "attempt", MaxAttempts, "err", lastErr)
	_ = d.Nack(false, false) // → DLQ
	return nil
}

func backoff(attempt int) time.Duration {
	base := time.Second
	d := base * time.Duration(1<<uint(attempt-1)) // 1,2,4,8,16
	if d > 30*time.Second {
		d = 30 * time.Second
	}
	return d
}
