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
	UserID         uint         `json:"user_id"`
	SourceNumber   string       `json:"source_number"`
	ReceptorNumber       string       `json:"receptor_number"`
	Content        string       `json:"content"`
	SmsType        entity.SmsType `json:"sms_type"`
}

type SendSMSResponse struct {
	Status     string `json:"status"`
	SMSContent string `json:"sms_content"`
}

func (s SMSService) SendNormalSMS(ctx context.Context, req SendSMSRequest) (SendSMSResponse, error) {
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

	message := entity.Message{
		UserID:       req.UserID,
		SourceNumber: req.SourceNumber,
		ReceptorNumber:     req.ReceptorNumber,
		Content:      req.Content,
		Type:         entity.SmsTypeNormal,
		Status:       entity.MessageStatusInitiated,
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
