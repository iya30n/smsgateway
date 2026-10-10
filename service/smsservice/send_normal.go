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

func (s SMSService) SendNormalSMS(ctx context.Context, req smsparam.SendSMSRequest) (smsparam.SendSMSResponse, error) {
	existingMessage, err := s.smsRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil {
		return smsparam.SendSMSResponse{}, err
	}

	if existingMessage != nil {
		return smsparam.SendSMSResponse{
			Status:     string(existingMessage.Status),
			SMSContent: existingMessage.Content,
		}, nil
	}

	message := entity.Message{
		UserID:         req.UserID,
		IdempotencyKey: req.IdempotencyKey,
		SourceNumber:   req.SourceNumber,
		ReceptorNumber: req.ReceptorNumber,
		Content:        req.Content,
		Type:           req.SmsType,
		Status:         entity.MessageStatusInitiated,
	}

	if err := s.smsRepo.CreateMessage(ctx, &message); err != nil {
		return smsparam.SendSMSResponse{}, err
	}

	body, err := proto.Marshal(&sms.SmsRequest{
		IdempotencyKey: message.IdempotencyKey,
		MessageId:      int64(message.ID),
		SourceNumber:   message.SourceNumber,
		ReceptorNumber: message.ReceptorNumber,
		Content:        message.Content,
		UserId:         uint64(message.UserID),
	})
	if err != nil {
		logger.Logger.Error("failed to marshal normal sms payload",
			zap.Uint("message_id", message.ID),
			zap.Error(err),
		)

		if refundErr := s.smsRepo.UpdateStateToFailed(ctx, message); refundErr != nil {
			logger.Logger.Error("failed to refund after marshal failure",
				zap.Uint("message_id", message.ID),
				zap.Error(refundErr),
			)
		}

		return smsparam.SendSMSResponse{}, err
	}

	if err := s.msgBroker.SendToNormalQueue(ctx, body); err != nil {
		logger.Logger.Error("failed to publish to sms.normal, refunding",
			zap.Uint("message_id", message.ID),
			zap.Error(err),
		)

		if refundErr := s.smsRepo.UpdateStateToFailed(ctx, message); refundErr != nil {
			logger.Logger.Error("failed to refund after publish failure",
				zap.Uint("message_id", message.ID),
				zap.Error(refundErr),
			)

			return smsparam.SendSMSResponse{}, refundErr
		}

		return smsparam.SendSMSResponse{}, err
	}

	message.Status = entity.MessageStatusQueued
	if err := s.smsRepo.UpdateState(ctx, message); err != nil {
		logger.Logger.Error("failed to persist queued state",
			zap.Uint("message_id", message.ID),
			zap.Error(err),
		)
	}

	return smsparam.SendSMSResponse{}, nil
}
