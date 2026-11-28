package cmtesting

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetLivePreviewVariables(ctx context.Context, params cm.GetLivePreviewVariablesParams) (cm.GetLivePreviewVariablesRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.requestEnvironmentID(ctx, params.SpaceID, params.EnvironmentID)

	if ts.environments.Get(params.SpaceID, environmentID) == nil {
		return livePreviewVariablesEnvironmentNotFound(params.EnvironmentID)
	}

	variables := ts.livePreviewVariables.Get(params.SpaceID, environmentID)
	if variables == nil {
		return snapshotResponse(livePreviewVariablesError(http.StatusNotFound, cm.ErrorSysIDNotFound, "The resource could not be found.")), nil
	}

	return livePreviewVariablesResponse(*variables, params.EnvironmentID), nil
}

//nolint:ireturn
func (ts *Handler) PutLivePreviewVariables(ctx context.Context, req *cm.LivePreviewVariablesData, params cm.PutLivePreviewVariablesParams) (cm.PutLivePreviewVariablesRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.requestEnvironmentID(ctx, params.SpaceID, params.EnvironmentID)

	if ts.environments.Get(params.SpaceID, environmentID) == nil {
		return livePreviewVariablesEnvironmentNotFound(params.EnvironmentID)
	}

	if params.XContentfulVersion < 0 {
		return snapshotResponse(livePreviewVariablesError(http.StatusBadRequest, "BadRequest", "The 'x-contentful-version' header is missing or invalid.")), nil
	}

	version := 1

	if variables := ts.livePreviewVariables.Get(params.SpaceID, environmentID); variables != nil {
		if params.XContentfulVersion != variables.Sys.Version {
			return snapshotResponse(livePreviewVariablesError(http.StatusConflict, cm.ErrorSysIDVersionMismatch, "The given version value is not the current one")), nil
		}

		version = variables.Sys.Version + 1
	}

	normalized, failure, err := validateLivePreviewVariablesData(req.Variables)
	if err != nil {
		return nil, err
	}

	if failure != nil {
		return snapshotResponse(failure), nil
	}

	variables := &cm.LivePreviewVariables{
		Sys: cm.LivePreviewVariablesSys{
			Space:       cm.NewSpaceLink(params.SpaceID),
			Environment: cm.NewEnvironmentLink(environmentID),
			Version:     version,
		},
		Variables: normalized,
	}
	ts.livePreviewVariables.Set(params.SpaceID, environmentID, variables)

	return livePreviewVariablesResponse(*variables, params.EnvironmentID), nil
}

//nolint:ireturn
func (ts *Handler) DeleteLivePreviewVariables(ctx context.Context, params cm.DeleteLivePreviewVariablesParams) (cm.DeleteLivePreviewVariablesRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.requestEnvironmentID(ctx, params.SpaceID, params.EnvironmentID)

	if ts.environments.Get(params.SpaceID, environmentID) == nil {
		return livePreviewVariablesEnvironmentNotFound(params.EnvironmentID)
	}

	ts.livePreviewVariables.Delete(params.SpaceID, environmentID)

	return snapshotResponse(&cm.NoContent{}), nil
}

func livePreviewVariablesResponse(variables cm.LivePreviewVariables, environmentID string) *cm.LivePreviewVariables {
	variables.Sys.Environment.Sys.ID = environmentID

	return snapshotResponse(&variables)
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
