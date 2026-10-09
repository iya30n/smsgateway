package smsservice

import (
	"context"
	"smsgateway/contract/sms"
	"smsgateway/entity"
	"smsgateway/param/smsparam"

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
	})
	if err != nil {
		// TODO: do something, think about it.
	}

	if err := s.msgBroker.SendToNormalQueue(ctx, body); err != nil {
		if err := s.smsRepo.UpdateStateToFailed(ctx, message); err != nil {
			return smsparam.SendSMSResponse{}, err
		}

		// TODO: return rich error from send method on msg broker, and log error(s) in it.
		return smsparam.SendSMSResponse{}, err
	}

	message.Status = entity.MessageStatusQueued
	s.smsRepo.UpdateState(ctx, message)

	return smsparam.SendSMSResponse{}, nil
}
