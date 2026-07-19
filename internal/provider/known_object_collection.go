package provider

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

type (
	knownObjectConverter[T any, R any]    func(context.Context, path.Path, T) (R, diag.Diagnostics)
	knownObjectMapConverter[T any, R any] func(context.Context, path.Path, string, T) (R, diag.Diagnostics)
)

func ConvertKnownObjectListElements[T any, R any](
	ctx context.Context,
	valuePath path.Path,
	elements []TypedObject[T],
	convert knownObjectConverter[T, R],
) ([]R, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	result := make([]R, 0, len(elements))

	for index, element := range elements {
		elementPath := valuePath.AtListIndex(index)
		value, valueDiags := KnownObjectValue(element, elementPath)
		diags.Append(valueDiags...)

		if valueDiags.HasError() {
			continue
		}

		converted, convertedDiags := convert(ctx, elementPath, value)
		diags.Append(convertedDiags...)

		if convertedDiags.HasError() {
			continue
		}

		result = append(result, converted)
	}

	if diags.HasError() {
		return nil, diags
	}

	return result, diags
}

func ConvertKnownObjectMapElements[T any, R any](
	ctx context.Context,
	valuePath path.Path,
	elements map[string]TypedObject[T],
	convert knownObjectMapConverter[T, R],
) ([]R, diag.Diagnostics) {
	keys := make([]string, 0, len(elements))
	for key := range elements {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	diags := diag.Diagnostics{}
	result := make([]R, 0, len(elements))

	for _, key := range keys {
		elementPath := valuePath.AtMapKey(key)
		value, valueDiags := KnownObjectValue(elements[key], elementPath)
		diags.Append(valueDiags...)

		if valueDiags.HasError() {
			continue
		}

		converted, convertedDiags := convert(ctx, elementPath, key, value)
		diags.Append(convertedDiags...)

		if convertedDiags.HasError() {
			continue
		}

		result = append(result, converted)
	}

	if diags.HasError() {
		return nil, diags
	}

	return result, diags
}
