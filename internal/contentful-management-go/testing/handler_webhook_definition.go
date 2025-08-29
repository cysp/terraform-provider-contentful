//nolint:dupl
package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreateWebhookDefinition(_ context.Context, req *cm.WebhookDefinitionFields, params cm.CreateWebhookDefinitionParams) (cm.CreateWebhookDefinitionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	webhookDefinition := NewWebhookDefinitionFromFields(params.SpaceID, generateResourceID(), *req)
	ts.webhookDefinitions.Set(params.SpaceID, webhookDefinition.Sys.ID, &webhookDefinition)

	return &cm.WebhookDefinitionStatusCode{
		StatusCode: http.StatusCreated,
		Response:   webhookDefinition,
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetWebhookDefinition(_ context.Context, params cm.GetWebhookDefinitionParams) (cm.GetWebhookDefinitionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	webhookDefinition := ts.webhookDefinitions.Get(params.SpaceID, params.WebhookDefinitionID)
	if webhookDefinition == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("WebhookDefinition not found"), nil), nil
	}

	return webhookDefinition, nil
}

//nolint:ireturn
func (ts *Handler) UpdateWebhookDefinition(_ context.Context, req *cm.WebhookDefinitionFields, params cm.UpdateWebhookDefinitionParams) (cm.UpdateWebhookDefinitionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	webhookDefinition := ts.webhookDefinitions.Get(params.SpaceID, params.WebhookDefinitionID)
	if webhookDefinition == nil {
		newWebhookDefinition := NewWebhookDefinitionFromFields(params.SpaceID, params.WebhookDefinitionID, *req)
		ts.webhookDefinitions.Set(params.SpaceID, newWebhookDefinition.Sys.ID, &newWebhookDefinition)

		return &cm.WebhookDefinitionStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newWebhookDefinition,
		}, nil
	}

	UpdateWebhookDefinitionFromFields(webhookDefinition, *req)

	return &cm.WebhookDefinitionStatusCode{
		StatusCode: http.StatusOK,
		Response:   *webhookDefinition,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteWebhookDefinition(_ context.Context, params cm.DeleteWebhookDefinitionParams) (cm.DeleteWebhookDefinitionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	webhookDefinition := ts.webhookDefinitions.Get(params.SpaceID, params.WebhookDefinitionID)
	if webhookDefinition == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("WebhookDefinition not found"), nil), nil
	}

	ts.webhookDefinitions.Delete(params.SpaceID, params.WebhookDefinitionID)

	return &cm.NoContent{}, nil
}
