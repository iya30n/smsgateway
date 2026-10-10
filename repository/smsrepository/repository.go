package smsrepository

import (
	"context"
	"smsgateway/entity"
)

type SMSRepository interface {
	CreateMessage(ctx context.Context, message *entity.Message) error
	GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (*entity.Message, error)
	UpdateStateToFailed(ctx context.Context, message entity.Message) error
	UpdateState(ctx context.Context, message entity.Message) error
	ListMessages(ctx context.Context, filter entity.MessageFilter) ([]entity.Message, int, error)
}
