package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
)

func TestContentTypeMetadataTaxonomyResponseRejectsUnknownLinkType(t *testing.T) {
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

	assert.True(t, diags.HasError())
	assert.True(t, actual.IsNull())
	assert.Equal(t, []string{itemPath.String()}, diagnosticPaths(t, diags))
}
