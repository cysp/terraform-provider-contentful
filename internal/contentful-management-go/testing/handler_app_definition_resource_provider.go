//nolint:dupl
package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetAppDefinitionResourceProvider(_ context.Context, params cm.GetAppDefinitionResourceProviderParams) (cm.GetAppDefinitionResourceProviderRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.OrganizationID == NonexistentID || params.AppDefinitionID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	appDefinitionResourceProvider := ts.appDefinitionResourceProviders.Get(params.OrganizationID, params.AppDefinitionID)
	if appDefinitionResourceProvider == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("AppDefinitionResourceProvider not found"), nil), nil
	}

	return appDefinitionResourceProvider, nil
}

//nolint:ireturn
func (ts *Handler) PutAppDefinitionResourceProvider(_ context.Context, req *cm.ResourceProviderRequest, params cm.PutAppDefinitionResourceProviderParams) (cm.PutAppDefinitionResourceProviderRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.OrganizationID == NonexistentID || params.AppDefinitionID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	appDefinitionResourceProvider := ts.appDefinitionResourceProviders.Get(params.OrganizationID, params.AppDefinitionID)
	if appDefinitionResourceProvider == nil {
		appDefinitionResourceProvider := NewAppDefinitionResourceProviderFromRequest(params.OrganizationID, params.AppDefinitionID, *req)

		ts.appDefinitionResourceProviders.Set(params.OrganizationID, params.AppDefinitionID, &appDefinitionResourceProvider)

		return &cm.ResourceProviderStatusCode{
			StatusCode: http.StatusCreated,
			Response:   appDefinitionResourceProvider,
		}, nil
	}

	UpdateAppDefinitionResourceProviderFromRequest(appDefinitionResourceProvider, params.OrganizationID, params.AppDefinitionID, *req)

	return &cm.ResourceProviderStatusCode{
		StatusCode: http.StatusOK,
		Response:   *appDefinitionResourceProvider,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteAppDefinitionResourceProvider(_ context.Context, params cm.DeleteAppDefinitionResourceProviderParams) (cm.DeleteAppDefinitionResourceProviderRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.OrganizationID == NonexistentID || params.AppDefinitionID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	appDefinitionResourceProvider := ts.appDefinitionResourceProviders.Get(params.OrganizationID, params.AppDefinitionID)
	if appDefinitionResourceProvider == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("AppDefinitionResourceProvider not found"), nil), nil
	}

	ts.appDefinitionResourceProviders.Delete(params.OrganizationID, params.AppDefinitionID)

	return &cm.NoContent{}, nil
}
