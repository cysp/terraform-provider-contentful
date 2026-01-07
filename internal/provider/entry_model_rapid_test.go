package provider_test

import (
	"context"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/cysp/terraform-provider-contentful/internal/provider/testdata"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
)

func TestEntryModelRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rapid.Check(t, func(t *rapid.T) {
		model := rapidEntryModelGenerator.Draw(t, "model")

		entryRequest, diags := model.ToEntryRequest(ctx)
		if diags.HasError() {
			t.Fatalf("ToEntryRequest failed: %v", diags.Errors())
		}

		entry := cmt.NewEntryFromRequest(
			model.SpaceID.ValueString(),
			model.EnvironmentID.ValueString(),
			model.ContentTypeID.ValueString(),
			model.EntryID.ValueString(),
			&entryRequest,
		)

		result, diags := NewEntryResourceModelFromResponse(ctx, entry)
		if diags.HasError() {
			t.Fatalf("NewEntryResourceModelFromResponse failed: %v", diags.Errors())
		}

		assert.Equal(t, model, result)
	})
}

var rapidEntryModelGenerator = rapid.Custom(func(t *rapid.T) EntryModel {
	spaceID := testdata.AlphanumericStringOfN(1, 10).Draw(t, "spaceID")
	environmentID := testdata.AlphanumericStringOfN(1, 10).Draw(t, "environmentID")
	contentTypeID := testdata.AlphanumericStringOfN(1, 10).Draw(t, "contentTypeID")
	entryID := testdata.AlphanumericStringOfN(1, 10).Draw(t, "entryID")

	model := EntryModel{
		IDIdentityModel:    NewIDIdentityModelFromMultipartID(spaceID, environmentID, entryID),
		EntryIdentityModel: NewEntryIdentityModel(spaceID, environmentID, entryID),
		ContentTypeID:      types.StringValue(contentTypeID),
	}

	hasFields := rapid.Bool().Draw(t, "hasFields")
	if hasFields {
		fields := make(map[string]jsontypes.Normalized)

		fieldKeys := rapid.SliceOfN(testdata.AlphanumericStringOfN(1, 10), 0, 5).Draw(t, "fieldKeys")
		for _, key := range fieldKeys {
			fields[key] = jsontypes.NewNormalizedValue(`"value"`)
		}

		model.Fields = NewTypedMap(fields)
	}

	hasMetadata := rapid.Bool().Draw(t, "hasMetadata")
	if hasMetadata {
		metadata := EntryMetadataValue{}

		hasMetadataConcepts := rapid.Bool().Draw(t, "hasMetadataConcepts")
		if hasMetadataConcepts {
			concepts := rapid.SliceOfN(testdata.AlphanumericStringOfN(0, 10), 0, 3).Draw(t, "concepts")
			metadata.Concepts = NewTypedListFromStringSlice(concepts)
		}

		hasTags := rapid.Bool().Draw(t, "hasMetadataTags")
		if hasTags {
			tags := rapid.SliceOfN(testdata.AlphanumericStringOfN(0, 10), 0, 3).Draw(t, "tags")
			metadata.Tags = NewTypedListFromStringSlice(tags)
		}

		model.Metadata = NewTypedObject(metadata)
	}

	return model
})
