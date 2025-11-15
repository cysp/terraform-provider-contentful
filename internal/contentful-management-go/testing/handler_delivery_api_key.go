package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn,revive
func (ts *Handler) CreateDeliveryApiKey(_ context.Context, req *cm.ApiKeyRequestData, params cm.CreateDeliveryApiKeyParams) (cm.CreateDeliveryApiKeyRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	apiKeyID := generateResourceID()
	apiKey := NewAPIKeyFromRequestFields(params.SpaceID, apiKeyID, *req)
	apiKey.AccessToken = generateResourceID()

	previewAPIKeyID := generateResourceID()
	previewAPIKey := cm.PreviewApiKey{
		Sys:         cm.NewPreviewAPIKeySys(params.SpaceID, previewAPIKeyID),
		AccessToken: generateResourceID(),
	}

	apiKey.PreviewAPIKey.SetTo(cm.NewPreviewAPIKeyLink(previewAPIKeyID))

	ts.apiKeys.Set(params.SpaceID, apiKeyID, &apiKey)
	ts.previewAPIKeys.Set(params.SpaceID, previewAPIKeyID, &previewAPIKey)

	return &cm.ApiKeyStatusCode{
		StatusCode: http.StatusCreated,
		Response:   apiKey,
	}, nil
}

//nolint:ireturn,revive
func (ts *Handler) GetDeliveryApiKey(_ context.Context, params cm.GetDeliveryApiKeyParams) (cm.GetDeliveryApiKeyRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.APIKeyID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	apiKey := ts.apiKeys.Get(params.SpaceID, params.APIKeyID)
	if apiKey == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("ApiKey not found"), nil), nil
	}

	return apiKey, nil
}

//nolint:ireturn,revive
func (ts *Handler) UpdateDeliveryApiKey(_ context.Context, req *cm.ApiKeyRequestData, params cm.UpdateDeliveryApiKeyParams) (cm.UpdateDeliveryApiKeyRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.APIKeyID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	apiKey := ts.apiKeys.Get(params.SpaceID, params.APIKeyID)
	if apiKey == nil {
		newApiKey := NewAPIKeyFromRequestFields(params.SpaceID, params.APIKeyID, *req)
		ts.apiKeys.Set(params.SpaceID, params.APIKeyID, &newApiKey)

		return &cm.ApiKeyStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newApiKey,
		}, nil
	}

	UpdateAPIKeyFromRequestFields(apiKey, *req)

	return &cm.ApiKeyStatusCode{
		StatusCode: http.StatusOK,
		Response:   *apiKey,
	}, nil
}

//nolint:ireturn,revive
func (ts *Handler) DeleteDeliveryApiKey(_ context.Context, params cm.DeleteDeliveryApiKeyParams) (cm.DeleteDeliveryApiKeyRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.APIKeyID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	apiKey := ts.apiKeys.Get(params.SpaceID, params.APIKeyID)
	if apiKey == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("ApiKey not found"), nil), nil
	}

	ts.apiKeys.Delete(params.SpaceID, params.APIKeyID)

	return &cm.NoContent{}, nil
}
