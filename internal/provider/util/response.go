package util

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

//nolint:ireturn
func HandleUnexpectedResponse(_ context.Context, response interface{}, action string) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Unexpected response from Contentful when "+action,
		fmt.Sprintf("Unexpected response: %+v", response),
	)
}
