package cmtesting_test

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
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

			response, err := handler.PutContentType(t.Context(), &request, contentTypeCreateParams())
			require.NoError(t, err)
			requireContentfulError(t, response, http.StatusUnprocessableEntity, "ValidationFailed", "Validation error")

			getResponse, err := handler.GetContentType(t.Context(), cm.GetContentTypeParams{
				SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "content-type",
			})
			require.NoError(t, err)
			requireContentfulError(t, getResponse, http.StatusNotFound, cm.ErrorSysIDNotFound, "ContentType not found")
		})
	}
}
