package cmtesting

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetLivePreviewVariables(_ context.Context, params cm.GetLivePreviewVariablesParams) (cm.GetLivePreviewVariablesRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return livePreviewVariablesEnvironmentNotFound(params.EnvironmentID)
	}

	variables := ts.livePreviewVariables.Get(params.SpaceID, params.EnvironmentID)
	if variables == nil {
		return livePreviewVariablesError(http.StatusNotFound, cm.ErrorSysIDNotFound, "The resource could not be found."), nil
	}

	return variables, nil
}

//nolint:ireturn
func (ts *Handler) PutLivePreviewVariables(_ context.Context, req *cm.LivePreviewVariablesData, params cm.PutLivePreviewVariablesParams) (cm.PutLivePreviewVariablesRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return livePreviewVariablesEnvironmentNotFound(params.EnvironmentID)
	}

	if params.XContentfulVersion < 0 {
		return livePreviewVariablesError(http.StatusBadRequest, "BadRequest", "The 'x-contentful-version' header is missing or invalid."), nil
	}

	version := 1

	if variables := ts.livePreviewVariables.Get(params.SpaceID, params.EnvironmentID); variables != nil {
		if params.XContentfulVersion != variables.Sys.Version {
			return livePreviewVariablesError(http.StatusConflict, cm.ErrorSysIDVersionMismatch, "The given version value is not the current one"), nil
		}

		version = variables.Sys.Version + 1
	}

	normalized, failure, err := validateLivePreviewVariablesData(req.Variables)
	if err != nil {
		return nil, err
	}

	if failure != nil {
		return failure, nil
	}

	variables := &cm.LivePreviewVariables{
		Sys: cm.LivePreviewVariablesSys{
			Space:       cm.NewSpaceLink(params.SpaceID),
			Environment: cm.NewEnvironmentLink(params.EnvironmentID),
			Version:     version,
		},
		Variables: normalized,
	}
	ts.livePreviewVariables.Set(params.SpaceID, params.EnvironmentID, variables)

	return variables, nil
}

//nolint:ireturn
func (ts *Handler) DeleteLivePreviewVariables(_ context.Context, params cm.DeleteLivePreviewVariablesParams) (cm.DeleteLivePreviewVariablesRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return livePreviewVariablesEnvironmentNotFound(params.EnvironmentID)
	}

	ts.livePreviewVariables.Delete(params.SpaceID, params.EnvironmentID)

	return &cm.NoContent{}, nil
}

func livePreviewVariablesError(statusCode int, id, message string) *cm.LivePreviewVariablesErrorStatusCode {
	return &cm.LivePreviewVariablesErrorStatusCode{
		StatusCode: statusCode,
		Response:   cm.NewErrorLivePreviewVariablesError(NewContentfulManagementError(id, &message, nil)),
	}
}

func livePreviewVariablesEnvironmentNotFound(environmentID string) (*cm.LivePreviewVariablesErrorStatusCode, error) {
	details, err := json.Marshal(map[string]string{"type": "Environment", "id": environmentID})
	if err != nil {
		return nil, fmt.Errorf("encode missing environment details: %w", err)
	}

	return &cm.LivePreviewVariablesErrorStatusCode{StatusCode: http.StatusNotFound, Response: cm.NewErrorLivePreviewVariablesError(NewContentfulManagementError(cm.ErrorSysIDNotFound, new("The resource could not be found."), details))}, nil
}
