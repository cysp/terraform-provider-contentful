package cmtesting

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreateWebhookDefinition(_ context.Context, req *cm.WebhookDefinitionData, params cm.CreateWebhookDefinitionParams) (cm.CreateWebhookDefinitionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if len(req.Topics) == 0 {
		return snapshotResponse(newEmptyWebhookTopicsValidationError()), nil
	}

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil)), nil
	}

	validationResponse, err := validateWebhookCreationHeaders(http.MethodPost, req.Headers)
	if err != nil || validationResponse != nil {
		return validationResponse, err
	}

	webhookDefinition := NewWebhookDefinitionFromFields(params.SpaceID, generateResourceID(), *req)
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

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil)), nil
	}

	webhookDefinition := ts.webhookDefinitions.Get(params.SpaceID, params.WebhookDefinitionID)
	if webhookDefinition == nil {
		validationResponse, err := validateWebhookCreationHeaders(http.MethodPut, req.Headers)
		if err != nil || validationResponse != nil {
			return validationResponse, err
		}

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

	if validationError := UpdateWebhookDefinitionFromFields(webhookDefinition, *req); validationError != nil {
		return snapshotResponse(validationError), nil
	}

	response := redactWebhookDefinitionSecrets(*webhookDefinition)

	return snapshotResponse(&cm.WebhookDefinitionStatusCode{
		StatusCode: http.StatusOK,
		Response:   response,
	}), nil
}

func validateWebhookCreationHeaders(method string, headers cm.WebhookDefinitionHeaders) (*cm.ErrorStatusCode, error) {
	var (
		failures     []map[string]any
		missingValue bool
	)

	for index, header := range headers {
		if header.Value.IsSet() {
			continue
		}

		missingValue = true

		if method == http.MethodPost {
			failures = append(failures, map[string]any{
				"name": "invalid_type", "path": []any{"headers", index, "value"},
				"details": fmt.Sprintf(`Invalid input: expected string, received undefined at "headers.%d.value"`, index),
			})
		} else if !header.Secret.IsSet() {
			failures = append(failures, map[string]any{
				"name": "invalid_union", "path": []any{"headers", index},
				"details": fmt.Sprintf(`Invalid input at "headers.%d"`, index),
			})
		}
	}

	if len(failures) > 0 {
		details, err := json.Marshal(map[string]any{"errors": failures})
		if err != nil {
			return nil, fmt.Errorf("encode webhook header validation details: %w", err)
		}

		return NewContentfulManagementErrorStatusCodeValidationFailed(new("Validation error"), details), nil
	}

	if missingValue {
		return NewContentfulManagementErrorStatusCode(http.StatusInternalServerError, "ServerError", new("Failed to create webhook definition/s"), nil), nil
	}

	return nil, nil //nolint:nilnil // All header values are present.
}

func missingWebhookTopicsValidationDetails() []byte {
	return []byte(`{"errors":[{"name":"invalid_type","path":["topics"],"details":"Invalid input: expected array, received undefined at \"topics\""}]}`)
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
