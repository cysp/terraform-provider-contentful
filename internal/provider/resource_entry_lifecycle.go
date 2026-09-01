package provider

import (
	"context"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func setEntryIdentityStateAndVersion(
	ctx context.Context,
	identity *tfsdk.ResourceIdentity,
	state *tfsdk.State,
	private PrivateProviderData,
	model EntryModel,
	version int,
) diag.Diagnostics {
	diags := setResourceIdentityAndState(ctx, identity, state, entryIdentityAttributeNames(), &model)

	if diags.HasError() {
		return diags
	}

	diags.Append(SetPrivateProviderData(ctx, private, "version", version)...)

	return diags
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
