package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentTypeResponseProjectionIncludesPublicationState(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		publishedVersion *int
		expected         types.Int64
	}{
		"published":   {publishedVersion: new(7), expected: types.Int64Value(7)},
		"unpublished": {publishedVersion: nil, expected: types.Int64Null()},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			contentType := cmt.NewContentTypeFromRequestFields("space", "environment", "content-type", cm.ContentTypeRequestData{
				Name: "Content type",
			})
			contentType.Sys.Version = 8

			if test.publishedVersion != nil {
				contentType.Sys.PublishedVersion.SetTo(*test.publishedVersion)
			}

			model, diagnostics := NewContentTypeResourceModelFromResponse(t.Context(), contentType)
			require.False(t, diagnostics.HasError())
			assert.Equal(t, test.expected, model.PublishedVersion)
		})
	}
}
