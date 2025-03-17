package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// func DiagsAppendResult2[A any](diags diag.Diagnostics, fn func() (A, diag.Diagnostics)) A {
// 	value, valueDiags := fn()
// 	diags.Append(valueDiags...)

// 	return value
// }

// func DiagsAppendResult2WithContext[A any](ctx context.Context, diags diag.Diagnostics, fn func(context.Context) (A, diag.Diagnostics)) A {
// 	value, valueDiags := fn(ctx)
// 	diags.Append(valueDiags...)

// 	return value
// }

func DiagsAppendResult1[R any, A any](diags diag.Diagnostics, fn func(A) (R, diag.Diagnostics), a A) R {
	value, valueDiags := fn(a)
	diags.Append(valueDiags...)

	return value
}

func DiagsAppendResult2[R any, A any, B any](diags diag.Diagnostics, fn func(A, B) (R, diag.Diagnostics), a A, b B) R {
	value, valueDiags := fn(a, b)
	diags.Append(valueDiags...)

	return value
}

func DiagsAppendResult3[R any, A any, B any, C any](diags diag.Diagnostics, fn func(A, B, C) (R, diag.Diagnostics), a A, b B, c C) R {
	value, valueDiags := fn(a, b, c)
	diags.Append(valueDiags...)

	return value
}
