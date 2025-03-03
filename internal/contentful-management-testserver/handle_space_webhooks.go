package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) setupSpaceWebhookDefinitionHandlers() {
	ts.serveMux.Handle("/spaces/{spaceID}/webhook_definitions", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")

		switch r.Method {
		case http.MethodPost:
			var webhookDefinition cm.WebhookDefinition
			_ = ReadContentfulManagementRequest(r, &webhookDefinition)

			webhookDefinition.Sys = cm.WebhookDefinitionSys{
				Type: cm.WebhookDefinitionSysTypeWebhookDefinition,
				ID:   ts.generateResourceID(),
			}

			ts.webhookDefinitions.Set(spaceID, webhookDefinition.Sys.ID, &webhookDefinition)

			_ = WriteContentfulManagementResponse(w, http.StatusCreated, &webhookDefinition)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))

	ts.serveMux.Handle("/spaces/{spaceID}/webhook_definitions/{webhookDefinitionID}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		webhookDefinitionID := r.PathValue("webhookDefinitionID")

		switch r.Method {
		case http.MethodGet:
			webhookDefinition, exists := ts.webhookDefinitions.Get(spaceID, webhookDefinitionID)
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
				return
			}
			_ = WriteContentfulManagementResponse(w, http.StatusOK, webhookDefinition)

		case http.MethodPut:
			var webhookDefinition cm.WebhookDefinition
			_ = ReadContentfulManagementRequest(r, &webhookDefinition)

			webhookDefinition.Sys = cm.WebhookDefinitionSys{
				Type: cm.WebhookDefinitionSysTypeWebhookDefinition,
				ID:   webhookDefinitionID,
			}

			ts.webhookDefinitions.Set(spaceID, webhookDefinition.Sys.ID, &webhookDefinition)
			_ = WriteContentfulManagementResponse(w, http.StatusOK, &webhookDefinition)

		case http.MethodDelete:
			ts.webhookDefinitions.Delete(spaceID, webhookDefinitionID)
			w.WriteHeader(http.StatusNoContent)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}
