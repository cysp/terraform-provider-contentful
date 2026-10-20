package provider

import (
	"encoding/json"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// reconcileAppInstallationParameters preserves intended top-level string values
// when Contentful returns a possible Secret mask. This does not verify the
// underlying plaintext. Callers supply the effective Plan after a successful
// write, or prior State during Read.
func reconcileAppInstallationParameters(response, reference jsontypes.Normalized) (jsontypes.Normalized, diag.Diagnostics) {
	var diags diag.Diagnostics

	if response.IsNull() || response.IsUnknown() || reference.IsNull() || reference.IsUnknown() {
		return response, diags
	}

	// Raw members retain unrelated numeric tokens without converting to float64.
	var responseObject, referenceObject map[string]json.RawMessage

	for _, value := range []struct {
		input  jsontypes.Normalized
		object *map[string]json.RawMessage
	}{
		{response, &responseObject},
		{reference, &referenceObject},
	} {
		raw := []byte(value.input.ValueString())
		if !json.Valid(raw) {
			diags.AddAttributeError(path.Root("parameters"), "Failed to reconcile app installation parameters", "App installation parameters could not be decoded as JSON.")

			return response, diags
		}

		if !strings.HasPrefix(strings.TrimSpace(string(raw)), "{") {
			return response, diags
		}

		err := json.Unmarshal(raw, value.object)
		if err != nil {
			diags.AddAttributeError(path.Root("parameters"), "Failed to reconcile app installation parameters", "App installation parameters could not be decoded as a JSON object.")

			return response, diags
		}
	}

	changed := false

	for key, raw := range responseObject {
		var returned, intended string

		if json.Unmarshal(raw, &returned) != nil || returned == "" || strings.Trim(returned, "*") != "" {
			continue
		}

		if json.Unmarshal(referenceObject[key], &intended) != nil || intended == "" || strings.Trim(intended, "*") == "" {
			continue
		}

		responseObject[key] = referenceObject[key]
		changed = true
	}

	if !changed {
		return response, diags
	}

	reconciled, err := json.Marshal(responseObject)
	if err != nil {
		diags.AddAttributeError(path.Root("parameters"), "Failed to reconcile app installation parameters", "App installation parameters could not be encoded as JSON.")

		return response, diags
	}

	return NewNormalizedJSONValue(reconciled), diags
}
