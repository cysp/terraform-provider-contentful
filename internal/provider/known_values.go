package provider

import (
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//nolint:ireturn // The concrete result type is supplied by each TypedObject caller.
func RequireKnownObject[T any](value TypedObject[T], valuePath path.Path) (T, diag.Diagnostics) {
	if object, ok := value.GetValue(); ok {
		return object, nil
	}

	var zero T

	diags := diag.Diagnostics{}

	if value.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unexpected unknown object", "The object value must be known before it can be sent to Contentful.")
	} else {
		diags.AddAttributeError(valuePath, "Unexpected null object", "Null object values are not valid collection elements.")
	}

	return zero, diags
}

func RequireKnownStringList(value types.List, valuePath path.Path) ([]string, diag.Diagnostics) {
	if value.IsUnknown() {
		return nil, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, "Unexpected unknown list", "The list must be known before it can be sent to Contentful.")}
	}

	if value.IsNull() {
		return nil, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, "Unexpected null list", "The list cannot be null.")}
	}

	elements := make([]types.String, 0, len(value.Elements()))
	diags := diag.Diagnostics{}

	for index, element := range value.Elements() {
		stringElement, ok := element.(types.String)
		if !ok {
			diags.AddAttributeError(
				valuePath.AtListIndex(index),
				"Unexpected list element type",
				fmt.Sprintf("Expected a string value, got %T. This is a provider implementation error.", element),
			)

			continue
		}

		elements = append(elements, stringElement)
	}

	if diags.HasError() {
		return nil, diags
	}

	return knownStringListElements(valuePath, elements)
}

func RequireKnownStringMap(value types.Map, valuePath path.Path) (map[string]string, diag.Diagnostics) {
	return convertRequiredKnownMap(value, valuePath, func(element attr.Value, elementPath path.Path) (string, diag.Diagnostics) {
		stringElement, ok := element.(types.String)
		if !ok {
			return "", diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
				elementPath,
				"Unexpected map element type",
				fmt.Sprintf("Expected a string value, got %T. This is a provider implementation error.", element),
			)}
		}

		return requestRequiredString(stringElement, elementPath)
	})
}

func RequireKnownStringListMap(value types.Map, valuePath path.Path) (map[string][]string, diag.Diagnostics) {
	return convertRequiredKnownMap(value, valuePath, func(element attr.Value, elementPath path.Path) ([]string, diag.Diagnostics) {
		listElement, ok := element.(types.List)
		if !ok {
			return nil, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
				elementPath,
				"Unexpected map element type",
				fmt.Sprintf("Expected a list value, got %T. This is a provider implementation error.", element),
			)}
		}

		return RequireKnownStringList(listElement, elementPath)
	})
}

func convertRequiredKnownMap[T any](
	value types.Map,
	valuePath path.Path,
	convert func(attr.Value, path.Path) (T, diag.Diagnostics),
) (map[string]T, diag.Diagnostics) {
	if value.IsUnknown() {
		return nil, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, "Unexpected unknown map", "The map must be known before it can be sent to Contentful.")}
	}

	if value.IsNull() {
		return nil, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, "Unexpected null map", "The map cannot be null.")}
	}

	keys := make([]string, 0, len(value.Elements()))
	for key := range value.Elements() {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	result := make(map[string]T, len(keys))
	diags := diag.Diagnostics{}

	for _, key := range keys {
		element := value.Elements()[key]
		converted, valueDiags := convert(element, valuePath.AtMapKey(key))
		diags.Append(valueDiags...)

		if !valueDiags.HasError() {
			result[key] = converted
		}
	}

	if diags.HasError() {
		return nil, diags
	}

	return result, diags
}
