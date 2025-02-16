package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

type ContentfulManagementErrorHandler struct {
	StatusCode int
	ID         string
	Message    *string
	Details    []byte
}

func NewContentfulManagementErrorHandler(statusCode int, id string, message *string, details []byte) *ContentfulManagementErrorHandler {
	return &ContentfulManagementErrorHandler{
		StatusCode: statusCode,
		ID:         id,
		Message:    message,
		Details:    details,
	}
}

func (e ContentfulManagementErrorHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	_ = WriteContentfulManagementResponse(w, e.StatusCode, &cm.Error{
		Sys: cm.ErrorSys{
			Type: cm.ErrorSysTypeError,
			ID:   e.ID,
		},
		Message: cm.NewOptPointerString(e.Message),
		Details: e.Details,
	})
}
