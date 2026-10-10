package smsworker

import (
	"context"
	"fmt"
	"smsgateway/contract/sms"
	"smsgateway/entity"
	"smsgateway/pkg/logger"
	"smsgateway/service"
	"smsgateway/service/smsservice"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type ExpressWorker struct {
	msgBroker       MessageBrokerAdapter
	smsOp           SmsOperator
	operatorService service.OperatorService
	smsSvc          smsservice.SMSService
	queueName       string
}

func NewExpressWorker(
	msgBroker MessageBrokerAdapter,
	smsOp SmsOperator,
	operatorService service.OperatorService,
	smsSvc smsservice.SMSService,
	queueName string,
) *ExpressWorker {
	return &ExpressWorker{
		msgBroker:       msgBroker,
		smsOp:           smsOp,
		operatorService: operatorService,
		smsSvc:          smsSvc,
		queueName:       queueName,
	}
}

func (w *ExpressWorker) Run(ctx context.Context) error {
	logger.Logger.Info("worker starting",
		zap.String("worker", "express"),
		zap.String("queue", w.queueName),
	)

	return w.msgBroker.Receive(ctx, w.queueName, func(ctx context.Context, d amqp.Delivery) error {
		return w.handle(ctx, d)
	})
}

func (w *ExpressWorker) handle(ctx context.Context, d amqp.Delivery) error {
	var req sms.ExpressSmsRequest
	if err := proto.Unmarshal(d.Body, &req); err != nil {
		logger.Logger.Warn("bad message, moving to DLQ",
			zap.String("queue", w.queueName),
			zap.Error(err),
		)

		_ = d.Nack(false, false)
		return nil
	}

	targets, err := w.operatorService.GetStableTargets(ctx)
	if err != nil {
		logger.Logger.Error("failed to resolve operator targets, requeueing",
			zap.Int64("message_id", req.MessageId),
			zap.Error(err),
		)

		_ = d.Nack(false, true)
		return nil
	}

	if len(targets) == 0 {
		logger.Logger.Warn("no operator target available, requeueing",
			zap.Int64("message_id", req.MessageId),
		)

		_ = d.Nack(false, true)
		return nil
	}

	for _, target := range targets {
		if w.sendWithRetries(ctx, target.Number, &req) {
			if err := w.smsSvc.UpdateState(ctx, entity.Message{
				ID:     uint(req.MessageId),
				Status: entity.MessageStatusSent,
			}); err != nil {
				logger.Logger.Error("delivered but failed to persist sent state, requeueing",
					zap.Int64("message_id", req.MessageId),
					zap.Uint("operator_id", target.OperatorID),
					zap.Error(err),
				)

				_ = d.Nack(false, true)
				return nil
			}

			logger.Logger.Info("express sms sent",
				zap.Int64("message_id", req.MessageId),
				zap.Uint("operator_id", target.OperatorID),
			)

			_ = d.Ack(false)
			return nil
		}

		if ctx.Err() != nil {
			_ = d.Nack(false, true)
			return nil
		}

		logger.Logger.Warn("operator target rejected the message, trying the next",
			zap.Int64("message_id", req.MessageId),
			zap.Uint("operator_id", target.OperatorID),
		)
	}

	if err := w.smsSvc.UpdateState(ctx, entity.Message{
		ID:           uint(req.MessageId),
		UserID:       uint(req.UserId),
		Status:       entity.MessageStatusFailed,
		FailedReason: fmt.Sprintf("no operator target accepted the message after %d attempts each", MaxAttempts),
	}); err != nil {
		logger.Logger.Error("failed to persist failed state",
			zap.Int64("message_id", req.MessageId),
			zap.Error(err),
		)
	}

	logger.Logger.Warn("express sms failed, moving to DLQ",
		zap.Int64("message_id", req.MessageId),
		zap.Int("targets_tried", len(targets)),
		zap.Int("attempts_each", MaxAttempts),
	)

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
