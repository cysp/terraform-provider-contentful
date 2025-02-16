package contentfulmanagementtestserver

import (
	"net/http"
)

func NewContentfulManagementErrorBadRequestHandler() *ContentfulManagementErrorHandler {
	return NewContentfulManagementErrorHandler(http.StatusBadRequest, "BadRequest", pointerTo("The request was malformed or contained invalid parameters."), nil)
}

func WriteContentfulManagementErrorBadRequestResponse(w http.ResponseWriter) error {
	return WriteContentfulManagementErrorResponse(w, http.StatusBadRequest, "BadRequest", pointerTo("The request was malformed or contained invalid parameters."), nil)
}

func WriteContentfulManagementErrorBadRequestResponseWithError(w http.ResponseWriter, err error) error {
	return WriteContentfulManagementErrorBadRequestResponseWithDetails(w, err.Error())
}

func WriteContentfulManagementErrorBadRequestResponseWithDetails(w http.ResponseWriter, details string) error {
	return WriteContentfulManagementErrorResponse(w, http.StatusBadRequest, "BadRequest", pointerTo("The request was malformed or contained invalid parameters."), []byte(details))
}
