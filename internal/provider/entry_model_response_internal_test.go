package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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
	assert.False(t, fields.Elements()["optional"].IsNull())
	assert.Equal(t, `null`, fields.Elements()["optional"].ValueString())
}

func TestMergeEntryFieldsWithFallback(t *testing.T) {
	t.Parallel()

	returnedValue := jsontypes.NewNormalizedValue(`{"en-US":"returned"}`)
	configuredValue := jsontypes.NewNormalizedValue(`{"en-US":"configured"}`)
	missingValue := jsontypes.NewNormalizedValue(`{"en-US":"missing"}`)
	returned := NewTypedMap(map[string]jsontypes.Normalized{"returned": returnedValue})
	configured := NewTypedMap(map[string]jsontypes.Normalized{
		"returned": configuredValue,
		"missing":  missingValue,
	})

	actual := mergeEntryFieldsWithFallback(returned, configured)

	assert.Equal(t, map[string]jsontypes.Normalized{
		"returned": returnedValue,
		"missing":  missingValue,
	}, actual.Elements())
	assert.Equal(t, map[string]jsontypes.Normalized{"returned": returnedValue}, returned.Elements())
	assert.Equal(t, map[string]jsontypes.Normalized{
		"returned": configuredValue,
		"missing":  missingValue,
	}, configured.Elements())
}

func TestMergeEntryFieldsWithFallbackInitializesNullResponse(t *testing.T) {
	t.Parallel()

	value := jsontypes.NewNormalizedValue(`{"en-US":"configured"}`)
	actual := mergeEntryFieldsWithFallback(
		NewTypedMapNull[jsontypes.Normalized](),
		NewTypedMap(map[string]jsontypes.Normalized{"field": value}),
	)

	assert.False(t, actual.IsNull())
	assert.Equal(t, map[string]jsontypes.Normalized{"field": value}, actual.Elements())
}
