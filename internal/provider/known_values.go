package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//nolint:ireturn // The concrete result type is supplied by each TypedObject caller.
func KnownObjectValue[T any](value TypedObject[T], valuePath path.Path) (T, diag.Diagnostics) {
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

func KnownStringValue(value types.String, valuePath path.Path) (string, diag.Diagnostics) {
	if !value.IsNull() && !value.IsUnknown() {
		return value.ValueString(), nil
	}

	diags := diag.Diagnostics{}

	if value.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unexpected unknown string", "The string value must be known before it can be sent to Contentful.")
	} else {
		diags.AddAttributeError(valuePath, "Unexpected null string", "Null string values are not valid collection elements.")
	}

	return "", diags
}

func KnownBoolValue(value types.Bool, valuePath path.Path) (bool, diag.Diagnostics) {
	if !value.IsNull() && !value.IsUnknown() {
		return value.ValueBool(), nil
	}

	diags := diag.Diagnostics{}

	if value.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unexpected unknown boolean", "The boolean value must be known before it can be sent to Contentful.")
	} else {
		diags.AddAttributeError(valuePath, "Unexpected null boolean", "The boolean value cannot be null.")
	}

	return false, diags
}

func KnownStringListValues(
	ctx context.Context,
	value TypedList[types.String],
	valuePath path.Path,
	unknownSummary string,
	unknownDetail string,
	nullSummary string,
	nullDetail string,
) ([]string, diag.Diagnostics) {
	if value.IsUnknown() {
		return nil, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, unknownSummary, unknownDetail)}
	}

	if value.IsNull() {
		return nil, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, nullSummary, nullDetail)}
	}

	result := make([]string, len(value.Elements()))

	diags := tfsdk.ValueAs(ctx, value, &result)
	if diags.HasError() {
		return nil, diags
	}

	return result, diags
}
