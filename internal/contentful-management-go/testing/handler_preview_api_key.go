package testing

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetPreviewAPIKey(_ context.Context, params cm.GetPreviewAPIKeyParams) (cm.GetPreviewAPIKeyRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.PreviewAPIKeyID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	previewAPIKey := ts.previewAPIKeys.Get(params.SpaceID, params.PreviewAPIKeyID)
	if previewAPIKey == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("PreviewApiKey not found"), nil), nil
	}

	return previewAPIKey, nil
}
