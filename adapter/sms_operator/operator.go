package smsoperator

type Operator interface {
	GetName() string
	SendSMS(sourceNumber string, destinationNumber string, message string) error
}