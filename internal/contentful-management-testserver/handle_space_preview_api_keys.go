package contentfulmanagementtestserver

import (
	"net/http"
)

func (ts *ContentfulManagementTestServer) setupSpacePreviewAPIKeyHandlers() {
	ts.serveMux.Handle("/spaces/{spaceID}/preview_api_keys/{previewAPIKeyID}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		previewAPIKeyID := r.PathValue("previewAPIKeyID")

		if spaceID == NonexistentID || previewAPIKeyID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		previewAPIKey := ts.previewAPIKeys.Get(spaceID, previewAPIKeyID)

		switch r.Method {
		case http.MethodGet:
			switch previewAPIKey {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				_ = WriteContentfulManagementResponse(w, http.StatusOK, previewAPIKey)
			}

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}
