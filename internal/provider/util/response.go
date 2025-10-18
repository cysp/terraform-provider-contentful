package util

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func HandleUnexpectedResponse(ctx context.Context, response interface{}, action string) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Unexpected response from Contentful when %s", action),
		fmt.Sprintf("Unexpected response: %+v", response),
	)
}
