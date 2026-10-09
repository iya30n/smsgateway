package smsoperator

type Operator interface {
	SendSMS(sourceNumber string, destinationNumber string, message string) error
	IsRetryable(httpStatusCode uint) bool
}