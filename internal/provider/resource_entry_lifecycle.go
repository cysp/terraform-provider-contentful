package provider

import (
	"context"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// checkpointEntryDraft replaces prior publication authority with the exact
// returned draft version only after its state is checkpointed and validated.
func checkpointEntryDraft(
	ctx context.Context,
	plan, response EntryModel,
	version int,
	fieldPolicy entryResponseFieldPolicy,
	identity *tfsdk.ResourceIdentity,
	state *tfsdk.State,
	private PrivateProviderData,
) (EntryModel, diag.Diagnostics) {
	diags := private.SetKey(ctx, entryPendingPublicationVersionPrivateKey, nil)
	response, consistencyDiags := projectEntryMutationResponse(ctx, plan, response, fieldPolicy)
	diags.Append(setResourceIdentityStateAndVersion(ctx, identity, state, private, entryIdentityAttributeNames(), &response, version)...)
	diags.Append(validateEntryDraftResponse(version, response.PublishedVersion)...)
	diags.Append(consistencyDiags...)

	if !diags.HasError() {
		diags.Append(SetPrivateProviderData(ctx, private, entryPendingPublicationVersionPrivateKey, version)...)
	}

	return response, diags
}

func validateEntryDraftResponse(version int, publishedVersion types.Int64) diag.Diagnostics {
	var diags diag.Diagnostics

	if !pendingLifecycleDraftIsValid(version, publishedVersion) {
		diags.AddError(
			"Unexpected entry draft response",
			fmt.Sprintf("Contentful returned draft version %d with an invalid publishedVersion.", version),
		)
	}

	return diags
}

func validateEntryPublicationResponse(sentVersion, responseVersion int, responsePublishedVersion cm.OptInt) diag.Diagnostics {
	var diags diag.Diagnostics

	publishedVersion, published := responsePublishedVersion.Get()
	switch {
	case !published:
		diags.AddError("Unexpected entry publication response", "Contentful accepted the publish request but omitted sys.publishedVersion.")
	case publishedVersion != sentVersion:
		diags.AddError("Unexpected entry publication response", fmt.Sprintf("Contentful accepted publication of version %d but returned publishedVersion %d.", sentVersion, publishedVersion))
	case responseVersion <= sentVersion:
		diags.AddError("Unexpected entry publication response", fmt.Sprintf("Contentful accepted publication of version %d but returned current version %d; the current version must be greater than the published version.", sentVersion, responseVersion))
	}

	return diags
}
