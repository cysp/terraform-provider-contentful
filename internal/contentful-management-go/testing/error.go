package cmtesting

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewContentfulManagementErrorStatusCodeBadRequest(message *string, details []byte) *cm.ErrorStatusCode {
	return NewContentfulManagementErrorStatusCode(http.StatusBadRequest, "BadRequest", message, details)
}

func NewContentfulManagementErrorStatusCodeValidationFailed(message *string, details []byte) *cm.ErrorStatusCode {
	return NewContentfulManagementErrorStatusCode(http.StatusUnprocessableEntity, "ValidationFailed", message, details)
}

func NewContentfulManagementErrorStatusCode(statusCode int, id string, message *string, details []byte) *cm.ErrorStatusCode {
	return &cm.ErrorStatusCode{
		StatusCode: statusCode,
		Response:   cm.NewErrorApplicationJSONError(NewContentfulManagementError(id, message, details)),
	}
}

func NewContentfulManagementErrorStatusCodeNotFound(message *string, details []byte) *cm.ErrorStatusCode {
	return NewContentfulManagementErrorStatusCode(http.StatusNotFound, cm.ErrorSysIDNotFound, message, details)
}

func NewContentfulManagementErrorStatusCodeVersionMismatch(message *string, details []byte) *cm.ErrorStatusCode {
	return NewContentfulManagementErrorStatusCode(http.StatusConflict, cm.ErrorSysIDVersionMismatch, message, details)
}

func NewContentfulManagementError(id string, message *string, details []byte) cm.Error {
	return cm.Error{
		Sys:     cm.NewErrorSys(id),
		Message: cm.NewOptPointerString(message),
		Details: details,
	}
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	_ = WriteContentfulManagementErrorResponse(
		w,
		http.StatusNotFound,
		cm.ErrorSysIDNotFound,
		new("The resource could not be found."),
		nil,
	)
}
