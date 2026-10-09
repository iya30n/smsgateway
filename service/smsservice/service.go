package smsservice

import (
	"context"
	"smsgateway/entity"
	"smsgateway/service"
)

type SMSService struct {
	smsRepo         SMSRepository
	msgBroker       MessageBroker
	operatorService service.OperatorService
}

func NewSMSService(smsRepo SMSRepository, operatorService service.OperatorService) SMSService {
	return SMSService{
		smsRepo:         smsRepo,
		operatorService: operatorService,
	}
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

func (s SMSService) UpdateState(ctx context.Context, message entity.Message) error {
	if message.Status == entity.MessageStatusFailed {
		return s.smsRepo.UpdateStateToFailed(ctx, message)
	}

	return s.smsRepo.UpdateState(ctx, message)
}
