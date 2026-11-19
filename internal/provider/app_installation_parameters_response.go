package provider

import (
	"encoding/json"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Only a returned top-level string may represent a Secret mask. After a write
// the reference is the effective plan. On refresh it is the previous state.
func reconcileAppInstallationParameters(response, reference types.Map) types.Map {
	if response.IsNull() || response.IsUnknown() || reference.IsNull() || reference.IsUnknown() {
		return response
	}

	elements, prior := response.Elements(), reference.Elements()
	for key, element := range elements {
		returned := decodedParameterString(element)
		if returned == "" || strings.Trim(returned, "*") != "" {
			continue
		}

		referenceText := decodedParameterString(prior[key])
		if strings.Trim(referenceText, "*") != "" {
			elements[key] = prior[key]
		}
	}

	return types.MapValueMust(jsontypes.NormalizedType{}, elements)
}

func decodedParameterString(element attr.Value) string {
	value, ok := element.(jsontypes.Normalized)
	if !ok || value.IsUnknown() || value.IsNull() {
		return ""
	}

	var decoded string

	err := json.Unmarshal([]byte(value.ValueString()), &decoded)
	if err != nil {
		return ""
	}

	return decoded
}
