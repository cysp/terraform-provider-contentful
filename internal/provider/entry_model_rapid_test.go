package provider_test

import (
	"context"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestEntryModelRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rapid.Check(t, func(t *rapid.T) {
		model := rapidEntryModelGenerator().Draw(t, "model")

		req, diags := model.ToEntryRequest(ctx)
		if diags.HasError() {
			t.Fatalf("ToEntryRequest failed: %v", diags.Errors())
		}

		entry := cm.Entry{
			Sys: cm.NewEntrySys(
				model.SpaceID.ValueString(),
				model.EnvironmentID.ValueString(),
				model.ContentTypeID.ValueString(),
				model.EntryID.ValueString(),
			),
			Fields:   req.Fields,
			Metadata: req.Metadata,
		}

		resultModel, diags := NewEntryResourceModelFromResponse(ctx, entry)
		if diags.HasError() {
			t.Fatalf("NewEntryResourceModelFromResponse failed: %v", diags.Errors())
		}

		assert.Equal(t, model, resultModel, "Fields and Metadata should be preserved")
		assert.Equal(t, model.EntryIdentityModel, resultModel.EntryIdentityModel, "Identity fields should be preserved")
		assert.Equal(t, model.ContentTypeID, resultModel.ContentTypeID, "ContentTypeID should be preserved")
		assert.Equal(t, model.ID, resultModel.ID, "ID should be preserved")
	})
}

// TestEntryModelRoundTrip_EmptyVsNilSlices tests the specific case of empty vs nil slices
// in the metadata concepts and tags fields to ensure proper round trip behavior.
func TestEntryModelRoundTrip_EmptyVsNilSlices(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Test case 1: nil concepts slice should remain null after round trip
	t.Run("nil_concepts_remains_null", func(t *testing.T) {
		metadata := cm.EntryMetadata{
			Concepts: nil, // nil slice
			Tags:     nil,
		}

		result, diags := NewEntryMetadataFromResponse(ctx, path.Root("metadata"), cm.NewOptEntryMetadata(metadata))
		require.False(t, diags.HasError(), "should not have errors")

		assert.True(t, result.Value().Concepts.IsNull(), "nil concepts should become null list")
		assert.True(t, result.Value().Tags.IsNull(), "nil tags should become null list")
	})

	// Test case 2: empty (non-nil) concepts slice - this is the problematic case
	t.Run("empty_concepts_roundtrip", func(t *testing.T) {
		metadata := cm.EntryMetadata{
			Concepts: []cm.TaxonomyConceptLink{}, // empty but non-nil slice
			Tags:     []cm.TagLink{},             // empty but non-nil slice
		}

		result, diags := NewEntryMetadataFromResponse(ctx, path.Root("metadata"), cm.NewOptEntryMetadata(metadata))
		require.False(t, diags.HasError(), "should not have errors")

		// Empty slices from the API should become known empty lists
		// to preserve the distinction from nil/absent fields
		isConceptsNull := result.Value().Concepts.IsNull()
		var conceptsLen int
		if !isConceptsNull {
			conceptsLen = len(result.Value().Concepts.Elements())
		}

		t.Logf("Empty concepts slice: IsNull=%v, Length=%d", isConceptsNull, conceptsLen)

		// Now convert back to request
		model := EntryModel{
			IDIdentityModel:    NewIDIdentityModelFromMultipartID("space1", "env1", "entry1"),
			EntryIdentityModel: EntryIdentityModel{SpaceID: types.StringValue("space1"), EnvironmentID: types.StringValue("env1"), EntryID: types.StringValue("entry1")},
			ContentTypeID:      types.StringValue("ct1"),
			Fields:             NewTypedMapNull[jsontypes.Normalized](),
			Metadata:           result,
		}

		req, diags := model.ToEntryRequest(ctx)
		require.False(t, diags.HasError(), "should not have errors converting to request")

		// Check what we got back
		conceptsAfterRoundTrip := req.Metadata.Value.Concepts
		tagsAfterRoundTrip := req.Metadata.Value.Tags

		t.Logf("After round trip: Concepts is nil=%v (len=%d), Tags is nil=%v (len=%d)",
			conceptsAfterRoundTrip == nil, len(conceptsAfterRoundTrip),
			tagsAfterRoundTrip == nil, len(tagsAfterRoundTrip))

		// Empty slices should round trip correctly: [] → known empty list → []
		// This preserves the distinction between empty and absent
		assert.False(t, isConceptsNull, "empty slice from API should become known empty list, not null")
		assert.NotNil(t, conceptsAfterRoundTrip, "empty list should round trip as empty slice, not nil")
		assert.Equal(t, 0, len(conceptsAfterRoundTrip), "should remain empty")
		
		assert.NotNil(t, tagsAfterRoundTrip, "empty list should round trip as empty slice, not nil")
		assert.Equal(t, 0, len(tagsAfterRoundTrip), "should remain empty")

		// Now do a second round trip to ensure consistency
		result2, diags := NewEntryMetadataFromResponse(ctx, path.Root("metadata"), req.Metadata)
		require.False(t, diags.HasError(), "should not have errors on second round trip")

		// The result should match the original
		assert.Equal(t, result.Value().Concepts.IsNull(), result2.Value().Concepts.IsNull(),
			"null/known state should be preserved across round trips")
		assert.Equal(t, len(result.Value().Concepts.Elements()), len(result2.Value().Concepts.Elements()),
			"element count should be preserved across round trips")
	})

	// Test case 3: populated slices should round trip correctly
	t.Run("populated_lists_roundtrip", func(t *testing.T) {
		metadata := cm.EntryMetadata{
			Concepts: []cm.TaxonomyConceptLink{
				cm.NewTaxonomyConceptLink("concept1"),
				cm.NewTaxonomyConceptLink("concept2"),
			},
			Tags: []cm.TagLink{
				cm.NewTagLink("tag1"),
			},
		}

		result, diags := NewEntryMetadataFromResponse(ctx, path.Root("metadata"), cm.NewOptEntryMetadata(metadata))
		require.False(t, diags.HasError(), "should not have errors")

		assert.False(t, result.Value().Concepts.IsNull(), "populated concepts should not be null")
		assert.Equal(t, 2, len(result.Value().Concepts.Elements()), "should have 2 concepts")

		assert.False(t, result.Value().Tags.IsNull(), "populated tags should not be null")
		assert.Equal(t, 1, len(result.Value().Tags.Elements()), "should have 1 tag")

		// Round trip back
		model := EntryModel{
			IDIdentityModel:    NewIDIdentityModelFromMultipartID("space1", "env1", "entry1"),
			EntryIdentityModel: EntryIdentityModel{SpaceID: types.StringValue("space1"), EnvironmentID: types.StringValue("env1"), EntryID: types.StringValue("entry1")},
			ContentTypeID:      types.StringValue("ct1"),
			Fields:             NewTypedMapNull[jsontypes.Normalized](),
			Metadata:           result,
		}

		req, diags := model.ToEntryRequest(ctx)
		require.False(t, diags.HasError(), "should not have errors")

		assert.NotNil(t, req.Metadata.Value.Concepts, "should have concepts")
		assert.Equal(t, 2, len(req.Metadata.Value.Concepts), "should preserve concept count")
		assert.Equal(t, "concept1", req.Metadata.Value.Concepts[0].Sys.ID)
		assert.Equal(t, "concept2", req.Metadata.Value.Concepts[1].Sys.ID)

		assert.NotNil(t, req.Metadata.Value.Tags, "should have tags")
		assert.Equal(t, 1, len(req.Metadata.Value.Tags), "should preserve tag count")
		assert.Equal(t, "tag1", req.Metadata.Value.Tags[0].Sys.ID)
	})
}

