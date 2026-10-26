//nolint:dupl
package cmtesting

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetResourceProvider(_ context.Context, params cm.GetResourceProviderParams) (cm.GetResourceProviderRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appDefinitionResourceProvider := ts.appDefinitionResourceProviders[params.AppDefinitionID]
	if appDefinitionResourceProvider == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("ResourceProvider not found"), nil)), nil
	}

	return snapshotResponse(appDefinitionResourceProvider), nil
}

//nolint:ireturn
func (ts *Handler) PutResourceProvider(_ context.Context, req *cm.ResourceProviderRequest, params cm.PutResourceProviderParams) (cm.PutResourceProviderRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appDefinition := ts.appDefinitions[params.AppDefinitionID]
	if appDefinition == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("AppDefinition not found"), nil)), nil
	}

	appDefinitionResourceProvider := ts.appDefinitionResourceProviders[params.AppDefinitionID]
	if appDefinitionResourceProvider == nil {
		appDefinitionResourceProvider := NewResourceProviderFromRequest(params.OrganizationID, params.AppDefinitionID, *req)
		ts.appDefinitionResourceProviders[params.AppDefinitionID] = &appDefinitionResourceProvider

		return snapshotResponse(&cm.ResourceProviderStatusCode{
			StatusCode: http.StatusCreated,
			Response:   appDefinitionResourceProvider,
		}), nil
	}

	UpdateResourceProviderFromRequest(appDefinitionResourceProvider, params.OrganizationID, params.AppDefinitionID, *req)

	return snapshotResponse(&cm.ResourceProviderStatusCode{
		StatusCode: http.StatusOK,
		Response:   *appDefinitionResourceProvider,
	}), nil
}

//nolint:ireturn
func (ts *Handler) DeleteResourceProvider(_ context.Context, params cm.DeleteResourceProviderParams) (cm.DeleteResourceProviderRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appDefinitionResourceProvider := ts.appDefinitionResourceProviders[params.AppDefinitionID]
	if appDefinitionResourceProvider == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("ResourceProvider not found"), nil)), nil
	}

	delete(ts.appDefinitionResourceProviders, params.AppDefinitionID)

	return snapshotResponse(&cm.NoContent{}), nil
}
