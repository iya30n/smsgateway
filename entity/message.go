package entity

type Message struct {
	ID             uint
	UserID         uint
	IdempotencyKey string
	SourceNumber   string
	ReceptorNumber string
	Content        string
	Type           SmsType
	Status         MessageStatus
	FailedReason   string
	CreatedAt      int64
	UpdatedAt      int64
}

// MessageFilter narrows a message listing query. Zero values mean the filter is
// not applied; CreatedFrom is inclusive and CreatedTo is exclusive, both as unix
// timestamps.
type MessageFilter struct {
	UserID         uint
	ReceptorNumber string
	Status         MessageStatus
	Type           SmsType
	CreatedFrom    int64
	CreatedTo      int64
	Limit          int
	Offset         int
}

type SmsType string

const (
	SmsTypeNormal  SmsType = "normal"
	SmsTypeExpress SmsType = "express"
)

type MessageStatus string

const (
	MessageStatusInitiated MessageStatus = "initiated"
	MessageStatusQueued    MessageStatus = "queued"
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusFailed    MessageStatus = "failed"
)

func (s MessageStatus) IsTerminal() bool {
	return s == MessageStatusSent || s == MessageStatusFailed
}
