package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreateOrUpdateEnvironment(_ context.Context, req *cm.EnvironmentData, params cm.CreateOrUpdateEnvironmentParams) (cm.CreateOrUpdateEnvironmentRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Space not found"), nil), nil
	}

	environment := ts.environments.Get(params.SpaceID, params.EnvironmentID)
	if environment == nil {
		environment := NewEnvironmentFromEnvironmentData(params.SpaceID, params.EnvironmentID, "ready", *req)
		ts.environments.Set(params.SpaceID, params.EnvironmentID, &environment)

		return &cm.EnvironmentStatusCode{
			StatusCode: http.StatusCreated,
			Response:   environment,
		}, nil
	}

	if params.XContentfulVersion.Or(1) != environment.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	UpdateEnvironmentFromEnvironmentData(environment, *req)

	return &cm.EnvironmentStatusCode{
		StatusCode: http.StatusOK,
		Response:   *environment,
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetEnvironment(_ context.Context, params cm.GetEnvironmentParams) (cm.GetEnvironmentRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environment := ts.environments.Get(params.SpaceID, params.EnvironmentID)
	if environment == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Environment not found"), nil), nil
	}

	return environment, nil
}

//nolint:ireturn
func (ts *Handler) DeleteEnvironment(_ context.Context, params cm.DeleteEnvironmentParams) (cm.DeleteEnvironmentRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environment := ts.environments.Get(params.SpaceID, params.EnvironmentID)
	if environment == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Environment not found"), nil), nil
	}

	ts.environments.Delete(params.SpaceID, params.EnvironmentID)

	return &cm.NoContent{}, nil
}
