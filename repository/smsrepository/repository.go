package smsrepository

import (
	"context"
	"smsgateway/entity"
)

type SMSRepository interface {
	CreateMessage(ctx context.Context, message entity.Message) error
	GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (*entity.Message, error)
}
