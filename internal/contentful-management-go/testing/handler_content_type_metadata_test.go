package cmtesting_test

import (
	"context"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/go-faster/jx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutContentTypeClearsAnnotationsAndPreservesTaxonomyWhenMetadataOmitted(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	request := newContentTypeRequest()
	request.Metadata.SetTo(cm.ContentTypeMetadata{
		Annotations: jx.Raw(`{"ContentType":[]}`),
		Taxonomy: []cm.ContentTypeMetadataTaxonomyItem{{
			Sys: cm.ContentTypeMetadataTaxonomyItemSys{
				Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
				LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConceptScheme,
				ID:       "furniture",
			},
		}},
	})
	created := createContentType(t, handler, &request)

	request.Metadata.Reset()
	updated := putContentType(t, handler, &request, created.Sys.Version)
	metadata, metadataOK := updated.Metadata.Get()
	require.True(t, metadataOK)
	assert.Nil(t, metadata.Annotations)
	require.Len(t, metadata.Taxonomy, 1)
	assert.Equal(t, "furniture", metadata.Taxonomy[0].Sys.ID)
}

func TestPutContentTypePreservesTaxonomyWhenMetadataOmitsTaxonomy(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	request := newContentTypeRequest()
	request.Metadata.SetTo(cm.ContentTypeMetadata{
		Taxonomy: []cm.ContentTypeMetadataTaxonomyItem{{
			Sys: cm.ContentTypeMetadataTaxonomyItemSys{
				Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
				LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConceptScheme,
				ID:       "furniture",
			},
		}},
	})
	created := createContentType(t, handler, &request)

	request.Metadata.SetTo(cm.ContentTypeMetadata{Annotations: jx.Raw(`{"ContentType":[]}`)})
	updated := putContentType(t, handler, &request, created.Sys.Version)
	metadata, metadataOK := updated.Metadata.Get()
	require.True(t, metadataOK)
	assert.JSONEq(t, `{"ContentType":[]}`, string(metadata.Annotations))
	require.Len(t, metadata.Taxonomy, 1)
	assert.Equal(t, "furniture", metadata.Taxonomy[0].Sys.ID)
}

func TestPutContentTypeReplacesMetadataWhenPresent(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	request := newContentTypeRequest()
	request.Metadata.SetTo(cm.ContentTypeMetadata{
		Annotations: jx.Raw(`{"ContentType":[]}`),
		Taxonomy: []cm.ContentTypeMetadataTaxonomyItem{{
			Sys: cm.ContentTypeMetadataTaxonomyItemSys{
				Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
				LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConceptScheme,
				ID:       "furniture",
			},
		}},
	})
	created := createContentType(t, handler, &request)

	request.Metadata.SetTo(cm.ContentTypeMetadata{Taxonomy: []cm.ContentTypeMetadataTaxonomyItem{}})
	updated := putContentType(t, handler, &request, created.Sys.Version)
	metadata, metadataOK := updated.Metadata.Get()
	require.True(t, metadataOK)
	assert.Nil(t, metadata.Annotations)
	assert.Empty(t, metadata.Taxonomy)
	assert.NotNil(t, metadata.Taxonomy)
}

func TestPutContentTypeAcceptsTaxonomyOnlyEmptyList(t *testing.T) {
	t.Parallel()

	handler := newContentTypeTestHandler()
	request := newContentTypeRequest()
	request.Metadata.SetTo(cm.ContentTypeMetadata{Taxonomy: []cm.ContentTypeMetadataTaxonomyItem{}})

	created := createContentType(t, handler, &request)
	metadata, metadataOK := created.Metadata.Get()
	require.True(t, metadataOK)
	assert.Nil(t, metadata.Annotations)
	assert.Empty(t, metadata.Taxonomy)
	assert.NotNil(t, metadata.Taxonomy)
}

func TestPutContentTypeRejectsEmptyMetadataObject(t *testing.T) {
	t.Parallel()

	testCases := map[string]cm.ContentTypeMetadata{
		"absent properties":                     {},
		"empty annotations":                     {Annotations: jx.Raw(`{ }`)},
		"empty annotations with empty taxonomy": {Annotations: jx.Raw(`{ }`), Taxonomy: []cm.ContentTypeMetadataTaxonomyItem{}},
		"null annotations":                      {Annotations: jx.Raw(`null`)},
	}

	for name, metadata := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			handler := newContentTypeTestHandler()
			request := newContentTypeRequest()
			request.Metadata.SetTo(metadata)

			response, err := handler.PutContentType(context.Background(), &request, contentTypeCreateParams())
			require.NoError(t, err)
			requireContentfulError(t, response, http.StatusUnprocessableEntity, "ValidationFailed", "Validation error")

			getResponse, err := handler.GetContentType(context.Background(), cm.GetContentTypeParams{
				SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "content-type",
			})
			require.NoError(t, err)
			requireContentfulError(t, getResponse, http.StatusNotFound, cm.ErrorSysIDNotFound, "ContentType not found")
		})
	}
}

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

	response, err := handler.PutContentType(context.Background(), request, contentTypeCreateParams())
	require.NoError(t, err)

	return requireContentTypeStatusCode(t, response, http.StatusCreated)
}

func putContentType(t *testing.T, handler *cmt.Handler, request *cm.ContentTypeRequestData, version int) cm.ContentType {
	t.Helper()

	response, err := handler.PutContentType(context.Background(), request, contentTypePutParams(version))
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

func requireContentfulError(t *testing.T, response any, expectedStatus int, expectedID, expectedMessage string) {
	t.Helper()

	statusCode, ok := response.(*cm.ErrorStatusCode)
	require.True(t, ok)
	assert.Equal(t, expectedStatus, statusCode.StatusCode)
	errorResponse, ok := statusCode.Response.GetError()
	require.True(t, ok)
	assert.Equal(t, expectedID, errorResponse.Sys.ID)
	assert.Equal(t, expectedMessage, errorResponse.Message.Or(""))
}

func requireContentfulConflictWithNonemptyMessage(t *testing.T, response any, expectedID string) {
	t.Helper()

	statusCode, ok := response.(*cm.ErrorStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusConflict, statusCode.StatusCode)
	errorResponse, ok := statusCode.Response.GetError()
	require.True(t, ok)
	assert.Equal(t, cm.ErrorSysTypeError, errorResponse.Sys.Type)
	assert.Equal(t, expectedID, errorResponse.Sys.ID)
	assert.NotEmpty(t, errorResponse.Message.Or(""))
}
