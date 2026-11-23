package provider

import (
	"context"
	"maps"
	"slices"

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
	stagedState := *state

	diags := stagedState.Set(ctx, stateModel)
	if diags.HasError() {
		return diags
	}

	diags.Append(publishResourceIdentityAndState(ctx, identity, state, identityAttributeNames, stagedState)...)

	return diags
}

// publishResourceIdentityAndState commits an already encoded state and its
// derived identity together. Neither target changes if identity encoding fails.
func publishResourceIdentityAndState(
	ctx context.Context,
	identity *tfsdk.ResourceIdentity,
	state *tfsdk.State,
	identityAttributeNames []string,
	stagedState tfsdk.State,
) diag.Diagnostics {
	diags := setResourceIdentityFromState(ctx, identity, stagedState, identityAttributeNames)
	if !diags.HasError() {
		*state = stagedState
	}

	return diags
}

func setResourceIdentityFromState(
	ctx context.Context,
	identity *tfsdk.ResourceIdentity,
	state tfsdk.State,
	identityAttributeNames []string,
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

	var diags diag.Diagnostics

	for _, attributeName := range identityAttributeNames {
		attributePath := path.Root(attributeName)

		var value types.String

		diags.Append(state.GetAttribute(ctx, attributePath, &value)...)

		if diags.HasError() {
			return diags
		}

		diags.Append(stagedIdentity.SetAttribute(ctx, attributePath, value)...)

		if diags.HasError() {
			return diags
		}
	}

	*identity = stagedIdentity

	return diags
}

// Framework requires an identity even when Read returns null state. Preserve a
// supplied identity, or reconstruct it from state before removing the resource.
func removeResourceFromStateWithIdentity(
	ctx context.Context,
	state *tfsdk.State,
	identity *tfsdk.ResourceIdentity,
	diags *diag.Diagnostics,
) {
	if identity != nil && identity.Raw.IsFullyNull() {
		names := slices.Sorted(maps.Keys(identity.Schema.GetAttributes()))
		diags.Append(setResourceIdentityFromState(ctx, identity, *state, names)...)
	}

	if !diags.HasError() {
		state.RemoveResource(ctx)
	}
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

// setResourceIdentityStateAndMutationVersion publishes representable recovery
// state, then records a version only when the response identified the endpoint.
// A contradiction clears the previous lock token: neither a response from another
// object nor the stale pre-mutation version may authorize the next write.
func setResourceIdentityStateAndMutationVersion(
	ctx context.Context,
	identity *tfsdk.ResourceIdentity,
	state *tfsdk.State,
	private PrivateProviderData,
	identityAttributeNames []string,
	stateModel any,
	version int,
	identityConfirmed bool,
) diag.Diagnostics {
	diags := setResourceIdentityAndState(ctx, identity, state, identityAttributeNames, stateModel)
	if diags.HasError() {
		return diags
	}

	diags.Append(setResourceMutationVersion(ctx, private, version, identityConfirmed)...)

	return diags
}

func setResourceMutationVersion(ctx context.Context, private PrivateProviderData, version int, identityConfirmed bool) diag.Diagnostics {
	if !identityConfirmed {
		return private.SetKey(ctx, "version", nil)
	}

	return SetPrivateProviderData(ctx, private, "version", version)
}

// Verifiers, public credential values and identity must move together. Encode
// the proposed state and identity before changing private data. Framework
// diagnostics alone do not roll back assignments to the response.
func publishStateAfterPrivate(ctx context.Context, identity *tfsdk.ResourceIdentity, state *tfsdk.State, identityNames []string, model any, publishPrivate func() diag.Diagnostics) diag.Diagnostics {
	stagedState := *state

	var stagedIdentity *tfsdk.ResourceIdentity

	if identity != nil {
		copied := *identity
		stagedIdentity = &copied
	}

	diags := setResourceIdentityAndState(ctx, stagedIdentity, &stagedState, identityNames, model)
	if diags.HasError() {
		return diags
	}

	diags.Append(publishPrivate()...)

	if diags.HasError() {
		return diags
	}

	if identity != nil {
		*identity = *stagedIdentity
	}

	*state = stagedState

	return diags
}
