package testdata

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"pgregory.net/rapid"
)

func JSONTypesNormalizedStringValue() *rapid.Generator[jsontypes.Normalized] {
	return rapid.Map(
		rapid.Map(
			rapid.String(),
			func(value string) string {
				bytes, _ := json.Marshal(value)
				return string(bytes)
			}),
		jsontypes.NewNormalizedValue,
	)
}
