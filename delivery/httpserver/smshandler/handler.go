package smshandler

import (
	"smsgateway/service"
	"smsgateway/validator/smsvalidator"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type Handler struct {
	e            *echo.Echo
	smsSvc       service.SMSService
	smsValidator smsvalidator.Validator
}

func New(e *echo.Echo, smsSvc service.SMSService, smsValidator smsvalidator.Validator) Handler {
	return Handler{e: e, smsSvc: smsSvc, smsValidator: smsValidator}
}

func (h Handler) SetupRoutes() {
	userGroup := h.e.Group("/sms")
	// NOTE: we can add dynamic rate limiter based on users usages and operator's capacity,
	// for now we will use a static rate limiter, because it's out of scope of this task,
	// and it will add more complexity to the code, and we don't have enough time to implement it properly.

	// NOTE: rate limit counts can change based on SLA and operator's capacity.
	// for now we will use a static rate limit.
	userGroup.POST("/send", h.SendNormalSMS, middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(50.0)))
	userGroup.POST("/send/express", h.SendExpressSMS, middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(5.0)))
}
