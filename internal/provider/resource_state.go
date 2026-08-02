package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func setResourceIdentityAndState(
	ctx context.Context,
	existingDiagnostics diag.Diagnostics,
	identity *tfsdk.ResourceIdentity,
	state *tfsdk.State,
	identityModel any,
	stateModel any,
) diag.Diagnostics {
	if existingDiagnostics.HasError() {
		return nil
	}

	if identity == nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Missing resource identity",
				"The provider attempted to publish state without a resource identity schema.",
			),
		}
	}

	stagedIdentity := *identity
	stagedState := *state

	diags := stagedIdentity.Set(ctx, identityModel)
	if diags.HasError() {
		return diags
	}

	diags.Append(stagedState.Set(ctx, stateModel)...)

	if diags.HasError() {
		return diags
	}

	*identity = stagedIdentity
	*state = stagedState

	return diags
}
