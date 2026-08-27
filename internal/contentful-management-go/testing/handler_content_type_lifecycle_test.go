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

func TestActivateContentTypeUsesContentfulPublicationVersioning(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	request := newContentTypeRequest()
	created := createContentType(t, handler, &request)
	assert.Equal(t, 1, created.Sys.Version)
	assert.False(t, created.Sys.PublishedVersion.IsSet())
	assert.False(t, created.Sys.PublishedAt.IsSet())

	staleResponse, err := handler.ActivateContentType(context.Background(), contentTypeActivateParams(0))
	require.NoError(t, err)
	requireContentfulConflictWithNonemptyMessage(t, staleResponse, cm.ErrorSysIDVersionMismatch)

	activated := activateContentType(t, handler, created.Sys.Version)
	assert.Equal(t, 2, activated.Sys.Version)
	assert.Equal(t, 1, activated.Sys.PublishedVersion.Or(0))
	assert.True(t, activated.Sys.PublishedAt.IsSet())
}

func TestDeactivateAndDeleteContentTypeUsesContentfulLifecycle(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	request := newContentTypeRequest()
	created := createContentType(t, handler, &request)
	activated := activateContentType(t, handler, created.Sys.Version)

	deletePublishedResponse, err := handler.DeleteContentType(context.Background(), contentTypeDeleteParams())
	require.NoError(t, err)
	requireContentfulError(t, deletePublishedResponse, http.StatusBadRequest, "BadRequest", "Cannot delete published")

	deactivateResponse, err := handler.DeactivateContentType(context.Background(), contentTypeDeactivateParams())
	require.NoError(t, err)

	deactivated, deactivatedOK := deactivateResponse.(*cm.ContentType)
	require.True(t, deactivatedOK)
	assert.Equal(t, activated.Sys.Version+1, deactivated.Sys.Version)
	assert.False(t, deactivated.Sys.PublishedVersion.IsSet())
	assert.False(t, deactivated.Sys.PublishedAt.IsSet())

	deactivateAgainResponse, err := handler.DeactivateContentType(context.Background(), contentTypeDeactivateParams())
	require.NoError(t, err)
	requireContentfulError(t, deactivateAgainResponse, http.StatusBadRequest, "BadRequest", "Not published")

	deleteResponse, err := handler.DeleteContentType(context.Background(), contentTypeDeleteParams())
	require.NoError(t, err)

	_, deleted := deleteResponse.(*cm.NoContent)
	require.True(t, deleted)

	editorResponse, err := handler.GetEditorInterface(context.Background(), contentTypeEditorInterfaceParams())
	require.NoError(t, err)
	requireContentfulError(t, editorResponse, http.StatusNotFound, cm.ErrorSysIDNotFound, "EditorInterface not found")
}

func TestPutContentTypeRejectsRemovingPublishedNonOmittedField(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	request := contentTypeRequestWithRemovableField()
	created := createContentType(t, handler, &request)
	activated := activateContentType(t, handler, created.Sys.Version)

	request.Fields = request.Fields[:1]
	response, err := handler.PutContentType(context.Background(), &request, contentTypePutParams(activated.Sys.Version))
	require.NoError(t, err)
	requireContentfulError(
		t,
		response,
		http.StatusBadRequest,
		"BadRequest",
		"You need to omit a field before deleting it",
	)

	stored := getContentType(t, handler)
	assert.Equal(t, 2, stored.Sys.Version)
	assert.Equal(t, 1, stored.Sys.PublishedVersion.Or(0))
	require.Len(t, stored.Fields, 2)
	assert.Equal(t, "obsolete", stored.Fields[1].ID)
	assert.False(t, stored.Fields[1].Omitted.Or(false))
}

func TestPutContentTypeRejectsRemovingFieldOmittedOnlyInDraft(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	request := contentTypeRequestWithRemovableField()
	created := createContentType(t, handler, &request)
	activated := activateContentType(t, handler, created.Sys.Version)

	request.Fields[1].Omitted = cm.NewOptBool(true)
	draft := putContentType(t, handler, &request, activated.Sys.Version)
	request.Fields = request.Fields[:1]
	response, err := handler.PutContentType(context.Background(), &request, contentTypePutParams(draft.Sys.Version))
	require.NoError(t, err)
	requireContentfulError(
		t,
		response,
		http.StatusBadRequest,
		"BadRequest",
		"You need to omit a field before deleting it",
	)

	stored := getContentType(t, handler)
	assert.Equal(t, 3, stored.Sys.Version)
	assert.Equal(t, 1, stored.Sys.PublishedVersion.Or(0))
	require.Len(t, stored.Fields, 2)
	assert.Equal(t, "obsolete", stored.Fields[1].ID)
	assert.True(t, stored.Fields[1].Omitted.Or(false))
}

