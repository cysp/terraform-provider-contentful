package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) setupSpaceAPIKeyHandlers() {
	ts.serveMux.Handle("/spaces/{spaceID}/api_keys", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")

		if spaceID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		switch r.Method {
		case http.MethodPost:
			var apiKeyRequestFields cm.ApiKeyRequestFields
			if err := ReadContentfulManagementRequest(r, &apiKeyRequestFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			apiKeyID := ts.generateResourceID()
			apiKey := NewAPIKeyFromRequestFields(spaceID, apiKeyID, apiKeyRequestFields)

			previewAPIKey := cm.PreviewApiKey{
				Sys:         NewPreviewAPIKeySys(spaceID, apiKeyID),
				AccessToken: apiKeyID,
			}

			apiKey.PreviewAPIKey.SetTo(cm.ApiKeyPreviewAPIKey{
				Sys: cm.ApiKeyPreviewAPIKeySys{
					Type:     cm.ApiKeyPreviewAPIKeySysTypeLink,
					LinkType: cm.ApiKeyPreviewAPIKeySysLinkTypePreviewApiKey,
					ID:       apiKeyID,
				},
			})

			apiKey.AccessToken = apiKeyID

			ts.apiKeys.Set(spaceID, apiKeyID, &apiKey)

			ts.previewAPIKeys.Set(spaceID, apiKeyID, &previewAPIKey)

			_ = WriteContentfulManagementResponse(w, http.StatusCreated, &apiKey)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))

	ts.serveMux.Handle("/spaces/{spaceID}/api_keys/{apiKeyID}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		apiKeyID := r.PathValue("apiKeyID")

		if spaceID == NonexistentID || apiKeyID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		apiKey, exists := ts.apiKeys.Get(spaceID, apiKeyID)
		if !exists {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		switch r.Method {
		case http.MethodGet:
			_ = WriteContentfulManagementResponse(w, http.StatusOK, apiKey)

		case http.MethodPut:
			var apiKeyRequestFields cm.ApiKeyRequestFields
			if err := ReadContentfulManagementRequest(r, &apiKeyRequestFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			if apiKey == nil {
				apiKey := NewAPIKeyFromRequestFields(spaceID, apiKeyID, apiKeyRequestFields)

				ts.apiKeys.Set(spaceID, apiKeyID, &apiKey)

				_ = WriteContentfulManagementResponse(w, http.StatusCreated, &apiKey)
			} else {
				UpdateAPIKeyFromRequestFields(apiKey, apiKeyRequestFields)

				_ = WriteContentfulManagementResponse(w, http.StatusOK, apiKey)
			}

		case http.MethodDelete:
			ts.apiKeys.Delete(spaceID, apiKeyID)

			w.WriteHeader(http.StatusNoContent)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}

// func (ts *ContentfulManagementTestServer) GetAPIKey(spaceID, apiKeyID string) (*cm.ApiKey, bool) {
// 	return ts.apiKeys.Get(spaceID, apiKeyID)
// }

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

func (ts *ContentfulManagementTestServer) DeleteAPIKey(spaceID, apiKeyID string) {
	ts.apiKeys.Delete(spaceID, apiKeyID)
}
