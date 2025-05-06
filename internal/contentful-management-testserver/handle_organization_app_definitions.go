package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) setupOrganizationAppDefinitionResourceProviderHandlers() {
	ts.serveMux.Handle("/organizations/{organizationID}/app_definitions/{appDefinitionID}/resource_provider", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		organizationID := r.PathValue("organizationID")
		appDefinitionID := r.PathValue("appDefinitionID")

		if organizationID == NonexistentID || appDefinitionID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		appDefinitionResourceProvider := ts.appDefinitionResourceProviders.Get(organizationID, appDefinitionID)

		switch r.Method {
		case http.MethodGet:
			switch appDefinitionResourceProvider {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				_ = WriteContentfulManagementResponse(w, http.StatusOK, appDefinitionResourceProvider)
			}

		case http.MethodPut:
			var appDefinitionResourceProviderRequest cm.ResourceProviderRequest
			if err := ReadContentfulManagementRequest(r, &appDefinitionResourceProviderRequest); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			switch appDefinitionResourceProvider {
			case nil:
				appDefinitionResourceProvider := NewAppDefinitionResourceProviderFromRequest(organizationID, appDefinitionID, appDefinitionResourceProviderRequest)
				ts.appDefinitionResourceProviders.Set(organizationID, appDefinitionID, &appDefinitionResourceProvider)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, &appDefinitionResourceProvider)
			default:
				UpdateAppDefinitionResourceProviderFromRequest(appDefinitionResourceProvider, organizationID, appDefinitionID, appDefinitionResourceProviderRequest)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, appDefinitionResourceProvider)
			}

		case http.MethodDelete:
			switch appDefinitionResourceProvider {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				ts.appDefinitionResourceProviders.Delete(organizationID, appDefinitionID)
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

func (ts *ContentfulManagementTestServer) SetAppDefinitionResourceProvider(organizationID, appDefinitionID string, resourceProviderRequest cm.ResourceProviderRequest) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	resourceProvider := NewAppDefinitionResourceProviderFromRequest(organizationID, appDefinitionID, resourceProviderRequest)
	ts.appDefinitionResourceProviders.Set(organizationID, appDefinitionID, &resourceProvider)
}
