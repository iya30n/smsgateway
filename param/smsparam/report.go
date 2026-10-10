package smsparam

import "smsgateway/entity"

const (
	DefaultReportPage    = 1
	DefaultReportPerPage = 20
	MaxReportPerPage     = 100
	ReportDateLayout     = "2006-01-02"
)

type ReportSMSRequest struct {
	UserID   uint                 `query:"user_id"`
	Receptor string               `query:"receptor"`
	Status   entity.MessageStatus `query:"status"`
	SmsType  entity.SmsType       `query:"sms_type"`
	From     string               `query:"from"`
	To       string               `query:"to"`
	Page     int                  `query:"page"`
	PerPage  int                  `query:"per_page"`
}

type ReportSMSResponse struct {
	Items      []ReportSMSItem `json:"items"`
	Pagination Pagination      `json:"pagination"`
}

type ReportSMSItem struct {
	ID             uint                 `json:"id"`
	UserID         uint                 `json:"user_id"`
	SourceNumber   string               `json:"source_number"`
	ReceptorNumber string               `json:"receptor_number"`
	Content        string               `json:"content"`
	Type           entity.SmsType       `json:"type"`
	Status         entity.MessageStatus `json:"status"`
	FailedReason   string               `json:"failed_reason"`
	CreatedAt      int64                `json:"created_at"`
}

type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
