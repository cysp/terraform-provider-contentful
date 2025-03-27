package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) setupSpaceEnablementsHandlers() {
	ts.serveMux.Handle("/spaces/{spaceID}/enablements", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")

		if spaceID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		enablements := ts.getOrCreateSpaceEnablements(spaceID)

		switch r.Method {
		case http.MethodGet:
			_ = WriteContentfulManagementResponse(w, http.StatusOK, enablements)

		case http.MethodPut:
			var enablementRequestFields cm.SpaceEnablementFields
			if err := ReadContentfulManagementRequest(r, &enablementRequestFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			UpdateSpaceEnablementFromFields(enablements, enablementRequestFields)

			_ = WriteContentfulManagementResponse(w, http.StatusOK, enablements)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}

func (ts *ContentfulManagementTestServer) getOrCreateSpaceEnablements(spaceID string) *cm.SpaceEnablement {
	enablements, ok := ts.enablements[spaceID]
	if !ok {
		enablements = pointerTo(NewSpaceEnablement(spaceID))
		ts.enablements[spaceID] = enablements
	}

	return enablements
}
