package contentfulmanagementtestserver

import (
	"net/http"
)

func (ts *ContentfulManagementTestServer) setupSpacePreviewAPIKeyHandlers() {
	ts.serveMux.Handle("/spaces/{spaceID}/preview_api_keys/{previewAPIKeyID}", http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		previewAPIKeyID := r.PathValue("previewAPIKeyID")

		switch r.Method {
		case http.MethodGet:
			previewAPIKey, found := ts.previewAPIKeys.Get(spaceID, previewAPIKeyID)
			if !found {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, previewAPIKey)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))
}
