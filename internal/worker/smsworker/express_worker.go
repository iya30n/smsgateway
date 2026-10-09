package smsworker

import (
	"context"
	"smsgateway/contract/sms"
	"smsgateway/service"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

type ExpressWorker struct {
	msgBroker       MessageBrokerAdapter
	smsOp           SmsOperator
	operatorService service.OperatorService
	queueName       string
}

func NewExpressWorker(
	msgBroker MessageBrokerAdapter,
	smsOp SmsOperator,
	operatorService service.OperatorService,
	queueName string,
) *ExpressWorker {
	return &ExpressWorker{
		msgBroker:       msgBroker,
		smsOp:           smsOp,
		operatorService: operatorService,
		queueName:       queueName,
	}
}

func (w *ExpressWorker) Run(ctx context.Context) error {
	return w.msgBroker.Receive(ctx, w.queueName, func(ctx context.Context, d amqp.Delivery) error {
		return w.handle(ctx, d)
	})
}

func (w *ExpressWorker) handle(ctx context.Context, d amqp.Delivery) error {
	var req sms.ExpressSmsRequest
	if err := proto.Unmarshal(d.Body, &req); err != nil {
		_ = d.Nack(false, false)
		return nil
	}

	targets, err := w.operatorService.GetStableTargets(ctx)
	if err != nil {
		// TODO: add logger
		_ = d.Nack(false, true)
		return nil
	}

	if len(targets) == 0 {
		_ = d.Nack(false, true)
		return nil
	}

	for _, target := range targets {
		if w.sendWithRetries(ctx, target.Number, &req) {
			_ = d.Ack(false)
			return nil
		}

		if ctx.Err() != nil {
			_ = d.Nack(false, true)
			return nil
		}
	}

	_ = d.Nack(false, false) // → DLQ
	return nil
}

func (w *ExpressWorker) sendWithRetries(ctx context.Context, sourceNumber string, req *sms.ExpressSmsRequest) bool {
	for attempt := 1; attempt <= MaxAttempts; attempt++ {
		statusCode := w.smsOp.SendSMS(ctx, sourceNumber, req.ReceptorNumber, req.Content)
		if statusCode == 200 {
			return true
		}

		if !w.smsOp.IsRetryable(statusCode) {
			return false
		}

		select {
		case <-time.After(backoff(attempt)):
		case <-ctx.Done():
			return false
		}
	}

	return false
}
