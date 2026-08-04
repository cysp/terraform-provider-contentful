package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntryModelToOptEntryFieldsRejectsUnresolvedContainer(t *testing.T) {
	t.Parallel()

	for name, fields := range map[string]TypedMap[jsontypes.Normalized]{
		"null":    NewTypedMapNull[jsontypes.Normalized](),
		"unknown": NewTypedMapUnknown[jsontypes.Normalized](),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := entryModelToOptEntryFields(t.Context(), EntryModel{Fields: fields})

			assert.Equal(t, cm.OptEntryFields{}, result)
			require.True(t, diags.HasError())
			assert.Equal(t, []string{"fields"}, requestDiagnosticPaths(t, diags))
		})
	}
}

func TestEntryModelToOptEntryFieldsChildSemantics(t *testing.T) {
	t.Parallel()

	result, diags := entryModelToOptEntryFields(t.Context(), EntryModel{
		Fields: NewTypedMap(map[string]jsontypes.Normalized{
			"terraform_null": jsontypes.NewNormalizedNull(),
			"json_null":      NewNormalizedJSONTypesNormalizedValue([]byte("null")),
		}),
	})

	require.False(t, diags.HasError(), diags.Errors())

	fields, ok := result.Get()
	require.True(t, ok)
	assert.NotContains(t, fields, "terraform_null")
	assert.JSONEq(t, "null", string(fields["json_null"]))
}

func TestEntryModelToOptEntryFieldsRejectsUnknownChild(t *testing.T) {
	t.Parallel()

	result, diags := entryModelToOptEntryFields(t.Context(), EntryModel{
		Fields: NewTypedMap(map[string]jsontypes.Normalized{
			"known":   NewNormalizedJSONTypesNormalizedValue([]byte(`"value"`)),
			"unknown": jsontypes.NewNormalizedUnknown(),
		}),
	})

	assert.Equal(t, cm.OptEntryFields{}, result)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{`fields["unknown"]`}, requestDiagnosticPaths(t, diags))
}
