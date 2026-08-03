package provider

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

type knownUnionAlternative[T any] struct {
	Name    string
	Path    path.Path
	Value   attr.Value
	Convert func() (T, diag.Diagnostics)
}

//nolint:ireturn // The selected converter supplies the concrete result type.
func convertExactlyOneKnownAlternative[T any](
	unionPath path.Path,
	alternatives ...knownUnionAlternative[T],
) (T, diag.Diagnostics) {
	var zero T

	selectedIndex := -1
	knownCount := 0
	names := make([]string, 0, len(alternatives))
	diags := diag.Diagnostics{}

	for index, alternative := range alternatives {
		names = append(names, alternative.Name)

		switch {
		case alternative.Value.IsUnknown():
			diags.AddAttributeError(
				alternative.Path,
				"Unexpected unknown union alternative",
				fmt.Sprintf("%q must be known before the union can be sent to Contentful.", alternative.Name),
			)
		case alternative.Value.IsNull():
		default:
			knownCount++
			selectedIndex = index
		}
	}

	if knownCount > 1 || (knownCount == 0 && !diags.HasError()) {
		diags.AddAttributeError(
			unionPath,
			"Invalid union value",
			fmt.Sprintf("Exactly one of %s must be known and non-null.", strings.Join(names, ", ")),
		)
	}

	if diags.HasError() {
		return zero, diags
	}

	result, conversionDiags := alternatives[selectedIndex].Convert()
	diags.Append(conversionDiags...)

	if diags.HasError() {
		return zero, diags
	}

	return result, diags
}
