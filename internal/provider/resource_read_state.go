package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// validateResourceReadIdentity checks the response against the state attributes
// used to address the GET. Unlike ResourceIdentity, these attributes are present
// during import and when upgrading legacy state. No response or private state may
// be published before this check succeeds.
func validateResourceReadIdentity(ctx context.Context, prior tfsdk.State, identityAttributeNames []string, model any) diag.Diagnostics {
	projected := prior

	diags := projected.Set(ctx, model)
	if diags.HasError() {
		return diags
	}

	for _, name := range identityAttributeNames {
		attributePath := path.Root(name)

		var expected, actual types.String

		valueDiags := prior.GetAttribute(ctx, attributePath, &expected)
		valueDiags.Append(projected.GetAttribute(ctx, attributePath, &actual)...)
		diags.Append(valueDiags...)

		if valueDiags.HasError() {
			return diags
		}

		if !actual.Equal(expected) {
			diags.AddAttributeError(attributePath, "Unexpected Contentful read identity", "The response identity differs from the requested Contentful endpoint. Prior state was retained.")
		}
	}

	return diags
}
