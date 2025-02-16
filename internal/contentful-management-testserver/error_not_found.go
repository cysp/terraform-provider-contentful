package contentfulmanagementtestserver

import (
	"net/http"
)

func NewContentfulManagementErrorNotFoundHandler() *ContentfulManagementErrorHandler {
	return NewContentfulManagementErrorHandler(http.StatusNotFound, "NotFound", pointerTo("The resource could not be found."), nil)
}

func WriteContentfulManagementErrorNotFoundResponse(w http.ResponseWriter) error {
	return WriteContentfulManagementErrorResponse(w, http.StatusNotFound, "NotFound", pointerTo("The resource could not be found."), nil)
}
