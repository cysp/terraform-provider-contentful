package testserver

import (
	"fmt"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) HandlePreviewAPIKey(spaceID string, previewAPIKey *cm.PreviewApiKey) {
	previewAPIKeyID := previewAPIKey.Sys.ID

	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.ServeMux.Handle(fmt.Sprintf("/spaces/%s/preview_api_keys/%s", spaceID, previewAPIKeyID), http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, previewAPIKey)
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))
}
