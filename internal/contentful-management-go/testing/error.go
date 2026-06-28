package cmtesting

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewContentfulManagementErrorStatusCodeBadRequest(message *string, details []byte) *cm.ErrorStatusCode {
	return &cm.ErrorStatusCode{
		StatusCode: http.StatusBadRequest,
		Response:   cm.NewErrorApplicationJSONError(NewContentfulManagementError("BadRequest", message, details)),
	}
}

func NewContentfulManagementErrorStatusCodeNotFound(message *string, details []byte) *cm.ErrorStatusCode {
	return &cm.ErrorStatusCode{
		StatusCode: http.StatusNotFound,
		Response:   cm.NewErrorApplicationJSONError(NewContentfulManagementError(cm.ErrorSysIDNotFound, message, details)),
	}
}

func NewContentfulManagementErrorStatusCodeVersionMismatch(message *string, details []byte) *cm.ErrorStatusCode {
	return &cm.ErrorStatusCode{
		StatusCode: http.StatusConflict,
		Response:   cm.NewErrorApplicationJSONError(NewContentfulManagementError(cm.ErrorSysIDVersionMismatch, message, details)),
	}
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
