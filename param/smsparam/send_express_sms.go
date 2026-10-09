package smsparam

import "smsgateway/entity"

type SendExpressSMSRequest struct {
	IdempotencyKey string         `json:"idempotency_key"`
	UserID         uint           `json:"user_id"`
	ReceptorNumber string         `json:"receptor_number"`
	Content        string         `json:"content"`
	SmsType        entity.SmsType `json:"sms_type"`
}

type SendExpressSMSResponse struct {
	Status     string `json:"status"`
	SMSContent string `json:"sms_content"`
}
