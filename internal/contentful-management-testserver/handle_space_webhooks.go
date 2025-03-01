//nolint:dupl
package contentfulmanagementtestserver

import (
	"fmt"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) HandleWebhookDefinitionCreation(spaceID string, webhookDefinition *cm.WebhookDefinition) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.serveMux.Handle(fmt.Sprintf("/spaces/%s/webhook_definitions", spaceID), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			_ = WriteContentfulManagementResponse(w, http.StatusCreated, webhookDefinition)
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}

func (ts *ContentfulManagementTestServer) HandleWebhookDefinition(spaceID string, webhookDefinition *cm.WebhookDefinition) {
	webhookDefinitionID := webhookDefinition.Sys.ID

	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.serveMux.Handle(fmt.Sprintf("/spaces/%s/webhook_definitions/%s", spaceID, webhookDefinitionID), http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, webhookDefinition)
		case http.MethodPut:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, webhookDefinition)
		case http.MethodDelete:
			responseWriter.WriteHeader(http.StatusNoContent)
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))
}
