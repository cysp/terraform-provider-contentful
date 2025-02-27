//nolint:dupl
package testserver

import (
	"fmt"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) HandleAPIKeyCreation(spaceID string, apiKey *cm.ApiKey) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.ServeMux.Handle(fmt.Sprintf("/spaces/%s/api_keys", spaceID), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			_ = WriteContentfulManagementResponse(w, http.StatusCreated, apiKey)
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}

func (ts *ContentfulManagementTestServer) HandleAPIKey(spaceID string, apiKey *cm.ApiKey) {
	apiKeyID := apiKey.Sys.ID

	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.ServeMux.Handle(fmt.Sprintf("/spaces/%s/api_keys/%s", spaceID, apiKeyID), http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, apiKey)
		case http.MethodPut:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, apiKey)
		case http.MethodDelete:
			responseWriter.WriteHeader(http.StatusNoContent)
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))
}
