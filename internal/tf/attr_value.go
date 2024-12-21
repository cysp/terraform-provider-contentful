package tf

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func ElementsAsStringSlice(ctx context.Context, value KnownAndPresentStringValuable) ([]string, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if IsNullOrUnknown(value) {
		return []string{}, diags
	}

	strings := make([]string, len(value.Elements()))
	diags.Append(value.ElementsAs(ctx, &strings, false)...)

	return strings, diags
}
