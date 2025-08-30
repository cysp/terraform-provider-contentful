//nolint:dupl
package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetResourceProvider(_ context.Context, params cm.GetResourceProviderParams) (cm.GetResourceProviderRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.OrganizationID == NonexistentID || params.AppDefinitionID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	appDefinitionResourceProvider := ts.appDefinitionResourceProviders.Get(params.OrganizationID, params.AppDefinitionID)
	if appDefinitionResourceProvider == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("ResourceProvider not found"), nil), nil
	}

	return appDefinitionResourceProvider, nil
}

//nolint:ireturn
func (ts *Handler) PutResourceProvider(_ context.Context, req *cm.ResourceProviderRequest, params cm.PutResourceProviderParams) (cm.PutResourceProviderRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.OrganizationID == NonexistentID || params.AppDefinitionID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	appDefinitionResourceProvider := ts.appDefinitionResourceProviders.Get(params.OrganizationID, params.AppDefinitionID)
	if appDefinitionResourceProvider == nil {
		appDefinitionResourceProvider := NewResourceProviderFromRequest(params.OrganizationID, params.AppDefinitionID, *req)

		ts.appDefinitionResourceProviders.Set(params.OrganizationID, params.AppDefinitionID, &appDefinitionResourceProvider)

		return &cm.ResourceProviderStatusCode{
			StatusCode: http.StatusCreated,
			Response:   appDefinitionResourceProvider,
		}, nil
	}

	UpdateResourceProviderFromRequest(appDefinitionResourceProvider, params.OrganizationID, params.AppDefinitionID, *req)

	return &cm.ResourceProviderStatusCode{
		StatusCode: http.StatusOK,
		Response:   *appDefinitionResourceProvider,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteResourceProvider(_ context.Context, params cm.DeleteResourceProviderParams) (cm.DeleteResourceProviderRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.OrganizationID == NonexistentID || params.AppDefinitionID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	appDefinitionResourceProvider := ts.appDefinitionResourceProviders.Get(params.OrganizationID, params.AppDefinitionID)
	if appDefinitionResourceProvider == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("ResourceProvider not found"), nil), nil
	}

	ts.appDefinitionResourceProviders.Delete(params.OrganizationID, params.AppDefinitionID)

	return &cm.NoContent{}, nil
}
