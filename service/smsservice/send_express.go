package smsservice

import (
	"context"
	"smsgateway/contract/sms"
	"smsgateway/entity"
	"smsgateway/param/smsparam"
	"smsgateway/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func (s SMSService) SendExpressSMS(ctx context.Context, req smsparam.SendExpressSMSRequest) (smsparam.SendExpressSMSResponse, error) {
	existingMessage, err := s.smsRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil {
		return smsparam.SendExpressSMSResponse{}, err
	}

	if existingMessage != nil {
		return smsparam.SendExpressSMSResponse{
			Status:     string(existingMessage.Status),
			SMSContent: existingMessage.Content,
		}, nil
	}

	message := entity.Message{
		UserID:         req.UserID,
		IdempotencyKey: req.IdempotencyKey,
		ReceptorNumber: req.ReceptorNumber,
		Content:        req.Content,
		Type:           req.SmsType,
		Status:         entity.MessageStatusInitiated,
	}

	if err := s.smsRepo.CreateMessage(ctx, &message); err != nil {
		return smsparam.SendExpressSMSResponse{}, err
	}

	body, err := proto.Marshal(&sms.ExpressSmsRequest{
		IdempotencyKey: message.IdempotencyKey,
		MessageId:      int64(message.ID),
		ReceptorNumber: message.ReceptorNumber,
		Content:        message.Content,
		UserId:         uint64(message.UserID),
	})
	if err != nil {
		logger.Logger.Error("failed to marshal express sms payload",
			zap.Uint("message_id", message.ID),
			zap.Error(err),
		)

		if refundErr := s.smsRepo.UpdateStateToFailed(ctx, message); refundErr != nil {
			logger.Logger.Error("failed to refund after marshal failure",
				zap.Uint("message_id", message.ID),
				zap.Error(refundErr),
			)
		}

		return smsparam.SendExpressSMSResponse{}, err
	}

	if err := s.msgBroker.SendToExpressQueue(ctx, body); err != nil {
		logger.Logger.Error("failed to publish to sms.express, refunding",
			zap.Uint("message_id", message.ID),
			zap.Error(err),
		)

		if refundErr := s.smsRepo.UpdateStateToFailed(ctx, message); refundErr != nil {
			logger.Logger.Error("failed to refund after publish failure",
				zap.Uint("message_id", message.ID),
				zap.Error(refundErr),
			)

			return smsparam.SendExpressSMSResponse{}, refundErr
		}

		return smsparam.SendExpressSMSResponse{}, err
	}

	message.Status = entity.MessageStatusQueued
	if err := s.smsRepo.UpdateState(ctx, message); err != nil {
		logger.Logger.Error("failed to persist queued state",
			zap.Uint("message_id", message.ID),
			zap.Error(err),
		)
	}

	return smsparam.SendExpressSMSResponse{}, nil
}
