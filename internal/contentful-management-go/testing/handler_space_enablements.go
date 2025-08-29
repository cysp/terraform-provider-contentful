package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetSpaceEnablements(_ context.Context, params cm.GetSpaceEnablementsParams) (cm.GetSpaceEnablementsRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	enablements, ok := ts.enablements[params.SpaceID]
	if !ok {
		enablements = pointerTo(NewSpaceEnablement(params.SpaceID))
		ts.enablements[params.SpaceID] = enablements
	}

	return pointerTo(cm.GetSpaceEnablementsApplicationVndContentfulManagementV1JSONOK(*enablements)), nil
}

//nolint:ireturn
func (ts *Handler) PutSpaceEnablements(_ context.Context, req *cm.SpaceEnablementFields, params cm.PutSpaceEnablementsParams) (cm.PutSpaceEnablementsRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	enablements, ok := ts.enablements[params.SpaceID]
	if !ok {
		enablements = pointerTo(NewSpaceEnablement(params.SpaceID))
		ts.enablements[params.SpaceID] = enablements
	}

	UpdateSpaceEnablementFromRequestFields(enablements, *req)

	return &cm.SpaceEnablementStatusCode{
		StatusCode: http.StatusOK,
		Response:   *enablements,
	}, nil
}
