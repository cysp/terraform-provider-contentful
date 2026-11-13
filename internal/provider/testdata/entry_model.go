package testdata

import (
	provider "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"pgregory.net/rapid"
)

// EntryModel exercises representable field JSON, including values that would
// need content-type and locale context to validate as a CMA Entry.
func EntryModel(spaceID, environmentID, contentTypeID, entryID string) *rapid.Generator[provider.EntryModel] {
	return rapid.Custom(func(t *rapid.T) provider.EntryModel {
		return provider.EntryModel{
			IDIdentityModel:    provider.NewIDIdentityModelFromMultipartID(spaceID, environmentID, entryID),
			EntryIdentityModel: provider.NewEntryIdentityModel(spaceID, environmentID, entryID),
			ContentTypeID:      types.StringValue(contentTypeID),
			Fields: rapid.Map(
				rapid.MapOfN(
					AlphanumericStringOfN(1, 10),
					NormalizedJSONValue(),
					0,
					5,
				),
				provider.NewTypedMap,
			).Draw(t, "fields"),
			Metadata: RandomZeroable(
				rapid.Map(
					rapid.Custom(func(t *rapid.T) provider.EntryMetadataValue {
						return provider.EntryMetadataValue{
							Concepts: RandomZeroable(
								rapid.Map(
									rapid.SliceOfN(AlphanumericStringOfN(0, 10), 0, 3),
									provider.NewTypedListFromStringSlice,
								),
							).Draw(t, "concepts"),
							Tags: RandomZeroable(
								rapid.Map(
									rapid.SliceOfN(AlphanumericStringOfN(0, 10), 0, 3),
									provider.NewTypedListFromStringSlice,
								),
							).Draw(t, "tags"),
						}
					}),
					provider.NewTypedObject,
				),
			).Draw(t, "metadata"),
			Timeouts: provider.TimeoutsNull(),
		}
	})
}
