package contentfulmanagementtestserver

import (
	"fmt"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) HandleAppInstallation(spaceID string, environmentID string, appDefinitionID string, appInstallation *cm.AppInstallation) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.serveMux.Handle(fmt.Sprintf("/spaces/%s/environments/%s/app_installations/%s", spaceID, environmentID, appDefinitionID), http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, appInstallation)
		case http.MethodPut:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, appInstallation)
		case http.MethodDelete:
			responseWriter.WriteHeader(http.StatusNoContent)
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))
}
