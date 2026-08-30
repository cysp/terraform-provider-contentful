//nolint:testpackage
package provider

import (
	"context"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntryModelToOptEntryFieldsPreservesNullValues(t *testing.T) {
	t.Parallel()

	fields, diags := entryModelToOptEntryFields(context.Background(), EntryModel{
		Fields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
			"title": NewTypedMap(map[string]jsontypes.Normalized{
				"en-AU": NewNormalizedJSONTypesNormalizedValue([]byte(`"Title"`)),
				"en-US": jsontypes.NewNormalizedNull(),
			}),
			"ignored-null": NewTypedMapNull[jsontypes.Normalized](),
		}),
	})

	require.False(t, diags.HasError())
	require.True(t, fields.IsSet())
	assert.JSONEq(t, `{"en-AU":"Title"}`, string(fields.Value["title"]))
	assert.NotContains(t, fields.Value, "ignored-null")
}

func TestEntryModelToOptEntryFieldsRejectsUnknownValuesWithoutPartialOutput(t *testing.T) {
	t.Parallel()

	fields, diags := entryModelToOptEntryFields(context.Background(), EntryModel{
		Fields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
			"title": NewTypedMap(map[string]jsontypes.Normalized{
				"en-AU": NewNormalizedJSONTypesNormalizedValue([]byte(`"Title"`)),
				"fr-FR": jsontypes.NewNormalizedUnknown(),
			}),
			"ignored-unknown": NewTypedMapUnknown[jsontypes.Normalized](),
		}),
	})

	assert.False(t, fields.IsSet())
	require.True(t, diags.HasError())

	paths := make([]string, 0, len(diags))
	for _, diagnostic := range diags {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok)

		paths = append(paths, withPath.Path().String())
	}

	assert.ElementsMatch(t, []string{
		`fields["title"]["fr-FR"]`,
		`fields["ignored-unknown"]`,
	}, paths)
}

func TestEntryModelToOptEntryFieldsOmitsNullFieldValue(t *testing.T) {
	t.Parallel()

	fields, diags := entryModelToOptEntryFields(context.Background(), EntryModel{
		Fields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
			"title": NewTypedMapNull[jsontypes.Normalized](),
		}),
	})

	require.False(t, diags.HasError())
	require.True(t, fields.IsSet())
	assert.NotContains(t, fields.Value, "title")
}

func TestEntryModelToOptEntryFieldsReportsInvalidLocalizedJSON(t *testing.T) {
	t.Parallel()

	fields, diags := entryModelToOptEntryFields(context.Background(), EntryModel{
		Fields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
			"title": NewTypedMap(map[string]jsontypes.Normalized{
				"en-AU": NewNormalizedJSONTypesNormalizedValue([]byte(`invalid`)),
			}),
		}),
	})

	require.True(t, diags.HasError())
	assert.False(t, fields.IsSet())
	require.Len(t, diags, 1)

	diagWithPath, ok := diags[0].(diag.DiagnosticWithPath)
	require.True(t, ok)
	assert.Equal(t, path.Root("fields").AtMapKey("title").AtMapKey("en-AU").String(), diagWithPath.Path().String())
}

func TestNewEntryFieldsFromResponseReturnsNullForUnsetFields(t *testing.T) {
	t.Parallel()

	fields, diags := NewEntryFieldsFromResponse(context.Background(), path.Root("fields"), cm.OptEntryFields{})

	require.False(t, diags.HasError())
	assert.True(t, fields.IsNull())
}

func TestNewEntryFieldsFromResponseSkipsInvalidLocalizedFields(t *testing.T) {
	t.Parallel()

	fields, diags := NewEntryFieldsFromResponse(
		context.Background(),
		path.Root("fields"),
		cm.NewOptEntryFields(cm.EntryFields{
			"title": []byte(`[]`),
			"name":  []byte(`{"en-AU":"Name"}`),
		}),
	)

	require.True(t, diags.HasError())
	require.False(t, fields.IsNull())
	assert.NotContains(t, fields.Elements(), "title")
	assert.Equal(t, `"Name"`, fields.Elements()["name"].Elements()["en-AU"].ValueString())
}

func TestNewEntryLocalizedFieldFromRawReportsInvalidLocalizedJSON(t *testing.T) {
	t.Parallel()

	field, diags := NewEntryLocalizedFieldFromRaw(path.Root("fields").AtMapKey("title"), []byte(`invalid`))

	require.True(t, diags.HasError())
	assert.True(t, field.IsNull())
}

func TestNewEntryLocalizedFieldFromRawPreservesNull(t *testing.T) {
	t.Parallel()

	field, diags := NewEntryLocalizedFieldFromRaw(path.Root("fields").AtMapKey("title"), []byte(`null`))

	require.False(t, diags.HasError())
	assert.True(t, field.IsNull())
}
