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
	created := putContentType(t, handler, &request, 1, http.StatusCreated)
	assert.Equal(t, 1, created.Sys.Version)
	assert.False(t, created.Sys.PublishedVersion.IsSet())
	assert.False(t, created.Sys.PublishedAt.IsSet())

	staleResponse, err := handler.ActivateContentType(context.Background(), contentTypeActivateParams(0))
	require.NoError(t, err)
	requireContentfulError(t, staleResponse, http.StatusConflict, cm.ErrorSysIDVersionMismatch, "")

	activated := activateContentType(t, handler, created.Sys.Version)
	assert.Equal(t, 2, activated.Sys.Version)
	assert.Equal(t, 1, activated.Sys.PublishedVersion.Or(0))
	assert.True(t, activated.Sys.PublishedAt.IsSet())
}

func TestDeactivateAndDeleteContentTypeUsesContentfulLifecycle(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	request := newContentTypeRequest()
	created := putContentType(t, handler, &request, 1, http.StatusCreated)
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
