package entity

type RequestLog struct {
	RequestID string
	UserID uint
	Endpoint string
	Method string
	RequestBody string
	ResponseBody string
	StatusCode int
	Duration int64
	Level string
	ErrorCode string
	ErrorMessage string
	CreatedAt int64
}