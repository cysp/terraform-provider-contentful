package provider_test

import (
	"context"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
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
		}

		return model
	})
}
