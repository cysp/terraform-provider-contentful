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

			apiKeyID := generateResourceID()
			apiKey := NewAPIKeyFromRequestFields(spaceID, apiKeyID, apiKeyRequestFields)
			apiKey.AccessToken = generateResourceID()

			previewAPIKeyID := generateResourceID()
			previewAPIKey := cm.PreviewApiKey{
				Sys:         NewPreviewAPIKeySys(spaceID, previewAPIKeyID),
				AccessToken: generateResourceID(),
			}

			apiKey.PreviewAPIKey.SetTo(cm.ApiKeyPreviewAPIKey{
				Sys: cm.ApiKeyPreviewAPIKeySys{
					Type:     cm.ApiKeyPreviewAPIKeySysTypeLink,
					LinkType: cm.ApiKeyPreviewAPIKeySysLinkTypePreviewApiKey,
					ID:       previewAPIKeyID,
				},
			})

			ts.apiKeys.Set(spaceID, apiKeyID, &apiKey)

			ts.previewAPIKeys.Set(spaceID, previewAPIKeyID, &previewAPIKey)

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

		apiKey := ts.apiKeys.Get(spaceID, apiKeyID)

		switch r.Method {
		case http.MethodGet:
			switch apiKey {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				_ = WriteContentfulManagementResponse(w, http.StatusOK, apiKey)
			}

		case http.MethodPut:
			var apiKeyRequestFields cm.ApiKeyRequestFields
			if err := ReadContentfulManagementRequest(r, &apiKeyRequestFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			switch apiKey {
			case nil:
				apiKey := NewAPIKeyFromRequestFields(spaceID, apiKeyID, apiKeyRequestFields)
				ts.apiKeys.Set(spaceID, apiKeyID, &apiKey)
				_ = WriteContentfulManagementResponse(w, http.StatusCreated, &apiKey)
			default:
				UpdateAPIKeyFromRequestFields(apiKey, apiKeyRequestFields)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, apiKey)
			}

		case http.MethodDelete:
			switch apiKey {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				ts.apiKeys.Delete(spaceID, apiKeyID)
				w.WriteHeader(http.StatusNoContent)
			}

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}
