package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// setResourceMutationState publishes representable response values while retaining
// the endpoint actually mutated. endpoint contains only addressed path components;
// generated IDs and related-resource references remain response-owned.
// Publication diagnostics are separate so callers can checkpoint private version
// data before reporting an identity contradiction.
func setResourceMutationState(
	ctx context.Context,
	identity *tfsdk.ResourceIdentity,
	state *tfsdk.State,
	identityAttributeNames []string,
	stateModel any,
	resourceName string,
	endpoint map[string]string,
) (diag.Diagnostics, diag.Diagnostics) {
	stagedState := *state

	diags := stagedState.Set(ctx, stateModel)
	if diags.HasError() {
		return diags, nil
	}

	reconciler := mutationResponseReconciler{resourceName: resourceName}

	components := make([]string, 0, len(identityAttributeNames))
	for _, attributeName := range identityAttributeNames {
		attributePath := path.Root(attributeName)

		var value types.String
		diags.Append(stagedState.GetAttribute(ctx, attributePath, &value)...)

		if diags.HasError() {
			return diags, nil
		}

		if requested, owned := endpoint[attributeName]; owned {
			reconciler.pinIdentity(attributePath, types.StringValue(requested), value, &value)
			diags.Append(stagedState.SetAttribute(ctx, attributePath, value)...)

			if diags.HasError() {
				return diags, nil
			}
		}

		components = append(components, value.ValueString())
	}

	diags.Append(stagedState.SetAttribute(ctx, path.Root("id"), NewIDIdentityModelFromMultipartID(components...).ID)...)

	if diags.HasError() {
		return diags, nil
	}

	diags.Append(publishResourceIdentityAndState(ctx, identity, state, identityAttributeNames, stagedState)...)

	return diags, reconciler.diagnostics
}
