package entity

import "github.com/shopspring/decimal"

type Transaction struct {
	ID uint
	UserID uint
	BalanceRequestID uint
	MessageID uint
	Credit decimal.Decimal
	Debit decimal.Decimal
	Type TransactionType
}

type TransactionType string

const (
	TransactionTypeSmsCharge TransactionType = "sms_charge"
	TransactionTypeRefund TransactionType = "refund"
	TransactionTypeReversal TransactionType = "reversal"
	TransactionTypemanualAdjustment TransactionType = "manual_adjustment"
)