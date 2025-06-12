package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) setupOrganizationAppDefinitionHandlers() {
	ts.serveMux.Handle("/organizations/{organizationID}/app_definitions", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		organizationID := r.PathValue("organizationID")

		if organizationID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		switch r.Method {
		case http.MethodPost:
			var appDefinitionRequest cm.AppDefinitionFields
			if err := ReadContentfulManagementRequest(r, &appDefinitionRequest); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			appDefinitionID := generateResourceID()

			appDefinition := NewAppDefinitionFromFields(organizationID, appDefinitionID, appDefinitionRequest)
			ts.appDefinitions.Set(organizationID, appDefinition.Sys.ID, &appDefinition)

			_ = WriteContentfulManagementResponse(w, http.StatusCreated, &appDefinition)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))

	ts.serveMux.Handle("/organizations/{organizationID}/app_definitions/{appDefinitionID}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		organizationID := r.PathValue("organizationID")
		appDefinitionID := r.PathValue("appDefinitionID")

		if organizationID == NonexistentID || appDefinitionID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		appDefinition := ts.appDefinitions.Get(organizationID, appDefinitionID)

		switch r.Method {
		case http.MethodGet:
			switch appDefinition {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				_ = WriteContentfulManagementResponse(w, http.StatusOK, appDefinition)
			}

		case http.MethodPut:
			var appDefinitionRequest cm.AppDefinitionFields
			if err := ReadContentfulManagementRequest(r, &appDefinitionRequest); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			switch appDefinition {
			case nil:
				appDefinition := NewAppDefinitionFromFields(organizationID, appDefinitionID, appDefinitionRequest)
				ts.appDefinitions.Set(organizationID, appDefinition.Sys.ID, &appDefinition)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, &appDefinition)
			default:
				UpdateAppDefinitionFromFields(appDefinition, organizationID, appDefinitionID, appDefinitionRequest)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, appDefinition)
			}

		case http.MethodDelete:
			switch appDefinition {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				ts.appDefinitions.Delete(organizationID, appDefinitionID)
				w.WriteHeader(http.StatusNoContent)
			}

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}

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

func (ts *ContentfulManagementTestServer) setupOrganizationAppDefinitionResourceTypeHandlers() {
	ts.serveMux.Handle("/organizations/{organizationID}/app_definitions/{appDefinitionID}/resource_provider/resource_types/{resourceTypeID}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		organizationID := r.PathValue("organizationID")
		appDefinitionID := r.PathValue("appDefinitionID")
		resourceTypeID := r.PathValue("resourceTypeID")

		if organizationID == NonexistentID || appDefinitionID == NonexistentID || resourceTypeID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		appDefinitionResourceProvider := ts.appDefinitionResourceProviders.Get(organizationID, appDefinitionID)
		if appDefinitionResourceProvider == nil {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		resourceProviderID := appDefinitionResourceProvider.Sys.ID
		appDefinitionResourceType := ts.appDefinitionResourceTypes.Get(organizationID, resourceTypeID)

		switch r.Method {
		case http.MethodGet:
			switch appDefinitionResourceType {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				_ = WriteContentfulManagementResponse(w, http.StatusOK, appDefinitionResourceType)
			}

		case http.MethodPut:
			var appDefinitionResourceTypeFields cm.ResourceTypeFields
			if err := ReadContentfulManagementRequest(r, &appDefinitionResourceTypeFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			switch appDefinitionResourceType {
			case nil:
				appDefinitionResourceType := NewAppDefinitionResourceTypeFromRequest(organizationID, appDefinitionID, resourceProviderID, resourceTypeID, appDefinitionResourceTypeFields)
				ts.appDefinitionResourceTypes.Set(organizationID, resourceTypeID, &appDefinitionResourceType)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, &appDefinitionResourceType)
			default:
				UpdateAppDefinitionResourceTypeFromFields(appDefinitionResourceType, organizationID, appDefinitionID, resourceProviderID, resourceTypeID, appDefinitionResourceTypeFields)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, appDefinitionResourceType)
			}

		case http.MethodDelete:
			switch appDefinitionResourceType {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				ts.appDefinitionResourceTypes.Delete(organizationID, resourceTypeID)
				w.WriteHeader(http.StatusNoContent)
			}

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}

func (ts *ContentfulManagementTestServer) SetAppDefinition(organizationID, appDefinitionID string, appDefinitionFields cm.AppDefinitionFields) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appDefinition := NewAppDefinitionFromFields(organizationID, appDefinitionID, appDefinitionFields)
	ts.appDefinitions.Set(organizationID, appDefinitionID, &appDefinition)
}

func (ts *ContentfulManagementTestServer) SetAppDefinitionResourceProvider(organizationID, appDefinitionID string, resourceProviderRequest cm.ResourceProviderRequest) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	resourceProvider := NewAppDefinitionResourceProviderFromRequest(organizationID, appDefinitionID, resourceProviderRequest)
	ts.appDefinitionResourceProviders.Set(organizationID, appDefinitionID, &resourceProvider)
}
