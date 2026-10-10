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

func NewSMSService(smsRepo SMSRepository, msgBroker MessageBroker, operatorService service.OperatorService) SMSService {
	return SMSService{
		smsRepo:         smsRepo,
		msgBroker:       msgBroker,
		operatorService: operatorService,
	}
}

type SMSRepository interface {
	CreateMessage(ctx context.Context, message *entity.Message) error
	GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (*entity.Message, error)
	UpdateStateToFailed(ctx context.Context, message entity.Message) error
	UpdateState(ctx context.Context, message entity.Message) error
	ListMessages(ctx context.Context, filter entity.MessageFilter) ([]entity.Message, int, error)
}

type MessageBroker interface {
	SendToNormalQueue(ctx context.Context, body []byte) error
	SendToExpressQueue(ctx context.Context, body []byte) error
}
