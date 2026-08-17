package provider_test

import (
	"bytes"
	"maps"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentTypeMetadataTaxonomyResponsePreservesUnknownLinkTypeAsKnownSentinel(t *testing.T) {
	t.Parallel()

	itemPath := path.Root("metadata").AtName("taxonomy").AtListIndex(0)
	actual, diags := NewContentTypeMetadataTaxonomyItemFromResponse(
		t.Context(),
		itemPath,
		cm.ContentTypeMetadataTaxonomyItem{
			Sys: cm.ContentTypeMetadataTaxonomyItemSys{
				Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
				ID:       "taxonomy",
				LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkType("unknown"),
			},
		},
	)

	assert.False(t, diags.HasError())
	require.Len(t, diags.Warnings(), 1)
	diagnostic, ok := diags.Warnings()[0].(diag.DiagnosticWithPath)
	require.True(t, ok)
	assert.Equal(t, itemPath, diagnostic.Path())
	assert.False(t, actual.IsNull())
	assert.True(t, actual.Value().TaxonomyConcept.IsNull())
	assert.True(t, actual.Value().TaxonomyConceptScheme.IsNull())

	_, requestDiags := actual.Value().ToContentTypeMetadataTaxonomyItem(itemPath)
	assert.True(t, requestDiags.HasError())
	assert.Equal(t, []string{itemPath.String()}, diagnosticPaths(t, requestDiags))
}

func TestContentTypeMetadataTaxonomyResponseUnknownLinkTypeIsRejectedAtHTTPBoundary(t *testing.T) {
	t.Parallel()

	contentfulServer, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	contentfulServer.RegisterSpaceEnvironment("space", "environment")
	contentfulServer.SetContentType("space", "environment", "content-type", cm.ContentTypeRequestData{
		Name:   "Content type",
		Fields: []cm.ContentTypeRequestDataFieldsItem{},
		Metadata: cm.NewOptContentTypeMetadata(cm.ContentTypeMetadata{
			Taxonomy: []cm.ContentTypeMetadataTaxonomyItem{
				{
					Sys: cm.ContentTypeMetadataTaxonomyItemSys{
						Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
						ID:       "taxonomy",
						LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConcept,
					},
				},
			},
		}),
	})

	var replacedUnknownLinkType atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		recorder := httptest.NewRecorder()
		contentfulServer.ServeHTTP(recorder, request)

		responseBody := bytes.Replace(
			recorder.Body.Bytes(),
			[]byte(`"linkType":"TaxonomyConcept"`),
			[]byte(`"linkType":"FutureTaxonomyType"`),
			1,
		)
		replacedUnknownLinkType.Store(bytes.Contains(responseBody, []byte(`"linkType":"FutureTaxonomyType"`)))

		maps.Copy(responseWriter.Header(), recorder.Header())

		responseWriter.WriteHeader(recorder.Code)
		_, _ = responseWriter.Write(responseBody)
	}))
	t.Cleanup(server.Close)

	client, err := cm.NewClient(
		server.URL,
		cm.NewAccessTokenSecuritySource("CFPAT-12345"),
		cm.WithClient(server.Client()),
	)
	require.NoError(t, err)

	response, err := client.GetContentType(t.Context(), cm.GetContentTypeParams{
		SpaceID:       "space",
		EnvironmentID: "environment",
		ContentTypeID: "content-type",
	})

	assert.True(t, replacedUnknownLinkType.Load())
	assert.Nil(t, response)
	assert.ErrorContains(t, err, `invalid value: FutureTaxonomyType`)
}
