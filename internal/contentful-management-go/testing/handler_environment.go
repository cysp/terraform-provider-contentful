package cmtesting

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
		return NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil), nil
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

	if alias := ts.environmentAliases.Get(params.SpaceID, params.EnvironmentID); alias != nil {
		target := ts.environments.Get(params.SpaceID, alias.Environment.Sys.ID)
		if target == nil {
			return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
		}

		environment := *target
		environment.Sys.ID = params.EnvironmentID
		environment.Name = params.EnvironmentID
		environment.Sys.AliasedEnvironment = cm.NewOptEnvironmentLink(alias.Environment)

		return &environment, nil
	}

	environment := ts.environments.Get(params.SpaceID, params.EnvironmentID)
	if environment == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	return environment, nil
}

//nolint:ireturn
func (ts *Handler) DeleteEnvironment(_ context.Context, params cm.DeleteEnvironmentParams) (cm.DeleteEnvironmentRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environment := ts.environments.Get(params.SpaceID, params.EnvironmentID)
	if environment == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	for _, locale := range ts.locales.List(params.SpaceID, params.EnvironmentID) {
		ts.locales.Delete(params.SpaceID, params.EnvironmentID, locale.Sys.ID)
	}

	ts.environments.Delete(params.SpaceID, params.EnvironmentID)
	ts.livePreviewVariables.Delete(params.SpaceID, params.EnvironmentID)

	return &cm.NoContent{}, nil
}
