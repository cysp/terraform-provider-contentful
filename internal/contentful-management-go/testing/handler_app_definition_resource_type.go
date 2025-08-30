package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetAppDefinitionResourceType(_ context.Context, params cm.GetAppDefinitionResourceTypeParams) (cm.GetAppDefinitionResourceTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.OrganizationID == NonexistentID || params.AppDefinitionID == NonexistentID || params.ResourceTypeID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	appDefinitionResourceProvider := ts.appDefinitionResourceProviders.Get(params.OrganizationID, params.AppDefinitionID)
	if appDefinitionResourceProvider == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("ResourceProvider not found"), nil), nil
	}

	appDefinitionResourceType := ts.appDefinitionResourceTypes.Get(params.OrganizationID, params.ResourceTypeID)
	if appDefinitionResourceType == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("AppDefinitionResourceType not found"), nil), nil
	}

	return appDefinitionResourceType, nil
}

//nolint:ireturn
func (ts *Handler) PutAppDefinitionResourceType(_ context.Context, req *cm.ResourceTypeFields, params cm.PutAppDefinitionResourceTypeParams) (cm.PutAppDefinitionResourceTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.OrganizationID == NonexistentID || params.AppDefinitionID == NonexistentID || params.ResourceTypeID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	appDefinitionResourceProvider := ts.appDefinitionResourceProviders.Get(params.OrganizationID, params.AppDefinitionID)
	if appDefinitionResourceProvider == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("ResourceProvider not found"), nil), nil
	}

	resourceProviderID := appDefinitionResourceProvider.Sys.ID

	appDefinitionResourceType := ts.appDefinitionResourceTypes.Get(params.OrganizationID, params.ResourceTypeID)

	if appDefinitionResourceType == nil {
		newResourceType := NewAppDefinitionResourceTypeFromRequest(params.OrganizationID, params.AppDefinitionID, resourceProviderID, params.ResourceTypeID, *req)
		ts.appDefinitionResourceTypes.Set(params.OrganizationID, params.ResourceTypeID, &newResourceType)

		return &cm.ResourceTypeStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newResourceType,
		}, nil
	}

	UpdateAppDefinitionResourceTypeFromFields(appDefinitionResourceType, params.OrganizationID, params.AppDefinitionID, resourceProviderID, params.ResourceTypeID, *req)

	return &cm.ResourceTypeStatusCode{
		StatusCode: http.StatusOK,
		Response:   *appDefinitionResourceType,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteAppDefinitionResourceType(_ context.Context, params cm.DeleteAppDefinitionResourceTypeParams) (cm.DeleteAppDefinitionResourceTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.OrganizationID == NonexistentID || params.AppDefinitionID == NonexistentID || params.ResourceTypeID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	appDefinitionResourceProvider := ts.appDefinitionResourceProviders.Get(params.OrganizationID, params.AppDefinitionID)
	if appDefinitionResourceProvider == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("ResourceProvider not found"), nil), nil
	}

	appDefinitionResourceType := ts.appDefinitionResourceTypes.Get(params.OrganizationID, params.ResourceTypeID)
	if appDefinitionResourceType == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("AppDefinitionResourceType not found"), nil), nil
	}

	ts.appDefinitionResourceTypes.Delete(params.OrganizationID, params.ResourceTypeID)

	return &cm.NoContent{}, nil
}
