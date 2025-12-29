package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreateAppDefinition(_ context.Context, req *cm.AppDefinitionData, params cm.CreateAppDefinitionParams) (cm.CreateAppDefinitionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appDefinitionID := generateResourceID()
	appDefinition := NewAppDefinitionFromFields(params.OrganizationID, appDefinitionID, *req)
	ts.appDefinitions[appDefinition.Sys.ID] = &appDefinition

	return &cm.AppDefinitionStatusCode{
		StatusCode: http.StatusCreated,
		Response:   appDefinition,
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetAppDefinition(_ context.Context, params cm.GetAppDefinitionParams) (cm.GetAppDefinitionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appDefinition := ts.appDefinitions[params.AppDefinitionID]
	if appDefinition == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("AppDefinition not found"), nil), nil
	}

	return appDefinition, nil
}

//nolint:ireturn
func (ts *Handler) PutAppDefinition(_ context.Context, req *cm.AppDefinitionData, params cm.PutAppDefinitionParams) (cm.PutAppDefinitionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appDefinition := ts.appDefinitions[params.AppDefinitionID]
	if appDefinition == nil {
		newAppDefinition := NewAppDefinitionFromFields(params.OrganizationID, params.AppDefinitionID, *req)
		ts.appDefinitions[newAppDefinition.Sys.ID] = &newAppDefinition

		return &cm.AppDefinitionStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newAppDefinition,
		}, nil
	}

	UpdateAppDefinitionFromFields(appDefinition, params.OrganizationID, params.AppDefinitionID, *req)

	return &cm.AppDefinitionStatusCode{
		StatusCode: http.StatusOK,
		Response:   *appDefinition,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteAppDefinition(_ context.Context, params cm.DeleteAppDefinitionParams) (cm.DeleteAppDefinitionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appDefinition := ts.appDefinitions[params.AppDefinitionID]
	if appDefinition == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("AppDefinition not found"), nil), nil
	}

	delete(ts.appDefinitions, params.AppDefinitionID)

	return &cm.NoContent{}, nil
}
