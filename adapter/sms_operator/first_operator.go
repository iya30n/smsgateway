package smsoperator

import (
	"context"
	"time"
)

type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

type FirstOperatorOperator struct {
	config Config
}

func NewFirstOperatorOperator(config Config) *FirstOperatorOperator {
	return &FirstOperatorOperator{config}
}

func (h *FirstOperatorOperator) SendSMS(ctx context.Context, sourceNumber string, receptorNumber string, message string) uint {
	// Implement the logic to send an SMS using First Operator's API
	// This may involve making an HTTP request to h.config.BaseURL with the necessary parameters
	// Handle the response and return any errors if the sending fails
	return 200
}

func (h FirstOperatorOperator) IsRetryable(httpStatusCode uint) bool {
	return (httpStatusCode >= 500 || httpStatusCode == 429)
}
