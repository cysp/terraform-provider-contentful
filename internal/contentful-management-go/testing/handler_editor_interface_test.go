package cmtesting_test

import (
	"context"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutEditorInterfaceRequiresActivatedContentType(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	contentTypeRequest := newContentTypeRequest()
	putContentType(t, handler, &contentTypeRequest, 1, http.StatusCreated)

	response, err := handler.PutEditorInterface(context.Background(), &cm.EditorInterfaceData{}, cm.PutEditorInterfaceParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "content-type", XContentfulVersion: 1,
	})
	require.NoError(t, err)
	requireContentfulError(
		t,
		response,
		http.StatusBadRequest,
		"BadRequest",
		"The content type you sent could not be found or was not activated.",
	)
}

func TestPutEditorInterfaceUsesContentfulVersioning(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	contentTypeRequest := newContentTypeRequest()
	created := putContentType(t, handler, &contentTypeRequest, 1, http.StatusCreated)
	activateContentType(t, handler, created.Sys.Version)

	editorInterfaceRequest := cm.EditorInterfaceData{
		Controls: cm.NewOptNilEditorInterfaceDataControlsItemArray([]cm.EditorInterfaceDataControlsItem{{
			FieldId: "title", WidgetNamespace: cm.NewOptString("builtin"), WidgetId: cm.NewOptString("singleLine"),
		}}),
	}
	response, err := handler.PutEditorInterface(context.Background(), &editorInterfaceRequest, cm.PutEditorInterfaceParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "content-type", XContentfulVersion: 1,
	})
	require.NoError(t, err)

	statusCode, statusCodeOK := response.(*cm.EditorInterfaceStatusCode)
	require.True(t, statusCodeOK)
	assert.Equal(t, http.StatusOK, statusCode.StatusCode)
	assert.Equal(t, 2, statusCode.Response.Sys.Version)

	staleResponse, err := handler.PutEditorInterface(context.Background(), &editorInterfaceRequest, cm.PutEditorInterfaceParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "content-type", XContentfulVersion: 1,
	})
	require.NoError(t, err)
	requireContentfulError(t, staleResponse, http.StatusConflict, cm.ErrorSysIDVersionMismatch, "")
}

func TestActivateContentTypeSynchronizesEditorInterface(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	request := newContentTypeRequest()
	created := putContentType(t, handler, &request, 1, http.StatusCreated)

	editorResponse, err := handler.GetEditorInterface(context.Background(), contentTypeEditorInterfaceParams())
	require.NoError(t, err)
	requireContentfulError(t, editorResponse, http.StatusNotFound, cm.ErrorSysIDNotFound, "EditorInterface not found")

	activated := activateContentType(t, handler, created.Sys.Version)
	editorInterface := getEditorInterface(t, handler)
	assert.Equal(t, 1, editorInterface.Sys.Version)
	assert.Equal(t, []string{"title"}, editorInterfaceControlFieldIDs(t, editorInterface))

	request.Fields = append(request.Fields, cm.ContentTypeRequestDataFieldsItem{ID: "subtitle", Name: "Subtitle", Type: "Symbol"})
	draftUpdated := putContentType(t, handler, &request, activated.Sys.Version, http.StatusOK)
	editorInterface = getEditorInterface(t, handler)
	assert.Equal(t, 1, editorInterface.Sys.Version, "draft updates must not change the editor interface")
	assert.Equal(t, []string{"title"}, editorInterfaceControlFieldIDs(t, editorInterface))

	activateContentType(t, handler, draftUpdated.Sys.Version)
	editorInterface = getEditorInterface(t, handler)
	assert.Equal(t, 2, editorInterface.Sys.Version)
	assert.Equal(t, []string{"title", "subtitle"}, editorInterfaceControlFieldIDs(t, editorInterface))
}

func TestActivateContentTypePreservesClearedEditorInterfaceControls(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	request := newContentTypeRequest()
	created := putContentType(t, handler, &request, 1, http.StatusCreated)
	activated := activateContentType(t, handler, created.Sys.Version)

	response, err := handler.PutEditorInterface(context.Background(), &cm.EditorInterfaceData{}, cm.PutEditorInterfaceParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "content-type", XContentfulVersion: 1,
	})
	require.NoError(t, err)

	statusCode, statusCodeOK := response.(*cm.EditorInterfaceStatusCode)
	require.True(t, statusCodeOK)
	assert.False(t, statusCode.Response.Controls.IsSet())

	request.Fields = append(request.Fields, cm.ContentTypeRequestDataFieldsItem{ID: "subtitle", Name: "Subtitle", Type: "Symbol"})
	draftUpdated := putContentType(t, handler, &request, activated.Sys.Version, http.StatusOK)
	activateContentType(t, handler, draftUpdated.Sys.Version)

	editorInterface := getEditorInterface(t, handler)
	assert.Equal(t, 3, editorInterface.Sys.Version)
	assert.False(t, editorInterface.Controls.IsSet())
}

func TestActivateContentTypeAddsControlsOnlyForNewFields(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	request := newContentTypeRequest()
	created := putContentType(t, handler, &request, 1, http.StatusCreated)
	activated := activateContentType(t, handler, created.Sys.Version)

	editorInterfaceRequest := cm.EditorInterfaceData{
		Controls: cm.NewOptNilEditorInterfaceDataControlsItemArray([]cm.EditorInterfaceDataControlsItem{}),
	}
	response, err := handler.PutEditorInterface(context.Background(), &editorInterfaceRequest, cm.PutEditorInterfaceParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "content-type", XContentfulVersion: 1,
	})
	require.NoError(t, err)

	statusCode, statusCodeOK := response.(*cm.EditorInterfaceStatusCode)
	require.True(t, statusCodeOK)
	assert.Empty(t, editorInterfaceControlFieldIDs(t, &statusCode.Response))

	request.Fields = append(request.Fields, cm.ContentTypeRequestDataFieldsItem{ID: "subtitle", Name: "Subtitle", Type: "Symbol"})
	draftUpdated := putContentType(t, handler, &request, activated.Sys.Version, http.StatusOK)
	activateContentType(t, handler, draftUpdated.Sys.Version)

	editorInterface := getEditorInterface(t, handler)
	assert.Equal(t, 3, editorInterface.Sys.Version)
	assert.Equal(t, []string{"subtitle"}, editorInterfaceControlFieldIDs(t, editorInterface))
}

func contentTypeEditorInterfaceParams() cm.GetEditorInterfaceParams {
	return cm.GetEditorInterfaceParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "content-type",
	}
}

func getEditorInterface(t *testing.T, handler *cmt.Handler) *cm.EditorInterface {
	t.Helper()

	response, err := handler.GetEditorInterface(context.Background(), contentTypeEditorInterfaceParams())
	require.NoError(t, err)

	editorInterface, ok := response.(*cm.EditorInterface)
	require.True(t, ok)

	return editorInterface
}

func editorInterfaceControlFieldIDs(t *testing.T, editorInterface *cm.EditorInterface) []string {
	t.Helper()

	controls, controlsOK := editorInterface.Controls.Get()
	require.True(t, controlsOK)

	fieldIDs := make([]string, len(controls))
	for index, control := range controls {
		fieldIDs[index] = control.FieldId
	}

	return fieldIDs
}
