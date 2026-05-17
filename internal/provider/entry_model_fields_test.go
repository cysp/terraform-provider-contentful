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

func TestEntryModelToOptEntryFieldsSkipsNullUnknownAndEncodesLocalizedValues(t *testing.T) {
	t.Parallel()

	fields, diags := entryModelToOptEntryFields(context.Background(), EntryModel{
		Fields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
			"title": NewTypedMap(map[string]jsontypes.Normalized{
				"en-AU": NewNormalizedJSONTypesNormalizedValue([]byte(`"Title"`)),
				"en-US": jsontypes.NewNormalizedNull(),
				"fr-FR": jsontypes.NewNormalizedUnknown(),
			}),
			"ignored-null":    NewTypedMapNull[jsontypes.Normalized](),
			"ignored-unknown": NewTypedMapUnknown[jsontypes.Normalized](),
		}),
	})

	require.False(t, diags.HasError())
	require.True(t, fields.IsSet())
	assert.JSONEq(t, `{"en-AU":"Title"}`, string(fields.Value["title"]))
	assert.NotContains(t, fields.Value, "ignored-null")
	assert.NotContains(t, fields.Value, "ignored-unknown")
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
	require.True(t, fields.IsSet())
	assert.Empty(t, fields.Value)
	require.Len(t, diags, 1)

	diagWithPath, ok := diags[0].(diag.DiagnosticWithPath)
	require.True(t, ok)
	assert.Equal(t, path.Root("fields").AtMapKey("title").AtMapKey("en-AU").String(), diagWithPath.Path().String())
}

func TestNewEntryFieldsFromResponseReturnsEmptyMapForUnsetFields(t *testing.T) {
	t.Parallel()

	fields, diags := NewEntryFieldsFromResponse(context.Background(), path.Root("fields"), cm.OptEntryFields{})

	require.False(t, diags.HasError())
	assert.False(t, fields.IsNull())
	assert.Empty(t, fields.Elements())
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
