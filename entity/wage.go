package entity

type Wage struct {
	ID uint
	UserID uint
	Amount float64
	Type WageType
	CreatedAt int64
}

type WageType string

const (
	WageTypeNormal WageType = "normal"
	WageTypeExpress WageType = "express"
)