package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

func NormalizeJSON(rawJSON []byte) string {
	var v any

	err := json.Unmarshal(rawJSON, &v)
	if err != nil {
		return string(rawJSON)
	}

	normalized, err := json.Marshal(v)
	if err != nil {
		return string(rawJSON)
	}

	return string(normalized)
}

func NewNormalizedJSONTypesNormalizedValue(bytes []byte) jsontypes.Normalized {
	return jsontypes.NewNormalizedValue(NormalizeJSON(bytes))
}
