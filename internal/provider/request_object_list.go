package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

type knownObjectListElementConverter[T any, R any] func(context.Context, path.Path, T) (R, diag.Diagnostics)

func convertKnownObjectListElements[T any, R any](
	ctx context.Context,
	valuePath path.Path,
	elements []TypedObject[T],
	convert knownObjectListElementConverter[T, R],
) ([]R, diag.Diagnostics) {
	result := make([]R, 0, len(elements))
	diags := diag.Diagnostics{}

	for index, element := range elements {
		elementPath := valuePath.AtListIndex(index)
		value, valueDiags := RequireKnownObject(element, elementPath)
		diags.Append(valueDiags...)

		if valueDiags.HasError() {
			continue
		}

		converted, conversionDiags := convert(ctx, elementPath, value)
		diags.Append(conversionDiags...)

		if !conversionDiags.HasError() {
			result = append(result, converted)
		}
	}

	if diags.HasError() {
		return nil, diags
	}

	return result, diags
}
