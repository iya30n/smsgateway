package userparam

import "github.com/shopspring/decimal"

type IncreaseBalanceRequest struct {
	UserID uint            `json:"user_id"`
	Amount decimal.Decimal `json:"amount"`
}

type IncreaseBalanceResponse struct {
	NewBalance decimal.Decimal `json:"new_balance"`
}
