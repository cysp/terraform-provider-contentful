package contentfulmanagementtestserver

import (
	"net/http"
)

func WriteContentfulManagementErrorNotFoundResponse(w http.ResponseWriter) error {
	return WriteContentfulManagementErrorResponse(w, http.StatusNotFound, "NotFound", pointerTo("The resource could not be found."), nil)
}
