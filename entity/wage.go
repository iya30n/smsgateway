package entity

import "github.com/shopspring/decimal"

type Wage struct {
	ID        uint
	UserID    uint
	Amount    decimal.Decimal
	Type      WageType
	CreatedAt int64
}

type WageType string

const (
	WageTypeNormal  WageType = "normal"
	WageTypeExpress WageType = "express"
)
