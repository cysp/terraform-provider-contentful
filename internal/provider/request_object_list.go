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
		value, ok := element.GetValue()

		if !ok {
			if element.IsUnknown() {
				diags.AddAttributeError(
					elementPath,
					"Unexpected unknown object",
					"The object value must be known before it can be sent to Contentful.",
				)
			} else {
				diags.AddAttributeError(
					elementPath,
					"Unexpected null object",
					"Null object values are not valid collection elements.",
				)
			}

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
