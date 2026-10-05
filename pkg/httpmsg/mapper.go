package httpmsg

import (
	"net/http"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/richerror"
)

func MapRichErrKindsToHttpResponse(err error) (code int, message string) {
	richErr, ok := err.(richerror.RichError)
	if !ok {
		// TODO: log for unhandled error (get the message of the error)
		return 500, errmsg.ErrorMsgSomethingWentWrong
	}

	switch richErr.Kind() {
	case richerror.KindInvalidInput:
		return http.StatusBadRequest, richErr.Error()
	case richerror.KindNotFound:
		return http.StatusNotFound, richErr.Error()
	case richerror.KindUnauthenticated:
		return http.StatusUnauthorized, richErr.Error()
	case richerror.KindForbidden:
		return http.StatusForbidden, richErr.Error()
	case richerror.KindUnexpected:
		// TODO: log unexpected errors
		return http.StatusInternalServerError, richErr.Error()
	default:
		// TODO: log unexpected errors
		return http.StatusInternalServerError, richErr.Error()
	}
}
