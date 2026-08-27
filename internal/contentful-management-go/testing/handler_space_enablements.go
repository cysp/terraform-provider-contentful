package cmtesting

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetSpaceEnablements(_ context.Context, params cm.GetSpaceEnablementsParams) (cm.GetSpaceEnablementsRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil), nil
	}

	enablements, ok := ts.enablements[params.SpaceID]
	if !ok {
		initial := NewSpaceEnablement(params.SpaceID)
		enablements = &initial
		ts.enablements[params.SpaceID] = enablements
	}

	return enablements, nil
}

//nolint:ireturn
func (ts *Handler) PutSpaceEnablements(_ context.Context, req *cm.SpaceEnablementData, params cm.PutSpaceEnablementsParams) (cm.PutSpaceEnablementsRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil), nil
	}

	enablements, ok := ts.enablements[params.SpaceID]
	if !ok {
		initial := NewSpaceEnablement(params.SpaceID)
		initial.Sys.Version = 1
		enablements = &initial
	}

	if params.XContentfulVersion != enablements.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	crossSpaceLinks, crossSpaceLinksPresent := req.CrossSpaceLinks.Get()

	spaceTemplates, spaceTemplatesPresent := req.SpaceTemplates.Get()
	if !crossSpaceLinksPresent || !spaceTemplatesPresent || crossSpaceLinks.Enabled != spaceTemplates.Enabled {
		return NewContentfulManagementErrorStatusCodeValidationFailed(
			new("Space templates and cross-space links must both be present and enabled or disabled together."),
			nil,
		), nil
	}

	UpdateSpaceEnablementFromRequestFields(enablements, *req)

	if !ok {
		ts.enablements[params.SpaceID] = enablements
	}

	return &cm.SpaceEnablementStatusCode{
		StatusCode: http.StatusOK,
		Response:   *enablements,
	}, nil
}
