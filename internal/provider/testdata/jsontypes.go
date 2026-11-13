package testdata

import (
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"pgregory.net/rapid"
)

// NormalizedJSONValue generates bounded representable JSON, independently of
// provider normalization. It does not impose a resource's schema or CMA rules.
func NormalizedJSONValue() *rapid.Generator[jsontypes.Normalized] {
	return rapid.Map(jsonValue(2), func(value any) jsontypes.Normalized {
		encoded, _ := json.Marshal(value)
		return jsontypes.NewNormalizedValue(string(encoded))
	})
}

func jsonValue(depth int) *rapid.Generator[any] {
	scalar := rapid.OneOf(
		rapid.Just[any](nil),
		rapid.Map(rapid.Bool(), func(value bool) any { return value }),
		rapid.Map(rapid.StringN(0, 20, 80), func(value string) any { return value }),
		rapid.Map(rapid.Int64(), func(value int64) any { return json.Number(strconv.FormatInt(value, 10)) }),
		rapid.SampledFrom([]any{
			json.Number("9007199254740993"), json.Number("-9007199254740993"),
			json.Number("1.234567890123456789"), json.Number("-0"), json.Number("1e+30"),
		}),
	)
	if depth == 0 {
		return scalar
	}

	child := jsonValue(depth - 1)
	return rapid.OneOf(
		scalar,
		rapid.Map(rapid.SliceOfN(child, 0, 3), func(value []any) any { return value }),
		rapid.Map(rapid.MapOfN(rapid.StringN(0, 8, 24), child, 0, 3), func(value map[string]any) any { return value }),
	)
}

func NormalizedJSONStringValue() *rapid.Generator[jsontypes.Normalized] {
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
