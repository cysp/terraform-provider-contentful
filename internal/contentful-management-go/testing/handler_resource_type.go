package cmtesting

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

type resourceTypeIdentity struct {
	organizationID  string
	appDefinitionID string
	resourceTypeID  string
}

//nolint:ireturn
func (ts *Handler) GetResourceType(_ context.Context, params cm.GetResourceTypeParams) (cm.GetResourceTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appDefinitionResourceProvider := ts.resourceProvider(params.OrganizationID, params.AppDefinitionID)
	if appDefinitionResourceProvider == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("ResourceProvider not found"), nil)), nil
	}

	identity := resourceTypeIdentity{params.OrganizationID, params.AppDefinitionID, params.ResourceTypeID}

	appDefinitionResourceType := ts.appDefinitionResourceTypes[identity]
	if appDefinitionResourceType == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("ResourceType not found"), nil)), nil
	}

	return snapshotResponse(appDefinitionResourceType), nil
}

//nolint:ireturn
func (ts *Handler) PutResourceType(_ context.Context, req *cm.ResourceTypeData, params cm.PutResourceTypeParams) (cm.PutResourceTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appDefinitionResourceProvider := ts.resourceProvider(params.OrganizationID, params.AppDefinitionID)
	if appDefinitionResourceProvider == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("ResourceProvider not found"), nil)), nil
	}

	resourceProviderID := appDefinitionResourceProvider.Sys.ID

	identity := resourceTypeIdentity{params.OrganizationID, params.AppDefinitionID, params.ResourceTypeID}

	appDefinitionResourceType := ts.appDefinitionResourceTypes[identity]

	if appDefinitionResourceType == nil {
		newResourceType := NewResourceTypeFromRequest(params.OrganizationID, params.AppDefinitionID, resourceProviderID, params.ResourceTypeID, *req)
		ts.appDefinitionResourceTypes[identity] = &newResourceType

		return snapshotResponse(&cm.ResourceTypeStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newResourceType,
		}), nil
	}

	UpdateResourceTypeFromFields(appDefinitionResourceType, params.OrganizationID, params.AppDefinitionID, resourceProviderID, params.ResourceTypeID, *req)

	return snapshotResponse(&cm.ResourceTypeStatusCode{
		StatusCode: http.StatusOK,
		Response:   *appDefinitionResourceType,
	}), nil
}

//nolint:ireturn
func (ts *Handler) DeleteResourceType(_ context.Context, params cm.DeleteResourceTypeParams) (cm.DeleteResourceTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appDefinitionResourceProvider := ts.resourceProvider(params.OrganizationID, params.AppDefinitionID)
	if appDefinitionResourceProvider == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("ResourceProvider not found"), nil)), nil
	}

	identity := resourceTypeIdentity{params.OrganizationID, params.AppDefinitionID, params.ResourceTypeID}

	appDefinitionResourceType := ts.appDefinitionResourceTypes[identity]
	if appDefinitionResourceType == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("ResourceType not found"), nil)), nil
	}

	delete(ts.appDefinitionResourceTypes, identity)

	return snapshotResponse(&cm.NoContent{}), nil
}
