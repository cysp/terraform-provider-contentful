package provider

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

type KnownUnionAlternative struct {
	Name  string
	Path  path.Path
	Value attr.Value
}

func ExactlyOneKnownAlternative(unionPath path.Path, alternatives ...KnownUnionAlternative) (KnownUnionAlternative, diag.Diagnostics) {
	selected := KnownUnionAlternative{}
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
		return KnownUnionAlternative{}, diags
	}

	if knownCount != 1 {
		diags.AddAttributeError(
			unionPath,
			"Invalid union value",
			fmt.Sprintf("Exactly one of %s must be known and non-null.", strings.Join(names, ", ")),
		)

		return KnownUnionAlternative{}, diags
	}

	return selected, diags
}