func rapidEntryModelGenerator() *rapid.Generator[EntryModel] {
	return rapid.Custom(func(t *rapid.T) EntryModel {
		spaceID := rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, "spaceID")
		environmentID := rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, "environmentID")
		contentTypeID := rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, "contentTypeID")
		entryID := rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, "entryID")

		model := EntryModel{
			IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID, entryID),
			EntryIdentityModel: EntryIdentityModel{
				SpaceID:       types.StringValue(spaceID),
				EnvironmentID: types.StringValue(environmentID),
				EntryID:       types.StringValue(entryID),
			},
			ContentTypeID: types.StringValue(contentTypeID),
		}

		hasFields := rapid.Bool().Draw(t, "hasFields")
		if hasFields {
			fieldsMap := make(map[string]jsontypes.Normalized)
			numFields := rapid.IntRange(0, 5).Draw(t, "numFields")
			for range numFields {
				key := rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, "fieldKey")
				jsonValue := `"test"`
				fieldsMap[key] = jsontypes.NewNormalizedValue(jsonValue)
			}
			model.Fields = NewTypedMap(fieldsMap)
		} else {
			model.Fields = NewTypedMapNull[jsontypes.Normalized]()
		}

		hasMetadata := rapid.Bool().Draw(t, "hasMetadata")
		if hasMetadata {
			metadata := EntryMetadataValue{}

			hasConcepts := rapid.Bool().Draw(t, "hasMetadataConcepts")
			if hasConcepts {
				numConcepts := rapid.IntRange(0, 3).Draw(t, "numMetadataConcepts")
				concepts := make([]types.String, numConcepts)
				for i := range concepts {
					concepts[i] = types.StringValue(rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, "metadataConcept"))
				}
				metadata.Concepts = NewTypedList(concepts)
			} else {
				metadata.Concepts = NewTypedListNull[types.String]()
			}

			hasTags := rapid.Bool().Draw(t, "hasMetadataTags")
			if hasTags {
				numTags := rapid.IntRange(0, 3).Draw(t, "numMetadataTags")
				tags := make([]types.String, numTags)
				for i := range tags {
					tags[i] = types.StringValue(rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, "metadataTag"))
				}
				metadata.Tags = NewTypedList(tags)
			} else {
				metadata.Tags = NewTypedListNull[types.String]()
			}

			model.Metadata = NewTypedObject(metadata)
		} else {
			model.Metadata = NewTypedObjectNull[EntryMetadataValue]()
		}

		return model
	})
}
