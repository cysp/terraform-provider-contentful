package testserver

import (
	"net/http"
)

func NewContentfulManagementErrorNotFoundHandler() *ContentfulManagementErrorHandler {
	return NewContentfulManagementErrorHandler(http.StatusNotFound, "NotFound", pointerTo("The resource could not be found."), nil)
}

func WriteContentfulManagementErrorNotFoundResponse(responseWriter http.ResponseWriter) error {
	return WriteContentfulManagementErrorResponse(responseWriter, http.StatusNotFound, "NotFound", pointerTo("The resource could not be found."), nil)
}
