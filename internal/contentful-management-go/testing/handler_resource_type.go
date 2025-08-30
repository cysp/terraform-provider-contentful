package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetResourceType(_ context.Context, params cm.GetResourceTypeParams) (cm.GetResourceTypeRes, error) {
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
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("ResourceType not found"), nil), nil
	}

	return appDefinitionResourceType, nil
}

//nolint:ireturn
func (ts *Handler) PutResourceType(_ context.Context, req *cm.ResourceTypeFields, params cm.PutResourceTypeParams) (cm.PutResourceTypeRes, error) {
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
		newResourceType := NewResourceTypeFromRequest(params.OrganizationID, params.AppDefinitionID, resourceProviderID, params.ResourceTypeID, *req)
		ts.appDefinitionResourceTypes.Set(params.OrganizationID, params.ResourceTypeID, &newResourceType)

		return &cm.ResourceTypeStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newResourceType,
		}, nil
	}

	UpdateResourceTypeFromFields(appDefinitionResourceType, params.OrganizationID, params.AppDefinitionID, resourceProviderID, params.ResourceTypeID, *req)

	return &cm.ResourceTypeStatusCode{
		StatusCode: http.StatusOK,
		Response:   *appDefinitionResourceType,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteResourceType(_ context.Context, params cm.DeleteResourceTypeParams) (cm.DeleteResourceTypeRes, error) {
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
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("ResourceType not found"), nil), nil
	}

	ts.appDefinitionResourceTypes.Delete(params.OrganizationID, params.ResourceTypeID)

	return &cm.NoContent{}, nil
}