func TestPutContentTypeAllowsRemovingFieldOmittedInPublishedVersion(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	request := contentTypeRequestWithRemovableField()
	created := createContentType(t, handler, &request)
	activated := activateContentType(t, handler, created.Sys.Version)

	request.Fields[1].Omitted = cm.NewOptBool(true)
	draft := putContentType(t, handler, &request, activated.Sys.Version)
	activated = activateContentType(t, handler, draft.Sys.Version)

	request.Fields = request.Fields[:1]
	updated := putContentType(t, handler, &request, activated.Sys.Version)
	assert.Equal(t, 5, updated.Sys.Version)
	assert.Equal(t, 3, updated.Sys.PublishedVersion.Or(0))
	require.Len(t, updated.Fields, 1)
	assert.Equal(t, "title", updated.Fields[0].ID)
}

func TestPutContentTypeAllowsRemovingNeverPublishedDraftField(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	request := newContentTypeRequest()
	created := createContentType(t, handler, &request)
	activated := activateContentType(t, handler, created.Sys.Version)

	request.Fields = append(request.Fields, cm.ContentTypeRequestDataFieldsItem{
		ID: "draft-only", Name: "Draft only", Type: "Symbol",
	})
	draft := putContentType(t, handler, &request, activated.Sys.Version)
	request.Fields = request.Fields[:1]
	updated := putContentType(t, handler, &request, draft.Sys.Version)

	assert.Equal(t, 4, updated.Sys.Version)
	assert.Equal(t, 1, updated.Sys.PublishedVersion.Or(0))
	require.Len(t, updated.Fields, 1)
	assert.Equal(t, "title", updated.Fields[0].ID)
}

func TestPutContentTypeAllowsRemovingNonOmittedFieldAfterDeactivation(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	request := contentTypeRequestWithRemovableField()
	created := createContentType(t, handler, &request)
	activateContentType(t, handler, created.Sys.Version)

	response, err := handler.DeactivateContentType(context.Background(), contentTypeDeactivateParams())
	require.NoError(t, err)

	deactivated, ok := response.(*cm.ContentType)
	require.True(t, ok)
	assert.Equal(t, 3, deactivated.Sys.Version)
	assert.False(t, deactivated.Sys.PublishedVersion.IsSet())

	request.Fields = request.Fields[:1]
	updated := putContentType(t, handler, &request, deactivated.Sys.Version)
	assert.Equal(t, 4, updated.Sys.Version)
	assert.False(t, updated.Sys.PublishedVersion.IsSet())
	require.Len(t, updated.Fields, 1)
	assert.Equal(t, "title", updated.Fields[0].ID)
}

func contentTypeRequestWithRemovableField() cm.ContentTypeRequestData {
	request := newContentTypeRequest()
	request.Fields = append(request.Fields, cm.ContentTypeRequestDataFieldsItem{
		ID: "obsolete", Name: "Obsolete", Type: "Symbol",
	})

	return request
}

func getContentType(t *testing.T, handler *cmt.Handler) *cm.ContentType {
	t.Helper()

	response, err := handler.GetContentType(context.Background(), cm.GetContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "content-type",
	})
	require.NoError(t, err)

	contentType, ok := response.(*cm.ContentType)
	require.True(t, ok)

	return contentType
}

func contentTypeActivateParams(version int) cm.ActivateContentTypeParams {
	return cm.ActivateContentTypeParams{
		SpaceID:            "space",
		EnvironmentID:      "environment",
		ContentTypeID:      "content-type",
		XContentfulVersion: version,
	}
}

func contentTypeDeactivateParams() cm.DeactivateContentTypeParams {
	return cm.DeactivateContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "content-type",
	}
}

func contentTypeDeleteParams() cm.DeleteContentTypeParams {
	return cm.DeleteContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "content-type",
	}
}

func activateContentType(t *testing.T, handler *cmt.Handler, version int) cm.ContentType {
	t.Helper()

	response, err := handler.ActivateContentType(context.Background(), contentTypeActivateParams(version))
	require.NoError(t, err)

	return requireContentTypeStatusCode(t, response, http.StatusOK)
}
