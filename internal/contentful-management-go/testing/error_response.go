package testing

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func WriteContentfulManagementErrorResponse(w http.ResponseWriter, statusCode int, id string, message *string, details []byte) error {
	return WriteContentfulManagementResponse(w, statusCode, &cm.Error{
		Sys:     cm.NewErrorSys(id),
		Message: cm.NewOptPointerString(message),
		Details: details,
	})
}
