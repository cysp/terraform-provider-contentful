package cmtesting

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetAppInstallation(_ context.Context, params cm.GetAppInstallationParams) (cm.GetAppInstallationRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	appInstallation := ts.appInstallations.Get(params.SpaceID, params.EnvironmentID, params.AppDefinitionID)
	if appInstallation == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, []byte(`"AppInstallation does not exist."`)), nil
	}

	return appInstallationResponse(appInstallation, ts.appDefinitions[params.AppDefinitionID])
}

//nolint:ireturn
func (ts *Handler) PutAppInstallation(_ context.Context, req *cm.AppInstallationData, params cm.PutAppInstallationParams) (cm.PutAppInstallationRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	definition := ts.appDefinitions[params.AppDefinitionID]
	if definition == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("AppDefinition not found"), nil), nil
	}

	appInstallation := ts.appInstallations.Get(params.SpaceID, params.EnvironmentID, params.AppDefinitionID)

	parameters, validation, err := appInstallationParameters(*req, *definition, appInstallation)
	if err != nil {
		return nil, err
	}

	if validation != nil {
		return validation, nil
	}

	fields := cm.AppInstallationData{Parameters: parameters}
	if appInstallation == nil {
		newAppInstallation := NewAppInstallationFromFields(params.SpaceID, params.EnvironmentID, params.AppDefinitionID, fields)
		ts.appInstallations.Set(params.SpaceID, params.EnvironmentID, params.AppDefinitionID, &newAppInstallation)

		return &cm.AppInstallationStatusCode{
			StatusCode: http.StatusOK,
			Response:   newAppInstallation,
		}, nil
	}

	UpdateAppInstallationFromFields(appInstallation, fields)

	response, err := appInstallationResponse(appInstallation, definition)
	if err != nil {
		return nil, err
	}

	return &cm.AppInstallationStatusCode{
		StatusCode: http.StatusOK,
		Response:   *response,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteAppInstallation(_ context.Context, params cm.DeleteAppInstallationParams) (cm.DeleteAppInstallationRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appInstallation := ts.appInstallations.Get(params.SpaceID, params.EnvironmentID, params.AppDefinitionID)
	if appInstallation == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, []byte(`"AppInstallation does not exist."`)), nil
	}

	ts.appInstallations.Delete(params.SpaceID, params.EnvironmentID, params.AppDefinitionID)

	return &cm.NoContent{}, nil
}
