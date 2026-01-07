package testdata

import (
	"encoding/json"

	provider "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"pgregory.net/rapid"
)

func EntryModel(spaceID, environmentID, contentTypeID, entryID string) *rapid.Generator[provider.EntryModel] {
	return rapid.Custom(func(t *rapid.T) provider.EntryModel {
		return provider.EntryModel{
			IDIdentityModel:    provider.NewIDIdentityModelFromMultipartID(spaceID, environmentID, entryID),
			EntryIdentityModel: provider.NewEntryIdentityModel(spaceID, environmentID, entryID),
			ContentTypeID:      types.StringValue(contentTypeID),
			Fields: RandomZeroable(
				rapid.Map(
					rapid.MapOfN(
						AlphanumericStringOfN(1, 10),
						rapid.Map(
							rapid.Custom(func(t *rapid.T) string {
								value := rapid.String().Draw(t, "value")
								bytes, _ := json.Marshal(value)
								return string(bytes)
							}),
							jsontypes.NewNormalizedValue,
						),
						0,
						5,
					),
					provider.NewTypedMap,
				),
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
		}
	})
}
