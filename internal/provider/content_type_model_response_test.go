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

func TestContentTypeResponseProjectionIncludesExactPublishedVersion(t *testing.T) {
	t.Parallel()

	contentType := cmt.NewContentTypeFromRequestFields("space", "environment", "content-type", cm.ContentTypeRequestData{
		Name: "Content type",
	})
	contentType.Sys.Version = 8
	contentType.Sys.PublishedVersion.SetTo(7)

	model, diagnostics := NewContentTypeResourceModelFromResponse(t.Context(), contentType)
	require.False(t, diagnostics.HasError())
	assert.Equal(t, types.Int64Value(7), model.PublishedVersion)
}
