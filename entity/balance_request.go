package entity

import "github.com/shopspring/decimal"

type BalanceRequest struct {
	ID        uint
	UserID    uint
	Amount    decimal.Decimal
	Status    string
	CreatedAt int64
}

type BalanceRequestStatus string

const (
	BalanceRequestStatusInitiated BalanceRequestStatus = "initiated"
	BalanceRequestStatusPending   BalanceRequestStatus = "pending"
	BalanceRequestStatusApproved  BalanceRequestStatus = "approved"
	BalanceRequestStatusRejected  BalanceRequestStatus = "rejected"
)
