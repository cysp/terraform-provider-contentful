package contentfulmanagementtestserver

import (
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

			apiKeyID := ts.generateResourceID()

			apiKey.Sys = cm.ApiKeySys{
				Type: cm.ApiKeySysTypeApiKey,
				ID:   apiKeyID,
			}

			apiKey.AccessToken = apiKeyID

			ts.SetAPIKey(spaceID, &apiKey)

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
			_ = ReadContentfulManagementRequest(r, &apiKey)

			if existingAPIKey, found := ts.apiKeys.Get(spaceID, apiKeyID); found {
				existingAPIKey.Name = apiKey.Name
				existingAPIKey.Description = apiKey.Description
				existingAPIKey.Environments = apiKey.Environments

				_ = WriteContentfulManagementResponse(w, http.StatusOK, existingAPIKey)
			} else {
				apiKey.Sys = cm.ApiKeySys{
					Type: cm.ApiKeySysTypeApiKey,
					ID:   apiKeyID,
				}

				apiKey.AccessToken = apiKeyID

				ts.apiKeys.Set(spaceID, apiKeyID, &apiKey)

				_ = WriteContentfulManagementResponse(w, http.StatusCreated, &apiKey)
			}

		case http.MethodDelete:
			ts.apiKeys.Delete(spaceID, apiKeyID)
			w.WriteHeader(http.StatusNoContent)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}

func (ts *ContentfulManagementTestServer) GetAPIKey(spaceID, apiKeyID string) (*cm.ApiKey, bool) {
	return ts.apiKeys.Get(spaceID, apiKeyID)
}

func (ts *ContentfulManagementTestServer) SetAPIKey(spaceID string, apiKey *cm.ApiKey) {
	var previewAPIKeyID string
	if previewAPIKey, ok := apiKey.PreviewAPIKey.Get(); ok {
		previewAPIKeyID = previewAPIKey.Sys.ID
	}

	if previewAPIKeyID == "" {
		previewAPIKeyID = ts.generateResourceID()
	}

	ts.apiKeys.Set(spaceID, apiKey.Sys.ID, apiKey)

	previewAPIKey := cm.PreviewApiKey{
		Sys: cm.PreviewApiKeySys{
			Type: cm.PreviewApiKeySysTypePreviewApiKey,
			ID:   previewAPIKeyID,
		},
		AccessToken: previewAPIKeyID,
	}

	apiKey.PreviewAPIKey.SetTo(cm.ApiKeyPreviewAPIKey{
		Sys: cm.ApiKeyPreviewAPIKeySys{
			Type:     cm.ApiKeyPreviewAPIKeySysTypeLink,
			LinkType: cm.ApiKeyPreviewAPIKeySysLinkTypePreviewApiKey,
			ID:       previewAPIKeyID,
		},
	})

	ts.previewAPIKeys.Set(spaceID, previewAPIKeyID, &previewAPIKey)
}

func (ts *ContentfulManagementTestServer) DeleteApiKey(spaceID, apiKeyID string) {
	ts.apiKeys.Delete(spaceID, apiKeyID)
}
