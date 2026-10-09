package service

import (
	"context"
	"smsgateway/contract/sms"
	"smsgateway/entity"
	"smsgateway/param/smsparam"

	"google.golang.org/protobuf/proto"
)

type SMSService struct {
	smsRepo   SMSRepository
	msgBroker MessageBroker
}

func NewSMSService(smsRepo SMSRepository) SMSService {
	return SMSService{smsRepo: smsRepo}
}

type SMSRepository interface {
	CreateMessage(ctx context.Context, message *entity.Message) error
	GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (*entity.Message, error)
	UpdateStateToFailed(ctx context.Context, message entity.Message) error
	UpdateState(ctx context.Context, message entity.Message) error
}

type MessageBroker interface {
	SendToNormalQueue(ctx context.Context, body []byte) error
	SendToExpressQueue(ctx context.Context, body []byte) error
}

func (s SMSService) send(ctx context.Context, req smsparam.SendSMSRequest) (smsparam.SendSMSResponse, error) {
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

	if message.Type == entity.SmsTypeNormal {
		if err := s.msgBroker.SendToNormalQueue(ctx, body); err != nil {
			if err := s.smsRepo.UpdateStateToFailed(ctx, message); err != nil {
				return smsparam.SendSMSResponse{}, err
			}

			// TODO: return rich error from send method on msg broker, and log error(s) in it.
			return smsparam.SendSMSResponse{}, err
		}
	}

	if message.Type == entity.SmsTypeExpress {
		if err := s.msgBroker.SendToExpressQueue(ctx, body); err != nil {
			if err := s.smsRepo.UpdateStateToFailed(ctx, message); err != nil {
				return smsparam.SendSMSResponse{}, err
			}

			// TODO: return rich error from send method on msg broker, and log error(s) in it.
			return smsparam.SendSMSResponse{}, err
		}
	}

	message.Status = entity.MessageStatusQueued
	s.smsRepo.UpdateState(ctx, message)

	return smsparam.SendSMSResponse{}, nil
}

func (s SMSService) SendNormalSMS(ctx context.Context, req smsparam.SendSMSRequest) (smsparam.SendSMSResponse, error) {
	req.SmsType = entity.SmsTypeExpress

	return s.send(ctx, req)
}

func (s SMSService) SendExpressSMS(ctx context.Context, req smsparam.SendSMSRequest) (smsparam.SendSMSResponse, error) {
	req.SmsType = entity.SmsTypeExpress

	return s.send(ctx, req)
}

func (s SMSService) UpdateState(ctx context.Context, message entity.Message) error {
	if message.Status == entity.MessageStatusFailed {
		return s.smsRepo.UpdateStateToFailed(ctx, message)
	}

	return s.smsRepo.UpdateState(ctx, message)
}
