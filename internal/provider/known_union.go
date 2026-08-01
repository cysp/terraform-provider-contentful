package provider

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

type KnownUnionAlternative[T any] struct {
	Name    string
	Path    path.Path
	Value   attr.Value
	Convert func() (T, diag.Diagnostics)
}

func ConvertExactlyOneKnownAlternative[T any](unionPath path.Path, alternatives ...KnownUnionAlternative[T]) (T, diag.Diagnostics) {
	var zero T

	selected := KnownUnionAlternative[T]{}
	knownCount := 0
	names := make([]string, 0, len(alternatives))
	diags := diag.Diagnostics{}

	for _, alternative := range alternatives {
		names = append(names, alternative.Name)

		if alternative.Value.IsUnknown() {
			diags.AddAttributeError(
				alternative.Path,
				"Unexpected unknown union alternative",
				fmt.Sprintf("%q must be known before the union can be sent to Contentful.", alternative.Name),
			)

			continue
		}

		if alternative.Value.IsNull() {
			continue
		}

		knownCount++
		selected = alternative
	}

	if diags.HasError() {
		return zero, diags
	}

	if knownCount != 1 {
		diags.AddAttributeError(
			unionPath,
			"Invalid union value",
			fmt.Sprintf("Exactly one of %s must be known and non-null.", strings.Join(names, ", ")),
		)

		return zero, diags
	}

	result, conversionDiags := selected.Convert()
	diags.Append(conversionDiags...)

	if diags.HasError() {
		return zero, diags
	}

	return result, diags
}
