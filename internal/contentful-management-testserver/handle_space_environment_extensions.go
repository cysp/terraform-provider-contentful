//nolint:dupl
package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) SetupSpaceEnvironmentExtensionHandlers() {
	ts.serveMux.Handle("/spaces/{spaceID}/environments/{environmentID}/extensions/{extensionID}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		environmentID := r.PathValue("environmentID")
		extensionID := r.PathValue("extensionID")

		if spaceID == NonexistentID || environmentID == NonexistentID || extensionID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		extension := ts.extensions.Get(spaceID, environmentID, extensionID)

		switch r.Method {
		case http.MethodGet:
			switch extension {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				_ = WriteContentfulManagementResponse(w, http.StatusOK, extension)
			}

		case http.MethodPut:
			var extensionFields cm.ExtensionFields
			if err := ReadContentfulManagementRequest(r, &extensionFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			switch extension {
			case nil:
				appInstallation := NewExtensionFromFields(spaceID, environmentID, extensionID, extensionFields)
				ts.extensions.Set(spaceID, environmentID, extensionID, &appInstallation)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, &appInstallation)
			default:
				UpdateExtensionFromFields(extension, extensionFields)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, extension)
			}

		case http.MethodDelete:
			switch extension {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				ts.extensions.Delete(spaceID, environmentID, extensionID)
				w.WriteHeader(http.StatusNoContent)
			}

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}
