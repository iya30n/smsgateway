package smsservice

import (
	"context"
	"smsgateway/entity"
)

func (s SMSService) UpdateState(ctx context.Context, message entity.Message) error {
	if message.Status == entity.MessageStatusFailed {
		return s.smsRepo.UpdateStateToFailed(ctx, message)
	}

	return s.smsRepo.UpdateState(ctx, message)
}
