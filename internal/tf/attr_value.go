package tf

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AttrValue interface {
	IsNull() bool
	IsUnknown() bool
}

func IsNullOrUnknown(v AttrValue) bool {
	return v.IsNull() || v.IsUnknown()
}

func IsKnownAndPresent(v AttrValue) bool {
	return !v.IsNull() && !v.IsUnknown()
}

type KnownAndPresentStringValuable interface {
	AttrValue

	Elements() []attr.Value
	ElementsAs(ctx context.Context, dest interface{}, allowUnhandled bool) diag.Diagnostics
}

func KnownAndPresentStringValues(ctx context.Context, values KnownAndPresentStringValuable) ([]string, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if IsNullOrUnknown(values) {
		return []string{}, diags
	}

	elements := make([]types.String, len(values.Elements()))
	diags.Append(values.ElementsAs(ctx, &elements, false)...)

	strings := make([]string, 0, len(elements))

	for _, element := range elements {
		if IsKnownAndPresent(element) {
			strings = append(strings, element.ValueString())
		}
	}

	return strings, diags
}
