package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntryRequestDistinguishesTerraformNullFromJSONNull(t *testing.T) {
	t.Parallel()

	model := EntryModel{
		Fields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
			"terraform_null": NewTypedMapNull[jsontypes.Normalized](),
			"json_null": NewTypedMap(map[string]jsontypes.Normalized{
				"en-US": NewNormalizedJSONTypesNormalizedValue([]byte("null")),
			}),
			"value": NewTypedMap(map[string]jsontypes.Normalized{
				"en-US": NewNormalizedJSONTypesNormalizedValue([]byte(`"value"`)),
			}),
		}),
		Metadata: NewTypedObjectNull[EntryMetadataValue](),
	}

	request, diags := model.ToEntryRequest(t.Context())
	require.False(t, diags.HasError(), diags.Errors())

	fields, ok := request.Fields.Get()
	require.True(t, ok)
	assert.NotContains(t, fields, "terraform_null")
	assert.JSONEq(t, `{"en-US":null}`, string(fields["json_null"]))
	assert.JSONEq(t, `{"en-US":"value"}`, string(fields["value"]))
}

func TestEntryRequestUnknownFieldFailsWithoutPartialOutput(t *testing.T) {
	t.Parallel()

	model := EntryModel{
		Fields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
			"known": NewTypedMap(map[string]jsontypes.Normalized{
				"en-US": NewNormalizedJSONTypesNormalizedValue([]byte(`"value"`)),
			}),
			"unknown": NewTypedMapUnknown[jsontypes.Normalized](),
		}),
		Metadata: knownEntryMetadata(),
	}

	request, diags := model.ToEntryRequest(t.Context())
	assert.Equal(t, cm.EntryRequest{}, request)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{`fields["unknown"]`}, attributeDiagnosticPaths(t, diags))
}

func TestEntryRequestUnresolvedFieldsContainerFailsWithoutPartialOutput(t *testing.T) {
	t.Parallel()

	for name, fields := range map[string]TypedMap[TypedMap[jsontypes.Normalized]]{
		"null":    NewTypedMapNull[TypedMap[jsontypes.Normalized]](),
		"unknown": NewTypedMapUnknown[TypedMap[jsontypes.Normalized]](),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := EntryModel{
				Fields:   fields,
				Metadata: knownEntryMetadata(),
			}

			request, diags := model.ToEntryRequest(t.Context())
			assert.Equal(t, cm.EntryRequest{}, request)
			require.True(t, diags.HasError())
			assert.Equal(t, []string{"fields"}, attributeDiagnosticPaths(t, diags))
		})
	}
}

func TestEntryRequestOmitsUnresolvedOptionalComputedMetadataContainer(t *testing.T) {
	t.Parallel()

	for name, metadata := range map[string]TypedObject[EntryMetadataValue]{
		"null":    NewTypedObjectNull[EntryMetadataValue](),
		"unknown": NewTypedObjectUnknown[EntryMetadataValue](),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := EntryModel{
				Fields:   NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
				Metadata: metadata,
			}

			request, diags := model.ToEntryRequest(t.Context())
			require.False(t, diags.HasError(), diags.Errors())
			assert.Equal(t, cm.OptEntryMetadata{}, request.Metadata)
			fields, ok := request.Fields.Get()
			require.True(t, ok)
			assert.Empty(t, fields)
		})
	}
}

func TestEntryRequestInvalidMetadataChildrenFailWithoutPartialOutput(t *testing.T) {
	t.Parallel()

	for name, metadata := range map[string]TypedObject[EntryMetadataValue]{
		"null concept": NewTypedObject(EntryMetadataValue{
			Concepts: NewTypedList([]types.String{types.StringValue("concept"), types.StringNull()}),
			Tags:     NewTypedList([]types.String{types.StringValue("tag")}),
		}),
		"unknown concept": NewTypedObject(EntryMetadataValue{
			Concepts: NewTypedList([]types.String{types.StringValue("concept"), types.StringUnknown()}),
			Tags:     NewTypedList([]types.String{types.StringValue("tag")}),
		}),
		"null tag": NewTypedObject(EntryMetadataValue{
			Concepts: NewTypedList([]types.String{types.StringValue("concept")}),
			Tags:     NewTypedList([]types.String{types.StringValue("tag"), types.StringNull()}),
		}),
		"unknown tag": NewTypedObject(EntryMetadataValue{
			Concepts: NewTypedList([]types.String{types.StringValue("concept")}),
			Tags:     NewTypedList([]types.String{types.StringValue("tag"), types.StringUnknown()}),
		}),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := EntryModel{
				Fields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
					"known": NewTypedMap(map[string]jsontypes.Normalized{
						"en-US": NewNormalizedJSONTypesNormalizedValue([]byte(`"value"`)),
					}),
				}),
				Metadata: metadata,
			}

			request, diags := model.ToEntryRequest(t.Context())
			assert.Equal(t, cm.EntryRequest{}, request)
			require.True(t, diags.HasError())

			expectedPath := "metadata.concepts[1]"
			if name == "null tag" || name == "unknown tag" {
				expectedPath = "metadata.tags[1]"
			}

			assert.Equal(t, []string{expectedPath}, attributeDiagnosticPaths(t, diags))
		})
	}
}

func TestEntryRequestKnownMetadata(t *testing.T) {
	t.Parallel()

	model := EntryModel{
		Fields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
			"known": NewTypedMap(map[string]jsontypes.Normalized{
				"en-US": NewNormalizedJSONTypesNormalizedValue([]byte(`"value"`)),
			}),
		}),
		Metadata: knownEntryMetadata(),
	}

	request, diags := model.ToEntryRequest(t.Context())
	require.False(t, diags.HasError(), diags.Errors())

	metadata, ok := request.Metadata.Get()
	require.True(t, ok)
	require.Len(t, metadata.Concepts, 1)
	assert.Equal(t, "concept", metadata.Concepts[0].Sys.ID)
	require.Len(t, metadata.Tags, 1)
	assert.Equal(t, "tag", metadata.Tags[0].Sys.ID)
}

func knownEntryMetadata() TypedObject[EntryMetadataValue] {
	return NewTypedObject(EntryMetadataValue{
		Concepts: NewTypedList([]types.String{types.StringValue("concept")}),
		Tags:     NewTypedList([]types.String{types.StringValue("tag")}),
	})
}
