package smsparam

import "smsgateway/entity"

type SendSMSRequest struct {
	IdempotencyKey string         `json:"idempotency_key"`
	UserID         uint           `json:"user_id"`
	SourceNumber   string         `json:"source_number"`
	ReceptorNumber string         `json:"receptor_number"`
	Content        string         `json:"content"`
	SmsType        entity.SmsType `json:"sms_type"`
}

type SendSMSResponse struct {
	Status     string `json:"status"`
	SMSContent string `json:"sms_content"`
}
