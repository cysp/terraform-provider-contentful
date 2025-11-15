package testing

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewContentfulManagementErrorStatusCodeBadRequest(message *string, details []byte) *cm.ApplicationVndContentfulManagementV1JSONErrorStatusCode {
	return &cm.ApplicationVndContentfulManagementV1JSONErrorStatusCode{
		StatusCode: http.StatusBadRequest,
		Response:   cm.NewErrorApplicationVndContentfulManagementV1JSONError(NewContentfulManagementError("BadRequest", message, details)),
	}
}

func NewContentfulManagementErrorStatusCodeNotFound(message *string, details []byte) *cm.ApplicationVndContentfulManagementV1JSONErrorStatusCode {
	return &cm.ApplicationVndContentfulManagementV1JSONErrorStatusCode{
		StatusCode: http.StatusNotFound,
		Response:   cm.NewErrorApplicationVndContentfulManagementV1JSONError(NewContentfulManagementError(cm.ErrorSysIDNotFound, message, details)),
	}
}

func NewContentfulManagementErrorStatusCode(statusCode int, id string, message *string, details []byte) *cm.ApplicationVndContentfulManagementV1JSONErrorStatusCode {
	return &cm.ApplicationVndContentfulManagementV1JSONErrorStatusCode{
		StatusCode: statusCode,
		Response:   cm.NewErrorApplicationVndContentfulManagementV1JSONError(NewContentfulManagementError(id, message, details)),
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
		pointerTo("The resource could not be found."),
		nil,
	)
}
