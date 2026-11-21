package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/text/unicode/norm"
)

// Installation parameters can contain credentials. Report the key without quoting the JSON input.
type installationParameterValuesValidator struct{}

func (installationParameterValuesValidator) Description(context.Context) string {
	return "Each parameter value must be a non-null string containing valid JSON."
}

func (v installationParameterValuesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (installationParameterValuesValidator) ValidateMap(_ context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	for key, element := range req.ConfigValue.Elements() {
		if !element.IsUnknown() {
			_, diags := installationParameterJSON(req.Path.AtMapKey(key), element)
			resp.Diagnostics.Append(diags...)
		}
	}
}

func installationParameterJSON(valuePath path.Path, element attr.Value) (json.RawMessage, diag.Diagnostics) {
	var diags diag.Diagnostics

	value, ok := element.(jsontypes.Normalized)
	switch {
	case !ok:
		diags.AddAttributeError(valuePath, "Invalid installation parameter type", "Expected a JSON string for this parameter.")
	case value.IsUnknown():
		diags.AddAttributeError(valuePath, "Unknown installation parameter", "The parameter must be a known JSON string before it can be sent to Contentful.")
	case value.IsNull():
		diags.AddAttributeError(valuePath, "Null installation parameter", "Omit the key to remove the parameter, or use jsonencode(null) to send a JSON null.")
	default:
		raw := json.RawMessage(value.ValueString())
		if json.Valid(raw) {
			return raw, nil
		}

		diags.AddAttributeError(valuePath, "Invalid installation parameter JSON", "Encode the parameter value with jsonencode.")
	}

	return nil, diags
}

func encodeInstallationParameters(valuePath path.Path, value types.Map) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unexpected unknown app installation parameters", "App installation parameters must be known before they can be sent to Contentful.")

		return nil, diags
	}

	if value.IsNull() {
		return nil, nil
	}

	elements := value.Elements()

	members := make(map[string]json.RawMessage, len(elements))
	for key, element := range elements {
		raw, memberDiags := installationParameterJSON(valuePath.AtMapKey(key), element)
		diags.Append(memberDiags...)

		members[key] = raw
	}

	if diags.HasError() {
		return nil, diags
	}

	raw, err := json.Marshal(members)
	if err != nil {
		diags.AddAttributeError(valuePath, "Failed to encode installation parameters", "The parameters could not be encoded as a JSON object.")
	}

	return raw, diags
}

// Responses and legacy state use the same JSON representation.
func decodeInstallationParameters(valuePath path.Path, raw []byte) (types.Map, diag.Diagnostics) {
	elementType := jsontypes.NormalizedType{}

	null := types.MapNull(elementType)
	if raw == nil {
		return null, nil
	}

	var diags diag.Diagnostics

	var members map[string]json.RawMessage

	err := json.Unmarshal(raw, &members)
	if err != nil {
		diags.AddAttributeError(valuePath, "Invalid installation parameters", "Expected null or a JSON object keyed by parameter ID.")

		return null, diags
	}

	if members == nil {
		return null, nil
	}

	elements := make(map[string]attr.Value, len(members))
	for parameterID, member := range members {
		if !norm.NFC.IsNormalString(parameterID) {
			diags.AddAttributeError(valuePath, "Unrepresentable installation parameter ID", "Terraform would change a parameter ID during Unicode normalization. The Terraform parameters map requires NFC-normalized keys.")

			return null, diags
		}

		normalized, err := parseNormalizedJSONValue(member)
		if err != nil {
			diags.AddAttributeError(valuePath.AtMapKey(parameterID), "Invalid installation parameter", "The parameter value could not be normalized as JSON.")

			return null, diags
		}

		elements[parameterID] = jsontypes.NewNormalizedValue(escapeNonNFCJSON(normalized.ValueString()))
	}

	return types.MapValue(elementType, elements)
}

// Terraform normalizes strings to NFC. Keep literal non-NFC content inside a
// JSON member lossless by representing its non-ASCII runes with JSON escapes.
func escapeNonNFCJSON(text string) string {
	if norm.NFC.IsNormalString(text) {
		return text
	}

	var out strings.Builder

	for _, r := range text {
		if r < utf8.RuneSelf {
			out.WriteRune(r)

			continue
		}

		for _, unit := range utf16.AppendRune(nil, r) {
			fmt.Fprintf(&out, "\\u%04x", unit)
		}
	}

	return out.String()
}

func installationParametersEquivalent(ctx context.Context, plan, state types.Map) (bool, diag.Diagnostics) {
	if plan.Equal(state) {
		return true, nil
	}

	if plan.IsNull() || state.IsNull() || plan.IsUnknown() || state.IsUnknown() {
		return false, nil
	}

	planned, prior := plan.Elements(), state.Elements()
	if len(planned) != len(prior) {
		return false, nil
	}

	for parameterID, element := range planned {
		plannedValue, plannedOK := element.(jsontypes.Normalized)

		priorValue, priorOK := prior[parameterID].(jsontypes.Normalized)
		if !plannedOK || !priorOK {
			return false, nil
		}

		equal, diags := normalizedJSONEquivalent(ctx, plannedValue, priorValue)
		if !equal || diags.HasError() {
			return false, diags
		}
	}

	return true, nil
}
