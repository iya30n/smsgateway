package entity

import "github.com/shopspring/decimal"

type User struct {
	ID uint
	Name string
	Username string
	Balance decimal.Decimal
}