package contentfulmanagementtestserver

import (
	"encoding/json"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) setupSpaceAPIKeyHandlers() {
	ts.serveMux.Handle("/spaces/{spaceID}/api_keys", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")

		switch r.Method {
		case http.MethodPost:
			var apiKey cm.ApiKey
			_ = ReadContentfulManagementRequest(r, &apiKey)

			apiKey.Sys = cm.ApiKeySys{
				Type: cm.ApiKeySysTypeApiKey,
				ID:   ts.generateResourceID(),
			}

			ts.apiKeys.Set(spaceID, apiKey.Sys.ID, &apiKey)

			_ = WriteContentfulManagementResponse(w, http.StatusCreated, &apiKey)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))

	ts.serveMux.Handle("/spaces/{spaceID}/api_keys/{apiKeyID}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		apiKeyID := r.PathValue("apiKeyID")

		switch r.Method {
		case http.MethodGet:
			apiKey, found := ts.apiKeys.Get(spaceID, apiKeyID)
			if !found {
				_ = WriteContentfulManagementErrorNotFoundResponse(w)

				return
			}

			_ = WriteContentfulManagementResponse(w, http.StatusOK, apiKey)

		case http.MethodPut:
			var apiKey cm.ApiKey
			if err := json.NewDecoder(r.Body).Decode(&apiKey); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponse(w)

				return
			}

			ts.apiKeys.Set(spaceID, apiKeyID, &apiKey)

			_ = WriteContentfulManagementResponse(w, http.StatusOK, &apiKey)

		case http.MethodDelete:
			ts.apiKeys.Delete(spaceID, apiKeyID)
			w.WriteHeader(http.StatusNoContent)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}
