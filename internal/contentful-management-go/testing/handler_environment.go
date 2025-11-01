package testing

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetEnvironment(_ context.Context, params cm.GetEnvironmentParams) (cm.GetEnvironmentRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	environment := ts.environments.Get(params.SpaceID, params.EnvironmentID)
	if environment == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Environment not found"), nil), nil
	}

	return environment, nil
}

//nolint:ireturn
func (ts *Handler) CreateOrUpdateEnvironment(_ context.Context, req *cm.EnvironmentFields, params cm.CreateOrUpdateEnvironmentParams) (cm.CreateOrUpdateEnvironmentRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	environment := ts.environments.Get(params.SpaceID, params.EnvironmentID)
	if environment == nil {
		newEnvironment := NewEnvironmentFromRequest(params.SpaceID, params.EnvironmentID, *req)
		ts.environments.Set(params.SpaceID, params.EnvironmentID, &newEnvironment)

		result := cm.CreateOrUpdateEnvironmentCreated(newEnvironment)
		return &result, nil
	}

	UpdateEnvironmentFromRequest(environment, *req)

	result := cm.CreateOrUpdateEnvironmentOK(*environment)
	return &result, nil
}

//nolint:ireturn
func (ts *Handler) DeleteEnvironment(_ context.Context, params cm.DeleteEnvironmentParams) (cm.DeleteEnvironmentRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	environment := ts.environments.Get(params.SpaceID, params.EnvironmentID)
	if environment == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Environment not found"), nil), nil
	}

	ts.environments.Delete(params.SpaceID, params.EnvironmentID)

	return &cm.NoContent{}, nil
}
