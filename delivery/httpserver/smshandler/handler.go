package smshandler

import (
	"smsgateway/service"
	"smsgateway/validator/smsvalidator"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	e             *echo.Echo
	smsSvc       service.SMSService
	smsValidator smsvalidator.Validator
}

func New(e *echo.Echo, smsSvc service.SMSService, smsValidator smsvalidator.Validator) Handler {
	return Handler{e: e, smsSvc: smsSvc, smsValidator: smsValidator}
}

func (h Handler) SetupRoutes() {
	userGroup := h.e.Group("/sms")
	userGroup.POST("/send", h.SendNormalSMS)
	userGroup.POST("/send/express", h.SendExpressSMS)
}
