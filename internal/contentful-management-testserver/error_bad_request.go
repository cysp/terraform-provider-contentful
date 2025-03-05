package contentfulmanagementtestserver

import (
	"net/http"
)

func NewContentfulManagementErrorBadRequestHandler() *ContentfulManagementErrorHandler {
	return NewContentfulManagementErrorHandler(http.StatusBadRequest, "BadRequest", pointerTo("The request was malformed or contained invalid parameters."), nil)
}

func WriteContentfulManagementErrorBadRequestResponse(responseWriter http.ResponseWriter) error {
	return WriteContentfulManagementErrorResponse(responseWriter, http.StatusBadRequest, "BadRequest", pointerTo("The request was malformed or contained invalid parameters."), nil)
}

func WriteContentfulManagementErrorBadRequestResponseWithError(responseWriter http.ResponseWriter, err error) error {
	return WriteContentfulManagementErrorBadRequestResponseWithDetails(responseWriter, err.Error())
}

func WriteContentfulManagementErrorBadRequestResponseWithDetails(responseWriter http.ResponseWriter, details string) error {
	return WriteContentfulManagementErrorResponse(responseWriter, http.StatusBadRequest, "BadRequest", pointerTo("The request was malformed or contained invalid parameters."), []byte(details))
}
