package contentfulmanagementtestserver

import (
	"net/http"
)

func WriteContentfulManagementErrorBadRequestResponse(w http.ResponseWriter) error {
	return WriteContentfulManagementErrorResponse(w, http.StatusBadRequest, "BadRequest", pointerTo("The request was malformed or contained invalid parameters."), nil)
}

func WriteContentfulManagementErrorBadRequestResponseWithDetails(w http.ResponseWriter, details string) error {
	return WriteContentfulManagementErrorResponse(w, http.StatusBadRequest, "BadRequest", pointerTo("The request was malformed or contained invalid parameters."), []byte(details))
}

func WriteContentfulManagementErrorBadRequestResponseWithError(w http.ResponseWriter, err error) error {
	return WriteContentfulManagementErrorBadRequestResponseWithDetails(w, err.Error())
}
