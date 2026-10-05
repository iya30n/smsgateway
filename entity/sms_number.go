package entity

type SmsNumber struct {
	ID uint
	Number string
	OperatorID uint
	IsActive bool
	Type SmsType
}

type SmsType string

const (
	SmsTypeNormal SmsType = "normal"
	SmsTypeExpress SmsType = "express"
);