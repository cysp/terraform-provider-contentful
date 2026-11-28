package cmtesting

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetAppInstallation(ctx context.Context, params cm.GetAppInstallationParams) (cm.GetAppInstallationRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil)), nil
	}

	appInstallation := ts.appInstallations.Get(params.SpaceID, params.EnvironmentID, params.AppDefinitionID)
	if appInstallation == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(nil, []byte(`"AppInstallation does not exist."`))), nil
	}

	response := projectAppInstallation(ctx, *appInstallation)
	if _, ok := ctx.Value(appAccessTokenContextKey{}).(cm.CreateAppAccessTokenParams); ok {
		return snapshotResponse(&response), nil
	}

	return appInstallationResponse(&response, ts.appDefinitions[params.AppDefinitionID])
}

//nolint:ireturn
func (ts *Handler) PutAppInstallation(ctx context.Context, req *cm.AppInstallationData, params cm.PutAppInstallationParams) (cm.PutAppInstallationRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil)), nil
	}

	definition := ts.appDefinitions[params.AppDefinitionID]
	if definition == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("AppDefinition not found"), nil)), nil
	}

	appInstallation := ts.appInstallations.Get(params.SpaceID, params.EnvironmentID, params.AppDefinitionID)

	parameters, validation, err := appInstallationParameters(*req, *definition, appInstallation)
	if err != nil {
		return nil, err
	}

	if validation != nil {
		return snapshotResponse(validation), nil
	}

	fields := cm.AppInstallationData{Parameters: parameters}
	if appInstallation == nil {
		newAppInstallation := NewAppInstallationFromFields(params.SpaceID, params.EnvironmentID, params.AppDefinitionID, fields)
		ts.appInstallations.Set(params.SpaceID, params.EnvironmentID, params.AppDefinitionID, &newAppInstallation)

		return snapshotResponse(&cm.AppInstallationStatusCode{
			StatusCode: http.StatusOK,
			Response:   projectAppInstallation(ctx, newAppInstallation),
		}), nil
	}

	UpdateAppInstallationFromFields(appInstallation, fields)

	response, err := appInstallationResponse(appInstallation, definition)
	if err != nil {
		return nil, err
	}

	return snapshotResponse(&cm.AppInstallationStatusCode{
		StatusCode: http.StatusOK,
		Response:   projectAppInstallation(ctx, *response),
	}), nil
}

func projectAppInstallation(ctx context.Context, installation cm.AppInstallation) cm.AppInstallation {
	if addressed, ok := ctx.Value(addressedEnvironmentKey{}).(string); ok {
		installation.Sys.Environment.Sys.ID = addressed
	}

	return installation
}

//nolint:ireturn
func (ts *Handler) DeleteAppInstallation(_ context.Context, params cm.DeleteAppInstallationParams) (cm.DeleteAppInstallationRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appInstallation := ts.appInstallations.Get(params.SpaceID, params.EnvironmentID, params.AppDefinitionID)
	if appInstallation == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(nil, []byte(`"AppInstallation does not exist."`))), nil
	}

	ts.appInstallations.Delete(params.SpaceID, params.EnvironmentID, params.AppDefinitionID)

	return snapshotResponse(&cm.NoContent{}), nil
}
