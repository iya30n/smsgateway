package smsworker

import (
	"context"
	"fmt"
	"smsgateway/contract/sms"
	"smsgateway/entity"
	"smsgateway/pkg/logger"
	"smsgateway/service/smsservice"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const MaxAttempts = 3

type Worker struct {
	msgBroker MessageBrokerAdapter
	smsOp     SmsOperator
	smsSvc    smsservice.SMSService
	queueName string
}

func New(msgBroker MessageBrokerAdapter, smsOp SmsOperator, smsSvc smsservice.SMSService, queueName string) *Worker {
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
	logger.Logger.Info("worker starting",
		zap.String("worker", "normal"),
		zap.String("queue", w.queueName),
	)

	return w.msgBroker.Receive(ctx, w.queueName, func(ctx context.Context, d amqp.Delivery) error {
		return w.handle(ctx, d)
	})
}

func (w *Worker) handle(ctx context.Context, d amqp.Delivery) error {
	var req sms.SmsRequest
	if err := proto.Unmarshal(d.Body, &req); err != nil {
		logger.Logger.Warn("bad message, moving to DLQ",
			zap.String("queue", w.queueName),
			zap.Error(err),
		)

		_ = d.Nack(false, false)
		return nil
	}

	var smsStatusCode uint
	for attempt := 1; attempt <= MaxAttempts; attempt++ {
		smsStatusCode = w.smsOp.SendSMS(ctx, req.SourceNumber, req.ReceptorNumber, req.Content)
		if smsStatusCode == 200 {
			if err := w.markSent(ctx, req.MessageId); err != nil {
				logger.Logger.Error("delivered but failed to persist sent state, requeueing",
					zap.Int64("message_id", req.MessageId),
					zap.Error(err),
				)

				_ = d.Nack(false, true)
				return nil
			}

			logger.Logger.Info("sms sent",
				zap.Int64("message_id", req.MessageId),
				zap.Int("attempt", attempt),
			)

			_ = d.Ack(false)
			return nil
		}

		if !w.smsOp.IsRetryable(smsStatusCode) {
			break
		}

		delay := backoff(attempt)
		logger.Logger.Warn("operator returned a retryable status, backing off",
			zap.Int64("message_id", req.MessageId),
			zap.Uint("status_code", smsStatusCode),
			zap.Int("attempt", attempt),
			zap.Duration("backoff", delay),
		)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			_ = d.Nack(false, true)
			return nil
		}
	}

	if err := w.markFailed(ctx, req.MessageId, req.UserId, smsStatusCode); err != nil {
		logger.Logger.Error("failed to persist failed state",
			zap.Int64("message_id", req.MessageId),
			zap.Error(err),
		)
	}

	logger.Logger.Warn("sms failed, moving to DLQ",
		zap.Int64("message_id", req.MessageId),
		zap.Uint("status_code", smsStatusCode),
		zap.Int("attempt", MaxAttempts),
	)

	_ = d.Nack(false, false) // → DLQ
	return nil
}

func (w *Worker) markSent(ctx context.Context, messageID int64) error {
	return w.smsSvc.UpdateState(ctx, entity.Message{
		ID:     uint(messageID),
		Status: entity.MessageStatusSent,
	})
}

func (w *Worker) markFailed(ctx context.Context, messageID int64, userID uint64, statusCode uint) error {
	return w.smsSvc.UpdateState(ctx, entity.Message{
		ID:           uint(messageID),
		UserID:       uint(userID),
		Status:       entity.MessageStatusFailed,
		FailedReason: fmt.Sprintf("operator returned status %d after %d attempts", statusCode, MaxAttempts),
	})
}

func backoff(attempt int) time.Duration {
	base := time.Second
	d := base * time.Duration(1<<uint(attempt-1)) // 1,2,4,8,16
	if d > 30*time.Second {
		d = 30 * time.Second
	}
	return d
}
