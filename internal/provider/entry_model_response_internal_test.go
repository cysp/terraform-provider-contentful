package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntryFieldsFromResponsePreservesOmissionAsNull(t *testing.T) {
	t.Parallel()

	fields, diags := NewEntryFieldsFromResponse(t.Context(), path.Root("fields"), cm.OptEntryFields{})

	require.False(t, diags.HasError(), diags.Errors())
	assert.True(t, fields.IsNull())
	assert.False(t, fields.IsUnknown())
	assert.Empty(t, fields.Elements())
}

func TestNewEntryFieldsFromResponsePreservesKnownJSONNull(t *testing.T) {
	t.Parallel()

	fields, diags := NewEntryFieldsFromResponse(t.Context(), path.Root("fields"), cm.NewOptEntryFields(cm.EntryFields{
		"optional": jx.Raw(`null`),
	}))

	require.False(t, diags.HasError(), diags.Errors())
	require.False(t, fields.IsNull())
	require.Contains(t, fields.Elements(), "optional")
	assert.True(t, fields.Elements()["optional"].IsNull())
}
