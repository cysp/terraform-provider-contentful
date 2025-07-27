//nolint:dupl
package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetAppInstallation(_ context.Context, params cm.GetAppInstallationParams) (cm.GetAppInstallationRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID || params.AppDefinitionID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	appInstallation := ts.appInstallations.Get(params.SpaceID, params.EnvironmentID, params.AppDefinitionID)
	if appInstallation == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("AppInstallation not found"), nil), nil
	}

	return appInstallation, nil
}

//nolint:ireturn
func (ts *Handler) PutAppInstallation(_ context.Context, req *cm.AppInstallationFields, params cm.PutAppInstallationParams) (cm.PutAppInstallationRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID || params.AppDefinitionID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	appInstallation := ts.appInstallations.Get(params.SpaceID, params.EnvironmentID, params.AppDefinitionID)
	if appInstallation == nil {
		newAppInstallation := NewAppInstallationFromFields(params.SpaceID, params.EnvironmentID, params.AppDefinitionID, *req)
		ts.appInstallations.Set(params.SpaceID, params.EnvironmentID, params.AppDefinitionID, &newAppInstallation)

		return &cm.AppInstallationStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newAppInstallation,
		}, nil
	}

	UpdateAppInstallationFromFields(appInstallation, *req)

	return &cm.AppInstallationStatusCode{
		StatusCode: http.StatusOK,
		Response:   *appInstallation,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteAppInstallation(_ context.Context, params cm.DeleteAppInstallationParams) (cm.DeleteAppInstallationRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID || params.AppDefinitionID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	appInstallation := ts.appInstallations.Get(params.SpaceID, params.EnvironmentID, params.AppDefinitionID)
	if appInstallation == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("AppInstallation not found"), nil), nil
	}

	ts.appInstallations.Delete(params.SpaceID, params.EnvironmentID, params.AppDefinitionID)

	return &cm.NoContent{}, nil
}
