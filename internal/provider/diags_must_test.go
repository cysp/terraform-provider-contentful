package provider_test

import "github.com/hashicorp/terraform-plugin-framework/diag"

//nolint:ireturn
func DiagsNoErrorsMust[T any](value T, diags diag.Diagnostics) T {
	if diags.HasError() {
		panic(diags)
	}

	return value
}
