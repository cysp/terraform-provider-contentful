package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) setupSpaceWebhookDefinitionHandlers() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.serveMux.Handle("/spaces/{spaceID}/webhook_definitions", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")

		if spaceID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		switch r.Method {
		case http.MethodPost:
			var webhookDefinitionFields cm.WebhookDefinitionFields
			if err := ReadContentfulManagementRequest(r, &webhookDefinitionFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			webhookDefinition := NewWebhookDefinitionFromFields(spaceID, ts.generateResourceID(), webhookDefinitionFields)

			ts.webhookDefinitions.Set(spaceID, webhookDefinition.Sys.ID, &webhookDefinition)

			_ = WriteContentfulManagementResponse(w, http.StatusCreated, &webhookDefinition)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))

	ts.serveMux.Handle("/spaces/{spaceID}/webhook_definitions/{webhookDefinitionID}", http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		webhookDefinitionID := r.PathValue("webhookDefinitionID")

		if spaceID == NonexistentID || webhookDefinitionID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		webhookDefinition, exists := ts.webhookDefinitions.Get(spaceID, webhookDefinitionID)

		switch r.Method {
		case http.MethodGet:
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, webhookDefinition)

		case http.MethodPut:
			var webhookDefinitionFields cm.WebhookDefinitionFields
			if err := ReadContentfulManagementRequest(r, &webhookDefinitionFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(responseWriter, err)

				return
			}

			if exists {
				UpdateWebhookDefinitionFromFields(webhookDefinition, webhookDefinitionFields)

				_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, webhookDefinition)
			} else {
				webhookDefinition := NewWebhookDefinitionFromFields(spaceID, webhookDefinitionID, webhookDefinitionFields)

				ts.webhookDefinitions.Set(spaceID, webhookDefinition.Sys.ID, &webhookDefinition)

				_ = WriteContentfulManagementResponse(responseWriter, http.StatusCreated, &webhookDefinition)
			}

		case http.MethodDelete:
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			ts.webhookDefinitions.Delete(spaceID, webhookDefinitionID)

			responseWriter.WriteHeader(http.StatusNoContent)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))
}

func (ts *ContentfulManagementTestServer) GetWebhookDefinition(spaceID, webhookDefinitionID string) (*cm.WebhookDefinition, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	return ts.webhookDefinitions.Get(spaceID, webhookDefinitionID)
}

func (ts *ContentfulManagementTestServer) SetWebhookDefinition(spaceID string, webhookDefinition *cm.WebhookDefinition) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.webhookDefinitions.Set(spaceID, webhookDefinition.Sys.ID, webhookDefinition)
}

func (ts *ContentfulManagementTestServer) DeleteWebhookDefinition(spaceID, webhookDefinitionID string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.webhookDefinitions.Delete(spaceID, webhookDefinitionID)
}
