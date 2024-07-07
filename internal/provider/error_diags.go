package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func ErrorFromDiags(diags diag.Diagnostics) error {
	if diags.HasError() {
		return diagsError{diags: diags}
	}

	return nil
}

type diagsError struct {
	diags diag.Diagnostics
}

var _ error = diagsError{}

func (e diagsError) Error() string {
	return fmt.Sprintf("diags: %s", e.diags.Errors())
}
