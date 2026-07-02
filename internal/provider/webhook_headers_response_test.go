package provider_test

import (
	"context"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadHeaderValueMapFromResponsePreservesNullEmptyHeaders(t *testing.T) {
	t.Parallel()

	headers, diags := ReadHeaderValueMapFromResponse(
		context.Background(),
		path.Root("headers"),
		[]cm.WebhookDefinitionHeader{},
		NewTypedMapNull[TypedObject[WebhookHeaderValue]](),
	)

	require.False(t, diags.HasError(), diags)
	assert.True(t, headers.IsNull())
}
