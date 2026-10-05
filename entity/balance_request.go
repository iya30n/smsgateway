package entity

type BalanceRequest struct {
	ID uint
	UserID uint
	Amount float64
	Status string
	CreatedAt int64
}

type BalanceRequestStatus string

const (
	BalanceRequestStatusInitiated BalanceRequestStatus = "initiated"
	BalanceRequestStatusPending BalanceRequestStatus = "pending"
	BalanceRequestStatusApproved BalanceRequestStatus = "approved"
	BalanceRequestStatusRejected BalanceRequestStatus = "rejected"
)