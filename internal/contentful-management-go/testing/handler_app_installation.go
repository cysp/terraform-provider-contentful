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

	environmentID := ts.requestEnvironmentID(ctx, params.SpaceID, params.EnvironmentID)

	if ts.environments.Get(params.SpaceID, environmentID) == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil)), nil
	}

	appInstallation := ts.appInstallations.Get(params.SpaceID, environmentID, params.AppDefinitionID)
	if appInstallation == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(nil, []byte(`"AppInstallation does not exist."`))), nil
	}

	response := projectAppInstallation(*appInstallation, params.EnvironmentID)
	if _, ok := ctx.Value(appAccessTokenContextKey{}).(cm.CreateAppAccessTokenParams); ok {
		return snapshotResponse(&response), nil
	}

	return appInstallationResponse(&response, ts.appDefinitions[params.AppDefinitionID])
}

//nolint:ireturn
func (ts *Handler) PutAppInstallation(ctx context.Context, req *cm.AppInstallationData, params cm.PutAppInstallationParams) (cm.PutAppInstallationRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.requestEnvironmentID(ctx, params.SpaceID, params.EnvironmentID)

	if ts.environments.Get(params.SpaceID, environmentID) == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil)), nil
	}

	definition := ts.appDefinitions[params.AppDefinitionID]
	if definition == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("AppDefinition not found"), nil)), nil
	}

	appInstallation := ts.appInstallations.Get(params.SpaceID, environmentID, params.AppDefinitionID)

	parameters, validation, err := appInstallationParameters(*req, *definition, appInstallation)
	if err != nil {
		return nil, err
	}

	if validation != nil {
		return snapshotResponse(validation), nil
	}

	fields := cm.AppInstallationData{Parameters: parameters}
	if appInstallation == nil {
		newAppInstallation := NewAppInstallationFromFields(params.SpaceID, environmentID, params.AppDefinitionID, fields)
		ts.appInstallations.Set(params.SpaceID, environmentID, params.AppDefinitionID, &newAppInstallation)

		return snapshotResponse(&cm.AppInstallationStatusCode{
			StatusCode: http.StatusOK,
			Response:   projectAppInstallation(newAppInstallation, params.EnvironmentID),
		}), nil
	}

	UpdateAppInstallationFromFields(appInstallation, fields)

	response, err := appInstallationResponse(appInstallation, definition)
	if err != nil {
		return nil, err
	}

	return snapshotResponse(&cm.AppInstallationStatusCode{
		StatusCode: http.StatusOK,
		Response:   projectAppInstallation(*response, params.EnvironmentID),
	}), nil
}

func projectAppInstallation(installation cm.AppInstallation, environmentID string) cm.AppInstallation {
	installation.Sys.Environment.Sys.ID = environmentID

	return installation
}

//nolint:ireturn
func (ts *Handler) DeleteAppInstallation(ctx context.Context, params cm.DeleteAppInstallationParams) (cm.DeleteAppInstallationRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.requestEnvironmentID(ctx, params.SpaceID, params.EnvironmentID)

	appInstallation := ts.appInstallations.Get(params.SpaceID, environmentID, params.AppDefinitionID)
	if appInstallation == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(nil, []byte(`"AppInstallation does not exist."`))), nil
	}

	ts.appInstallations.Delete(params.SpaceID, environmentID, params.AppDefinitionID)

	return snapshotResponse(&cm.NoContent{}), nil
}
