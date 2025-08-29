package testing

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn,revive
func (ts *Handler) GetPreviewApiKey(_ context.Context, params cm.GetPreviewApiKeyParams) (cm.GetPreviewApiKeyRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	previewAPIKey := ts.previewAPIKeys.Get(params.SpaceID, params.PreviewAPIKeyID)
	if previewAPIKey == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("PreviewApiKey not found"), nil), nil
	}

	return previewAPIKey, nil
}
