package cmtesting

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreateWebhookDefinition(_ context.Context, req *cm.WebhookDefinitionCreateData, params cm.CreateWebhookDefinitionParams) (cm.CreateWebhookDefinitionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if len(req.Topics) == 0 {
		return snapshotResponse(newEmptyWebhookTopicsValidationError()), nil
	}

	if !ts.hasMasterEnvironment(params.SpaceID) {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil)), nil
	}

	fields := cm.WebhookDefinitionData{
		Name: req.Name, URL: req.URL, Topics: req.Topics, Filters: req.Filters,
		HttpBasicUsername: req.HttpBasicUsername, HttpBasicPassword: req.HttpBasicPassword,
		Transformation: req.Transformation, Active: req.Active,
		Headers: convertSlice(req.Headers, func(header cm.WebhookDefinitionCreateHeader) cm.WebhookDefinitionRequestHeader {
			return cm.WebhookDefinitionRequestHeader{Key: header.Key, Secret: header.Secret, Value: cm.NewOptNilString(header.Value)}
		}),
	}

	validationResponse, err := validateWebhookDefinition(nil, &fields)
	if err != nil || validationResponse != nil {
		return validationResponse, err
	}

	webhookDefinition := NewWebhookDefinitionFromFields(params.SpaceID, generateResourceID(), fields)
	ts.webhookDefinitions.Set(params.SpaceID, webhookDefinition.Sys.ID, &webhookDefinition)
	response := redactWebhookDefinitionSecrets(webhookDefinition)

	return snapshotResponse(&cm.WebhookDefinitionStatusCode{
		StatusCode: http.StatusOK,
		Response:   response,
	}), nil
}

//nolint:ireturn
func (ts *Handler) GetWebhookDefinition(_ context.Context, params cm.GetWebhookDefinitionParams) (cm.GetWebhookDefinitionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	webhookDefinition := ts.webhookDefinitions.Get(params.SpaceID, params.WebhookDefinitionID)
	if webhookDefinition == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("WebhookDefinition not found"), nil)), nil
	}

	response := redactWebhookDefinitionSecrets(*webhookDefinition)

	return snapshotResponse(&response), nil
}

//nolint:ireturn
func (ts *Handler) UpdateWebhookDefinition(_ context.Context, req *cm.WebhookDefinitionData, params cm.UpdateWebhookDefinitionParams) (cm.UpdateWebhookDefinitionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if len(req.Topics) == 0 {
		return snapshotResponse(newEmptyWebhookTopicsValidationError()), nil
	}

	if !ts.hasMasterEnvironment(params.SpaceID) {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil)), nil
	}

	webhookDefinition := ts.webhookDefinitions.Get(params.SpaceID, params.WebhookDefinitionID)

	validationResponse, err := validateWebhookDefinition(webhookDefinition, req)
	if err != nil || validationResponse != nil {
		return validationResponse, err
	}

	if webhookDefinition == nil {
		newWebhookDefinition := NewWebhookDefinitionFromFields(params.SpaceID, params.WebhookDefinitionID, *req)
		ts.webhookDefinitions.Set(params.SpaceID, newWebhookDefinition.Sys.ID, &newWebhookDefinition)
		response := redactWebhookDefinitionSecrets(newWebhookDefinition)

		return snapshotResponse(&cm.WebhookDefinitionStatusCode{
			StatusCode: http.StatusOK,
			Response:   response,
		}), nil
	}

	if version, present := params.XContentfulVersion.Get(); !present || version != webhookDefinition.Sys.Version {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeConflict(nil, nil)), nil
	}

	if validationError := applyWebhookDefinitionUpdate(webhookDefinition, *req); validationError != nil {
		return snapshotResponse(validationError), nil
	}

	response := redactWebhookDefinitionSecrets(*webhookDefinition)

	return snapshotResponse(&cm.WebhookDefinitionStatusCode{
		StatusCode: http.StatusOK,
		Response:   response,
	}), nil
}

func newEmptyWebhookTopicsValidationError() *cm.ErrorStatusCode {
	return NewContentfulManagementErrorStatusCodeValidationFailed(
		new("Validation error"),
		[]byte(`{"errors":[{"name":"topics","path":[],"details":"Topics cannot be empty"}]}`),
	)
}

//nolint:ireturn
func (ts *Handler) DeleteWebhookDefinition(_ context.Context, params cm.DeleteWebhookDefinitionParams) (cm.DeleteWebhookDefinitionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	webhookDefinition := ts.webhookDefinitions.Get(params.SpaceID, params.WebhookDefinitionID)
	if webhookDefinition == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("WebhookDefinition not found"), nil)), nil
	}

	ts.webhookDefinitions.Delete(params.SpaceID, params.WebhookDefinitionID)

	return snapshotResponse(&cm.NoContent{}), nil
}
