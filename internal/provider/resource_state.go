package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// setResourceIdentityAndState encodes state into a staged copy, derives the
// declared string identity attributes from that state, and assigns neither
// target if any operation reports an error. This publishes local Terraform
// values only; it does not make remote operations transactional or recoverable.
func setResourceIdentityAndState(
	ctx context.Context,
	identity *tfsdk.ResourceIdentity,
	state *tfsdk.State,
	identityAttributeNames []string,
	stateModel any,
) diag.Diagnostics {
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

	diags := stagedState.Set(ctx, stateModel)
	if diags.HasError() {
		return diags
	}

	for _, attributeName := range identityAttributeNames {
		attributePath := path.Root(attributeName)

		var value types.String

		diags.Append(stagedState.GetAttribute(ctx, attributePath, &value)...)

		if diags.HasError() {
			return diags
		}

		diags.Append(stagedIdentity.SetAttribute(ctx, attributePath, value)...)

		if diags.HasError() {
			return diags
		}
	}

	*identity = stagedIdentity
	*state = stagedState

	return diags
}

// setResourceIdentityStateAndVersion publishes the state/identity pair before
// recording its Contentful version. Encoding errors leave all three unchanged;
// a private-data error leaves the published state/identity pair intact. Callers
// decide when response-derived recovery state is publishable and report any
// consistency diagnostics after this checkpoint.
func setResourceIdentityStateAndVersion(
	ctx context.Context,
	identity *tfsdk.ResourceIdentity,
	state *tfsdk.State,
	private PrivateProviderData,
	identityAttributeNames []string,
	stateModel any,
	version int,
) diag.Diagnostics {
	diags := setResourceIdentityAndState(ctx, identity, state, identityAttributeNames, stateModel)
	if diags.HasError() {
		return diags
	}

	diags.Append(SetPrivateProviderData(ctx, private, "version", version)...)

	return diags
}
