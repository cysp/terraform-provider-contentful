package cmtesting

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetEditorInterface(ctx context.Context, params cm.GetEditorInterfaceParams) (cm.GetEditorInterfaceRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.requestEnvironmentID(ctx, params.SpaceID, params.EnvironmentID)

	editorInterface := ts.editorInterfaces.Get(params.SpaceID, environmentID, params.ContentTypeID)
	if editorInterface == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("EditorInterface not found"), nil)), nil
	}

	response := projectEditorInterfaceResponse(*editorInterface, params.EnvironmentID)

	return snapshotResponse(&response), nil
}

//nolint:ireturn
func (ts *Handler) PutEditorInterface(ctx context.Context, req *cm.EditorInterfaceData, params cm.PutEditorInterfaceParams) (cm.PutEditorInterfaceRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.requestEnvironmentID(ctx, params.SpaceID, params.EnvironmentID)

	contentType := ts.contentTypes.Get(params.SpaceID, environmentID, params.ContentTypeID)
	if contentType == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("ContentType not found"), nil)), nil
	}

	editorInterface := ts.editorInterfaces.Get(params.SpaceID, environmentID, params.ContentTypeID)
	if editorInterface == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeBadRequest(new("The content type you sent could not be found or was not activated."), nil)), nil
	}

	if params.XContentfulVersion != editorInterface.Sys.Version {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil)), nil
	}

	UpdateEditorInterfaceFromFields(editorInterface, *req)

	return snapshotResponse(&cm.EditorInterfaceStatusCode{
		StatusCode: http.StatusOK,
		Response:   projectEditorInterfaceResponse(*editorInterface, params.EnvironmentID),
	}), nil
}

func projectEditorInterfaceResponse(editorInterface cm.EditorInterface, environmentID string) cm.EditorInterface {
	editorInterface.Sys.Environment.Sys.ID = environmentID

	return editorInterface
}
