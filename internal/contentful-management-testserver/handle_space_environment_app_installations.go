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

		switch r.Method {
		case http.MethodGet:
			appInstallation, found := ts.appInstallations.Get(spaceID, environmentID, appDefinitionID)
			if !found {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, appInstallation)

		case http.MethodPut:
			if appDefinitionID == "nonexistent" {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			appInstallation := &cm.AppInstallation{}
			_ = ReadContentfulManagementRequest(r, appInstallation)

			appInstallation.Sys = cm.AppInstallationSys{
				Type: cm.AppInstallationSysTypeAppInstallation,
			}

			ts.appInstallations.Set(spaceID, environmentID, appDefinitionID, appInstallation)

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, appInstallation)

		case http.MethodDelete:
			ts.appInstallations.Delete(spaceID, environmentID, appDefinitionID)
			responseWriter.WriteHeader(http.StatusNoContent)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))
}
