package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) SetupSpaceEnvironmentAppInstallationHandlers() {
	ts.serveMux.Handle("/spaces/{spaceID}/environments/{environmentID}/app_installations/{appDefinitionID}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		environmentID := r.PathValue("environmentID")
		appDefinitionID := r.PathValue("appDefinitionID")

		if spaceID == NonexistentID || environmentID == NonexistentID || appDefinitionID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.RLock()
		defer ts.mu.RUnlock()

		appInstallation := ts.appInstallations.Get(spaceID, environmentID, appDefinitionID)

		switch r.Method {
		case http.MethodGet:
			switch appInstallation {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				_ = WriteContentfulManagementResponse(w, http.StatusOK, appInstallation)
			}

		case http.MethodPut:
			var appInstallationFields cm.AppInstallationFields
			if err := ReadContentfulManagementRequest(r, &appInstallationFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			switch appInstallation {
			case nil:
				appInstallation := NewAppInstallationFromFields(spaceID, environmentID, appDefinitionID, appInstallationFields)
				ts.appInstallations.Set(spaceID, environmentID, appDefinitionID, &appInstallation)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, &appInstallation)
			default:
				UpdateAppInstallationFromFields(appInstallation, appInstallationFields)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, appInstallation)
			}

		case http.MethodDelete:
			switch appInstallation {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				ts.appInstallations.Delete(spaceID, environmentID, appDefinitionID)
				w.WriteHeader(http.StatusNoContent)
			}

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}

func (ts *ContentfulManagementTestServer) AddAppDefinitionID(appDefinitionID string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.appDefinitionIDs[appDefinitionID] = struct{}{}
}
