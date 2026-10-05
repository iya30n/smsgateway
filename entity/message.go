package entity

type Message struct {
	ID uint
	UserID uint
	IdempotencyKey string
	OperatorID uint
	OriginNumber string
	DestinationNumber string
	Content string
	Status MessageStatus
	FailedReason string
	CreatedAt int64
	UpdatedAt int64
}

type MessageStatus string

const (
	MessageStatusInitiated MessageStatus = "initiated"
	MessageStatusQueued MessageStatus = "queued"
	MessageStatusSent MessageStatus = "sent"
	MessageStatusFailed MessageStatus = "failed"
)