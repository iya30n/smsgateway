package service

import (
	"context"
	"smsgateway/entity"
)

type SMSService struct {
	smsRepo SMSRepository
}

func NewSMSService(smsRepo SMSRepository) SMSService {
	return SMSService{smsRepo: smsRepo}
}

type SMSRepository interface {
	CreateMessage(ctx context.Context, message entity.Message) error
	GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (*entity.Message, error)
}

type SendSMSRequest struct {
	IdempotencyKey string       `json:"idempotency_key"`
	UserID       uint         `json:"user_id"`
	MobileNumber string         `json:"mobile_number"`
	Content      string         `json:"content"`
	SmsType      entity.SmsType `json:"sms_type"`
}

type SendSMSResponse struct {
	Status     string `json:"status"`
	SMSContent string `json:"sms_content"`
}

func (s SMSService) SendNormalSMS(ctx context.Context, req SendSMSRequest) (SendSMSResponse, error) {
	// TODO: think about origin number

	existingMessage, err := s.smsRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil {
		return SendSMSResponse{}, err
	}

	if existingMessage != nil {
		return SendSMSResponse{
			Status:     string(existingMessage.Status),
			SMSContent: existingMessage.Content,
		}, nil
	}

	// TODO: implement circuit breaker pattern to choose a stable operator
	// TODO: implement logic to choose an available origin number for the message

	message := entity.Message{
		UserID:            req.UserID,
		DestinationNumber: req.MobileNumber,
		Content:           req.Content,
		Type:              entity.SmsTypeNormal,
		Status:            entity.MessageStatusInitiated,
	}

	// TODO: send the message to queue

	if err := s.smsRepo.CreateMessage(ctx, message); err != nil {
		return SendSMSResponse{}, err
	}

	return SendSMSResponse{}, nil
}

func (s SMSService) SendExpressSMS(ctx context.Context, req SendSMSRequest) (SendSMSResponse, error) {
	req.SmsType = entity.SmsTypeExpress

	return s.SendNormalSMS(ctx, req)
}
