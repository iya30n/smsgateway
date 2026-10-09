package smsservice

import (
	"context"
	"smsgateway/contract/sms"
	"smsgateway/entity"
	"smsgateway/param/smsparam"

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
	})
	if err != nil {
		// TODO: do something, think about it.
	}

	if err := s.msgBroker.SendToExpressQueue(ctx, body); err != nil {
		if err := s.smsRepo.UpdateStateToFailed(ctx, message); err != nil {
			return smsparam.SendExpressSMSResponse{}, err
		}

		// TODO: return rich error from send method on msg broker, and log error(s) in it.
		return smsparam.SendExpressSMSResponse{}, err
	}

	message.Status = entity.MessageStatusQueued
	s.smsRepo.UpdateState(ctx, message)

	return smsparam.SendExpressSMSResponse{}, nil
}
