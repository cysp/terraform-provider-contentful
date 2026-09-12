package cmtesting_test

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newContentTypeTestHandler() *cmt.Handler {
	handler := cmt.NewHandler()
	handler.RegisterSpaceEnvironment("space", "environment", "ready")

	return handler
}

func newContentTypeRequest() cm.ContentTypeRequestData {
	return cm.ContentTypeRequestData{
		Name:         "Content Type",
		DisplayField: "title",
		Fields:       []cm.ContentTypeRequestDataFieldsItem{{ID: "title", Name: "Title", Type: "Symbol"}},
	}
}

func contentTypePutParams(version int) cm.PutContentTypeParams {
	return cm.PutContentTypeParams{
		SpaceID:            "space",
		EnvironmentID:      "environment",
		ContentTypeID:      "content-type",
		XContentfulVersion: cm.NewOptInt(version),
	}
}

func contentTypeCreateParams() cm.PutContentTypeParams {
	return cm.PutContentTypeParams{
		SpaceID:       "space",
		EnvironmentID: "environment",
		ContentTypeID: "content-type",
	}
}

func createContentType(t *testing.T, handler *cmt.Handler, request *cm.ContentTypeRequestData) cm.ContentType {
	t.Helper()

	response, err := handler.PutContentType(t.Context(), request, contentTypeCreateParams())
	require.NoError(t, err)

	return requireContentTypeStatusCode(t, response, http.StatusCreated)
}

func putContentType(t *testing.T, handler *cmt.Handler, request *cm.ContentTypeRequestData, version int) cm.ContentType {
	t.Helper()

	response, err := handler.PutContentType(t.Context(), request, contentTypePutParams(version))
	require.NoError(t, err)

	return requireContentTypeStatusCode(t, response, http.StatusOK)
}

func requireContentTypeStatusCode(t *testing.T, response any, expectedStatus int) cm.ContentType {
	t.Helper()

	statusCode, ok := response.(*cm.ContentTypeStatusCode)
	require.True(t, ok)
	assert.Equal(t, expectedStatus, statusCode.StatusCode)

	return statusCode.Response
}
