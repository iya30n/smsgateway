package userhandler

import (
	"smsgateway/service"
	"smsgateway/validator/uservalidator"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	e             *echo.Echo
	userSvc       service.UserService
	userValidator uservalidator.Validator
}

func New(e *echo.Echo, userSvc service.UserService, userValidator uservalidator.Validator) Handler {
	return Handler{e: e, userSvc: userSvc, userValidator: userValidator}
}

func (h Handler) SetupRoutes() {
	userGroup := h.e.Group("/user")
	userGroup.GET("/increase-balance", h.IncreaseBalance)
}
