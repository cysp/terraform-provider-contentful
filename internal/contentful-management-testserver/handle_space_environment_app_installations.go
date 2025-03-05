package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) SetupSpaceEnvironmentAppInstallationHandlers() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.appInstallations = NewSpaceEnvironmentMap[*cm.AppInstallation]()

	ts.serveMux.Handle("/spaces/{spaceID}/environments/{environmentID}/app_installations/{appDefinitionID}", http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		environmentID := r.PathValue("environmentID")
		appDefinitionID := r.PathValue("appDefinitionID")

		if spaceID == NonexistentID || environmentID == NonexistentID || appDefinitionID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		appInstallation, exists := ts.appInstallations.Get(spaceID, environmentID, appDefinitionID)

		switch r.Method {
		case http.MethodGet:
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, appInstallation)

		case http.MethodPut:
			appInstallationFields := cm.AppInstallationFields{}
			if err := ReadContentfulManagementRequest(r, &appInstallationFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(responseWriter, err)

				return
			}

			if exists {
				UpdateAppInstallationFromFields(appInstallation, appInstallationFields)

				_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, appInstallation)
			} else {
				appInstallation := NewAppInstallationFromFields(spaceID, environmentID, appDefinitionID, appInstallationFields)

				ts.appInstallations.Set(spaceID, environmentID, appDefinitionID, &appInstallation)

				_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, &appInstallation)
			}

		case http.MethodDelete:
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			ts.appInstallations.Delete(spaceID, environmentID, appDefinitionID)

			responseWriter.WriteHeader(http.StatusNoContent)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))
}
