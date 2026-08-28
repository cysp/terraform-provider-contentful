package provider

import (
	"context"
	"encoding/json"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const entryPendingPublishVersionPrivateKey = "pending-publish-version"

func setEntryIdentityStateAndVersion(
	ctx context.Context,
	identity *tfsdk.ResourceIdentity,
	state *tfsdk.State,
	private PrivateProviderData,
	model EntryModel,
	version int,
) diag.Diagnostics {
	var identityModel EntryIdentityModel

	diags := CopyAttributeValues(ctx, &identityModel, &model)
	if diags.HasError() {
		return diags
	}

	diags.Append(setResourceIdentityAndState(ctx, identity, state, &identityModel, &model)...)

	if diags.HasError() {
		return diags
	}

	diags.Append(SetPrivateProviderData(ctx, private, "version", version)...)

	return diags
}

func entryPendingPublishVersion(ctx context.Context, private PrivateProviderData) (int, bool, diag.Diagnostics) {
	value, diags := private.GetKey(ctx, entryPendingPublishVersionPrivateKey)
	if diags.HasError() || len(value) == 0 {
		return 0, false, diags
	}

	var (
		version int
		err     = json.Unmarshal(value, &version)
	)
	if err != nil {
		diags.AddError("Failed to read pending entry publication", err.Error())
	} else if version <= 0 {
		diags.AddError("Failed to read pending entry publication", "The stored pending publication version must be positive.")
	}

	return version, !diags.HasError(), diags
}

func setEntryPendingPublishVersion(ctx context.Context, private PrivateProviderData, version int) diag.Diagnostics {
	return SetPrivateProviderData(ctx, private, entryPendingPublishVersionPrivateKey, version)
}

func clearEntryPendingPublishVersion(ctx context.Context, private PrivateProviderData) diag.Diagnostics {
	return private.SetKey(ctx, entryPendingPublishVersionPrivateKey, nil)
}

func entryPublicationRecoveryAuthorized(
	version, pendingVersion int,
	pending bool,
	publishedVersion int64,
	published bool,
) bool {
	return pending && pendingVersion == version &&
		(!published || (publishedVersion > 0 && publishedVersion < int64(pendingVersion)))
}

func entryPendingPublicationShouldBeCleared(
	version, pendingVersion int,
	pending bool,
	publishedVersion int64,
	published bool,
) bool {
	return pending && !entryPublicationRecoveryAuthorized(version, pendingVersion, pending, publishedVersion, published)
}

func entryPublicationRecoveryRequired(version, pendingVersion int, pending bool, publishedVersion types.Int64) bool {
	if publishedVersion.IsUnknown() {
		return false
	}

	return entryPublicationRecoveryAuthorized(
		version,
		pendingVersion,
		pending,
		publishedVersion.ValueInt64(),
		!publishedVersion.IsNull(),
	)
}

func entryPublicationResponseFieldPolicy(
	normalPolicy entryResponseFieldPolicy,
	sentVersion int,
	responseVersion int,
	publishedVersion cm.OptInt,
) entryResponseFieldPolicy {
	if normalPolicy == entryResponseFieldsCreationDefaults &&
		entryPublicationResponseIsExact(sentVersion, responseVersion, publishedVersion) {
		return normalPolicy
	}

	return entryResponseFieldsExact
}

func entryPublicationResponseIsExact(sentVersion, responseVersion int, responsePublishedVersion cm.OptInt) bool {
	publishedVersion, published := responsePublishedVersion.Get()

	return published && publishedVersion == sentVersion && responseVersion == sentVersion+1
}

func validateEntryDraftResponse(version int, publishedVersion types.Int64) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case version <= 0:
		diags.AddError("Unexpected entry draft response", fmt.Sprintf("Contentful returned nonpositive draft version %d.", version))
	case publishedVersion.IsUnknown():
		diags.AddError("Unexpected entry draft response", "Contentful returned an unknown published version for a completed draft mutation.")
	case !publishedVersion.IsNull() && (publishedVersion.ValueInt64() <= 0 || publishedVersion.ValueInt64() >= int64(version)):
		diags.AddError("Unexpected entry draft response", fmt.Sprintf("Contentful returned draft version %d with publishedVersion %d.", version, publishedVersion.ValueInt64()))
	}

	return diags
}

func validateObservedEntryLifecycle(version int, publishedVersion cm.OptInt) diag.Diagnostics {
	var diags diag.Diagnostics

	published, hasPublished := publishedVersion.Get()

	switch {
	case version <= 0:
		diags.AddError("Unexpected entry lifecycle", fmt.Sprintf("Contentful returned nonpositive version %d.", version))
	case hasPublished && published <= 0:
		diags.AddError("Unexpected entry lifecycle", fmt.Sprintf("Contentful returned nonpositive publishedVersion %d.", published))
	case hasPublished && published >= version:
		diags.AddWarning("Unusual entry lifecycle", fmt.Sprintf("Contentful returned version %d with publishedVersion %d; Terraform preserved the representable remote state.", version, published))
	}

	return diags
}

func validateEntryStateLifecycle(version int, publishedVersion types.Int64) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case publishedVersion.IsUnknown():
		diags.AddError("Unexpected entry lifecycle", "The stored published version must be known.")
	case !publishedVersion.IsNull() && publishedVersion.ValueInt64() <= 0:
		diags.AddError("Unexpected entry lifecycle", fmt.Sprintf("The stored published version %d must be positive.", publishedVersion.ValueInt64()))
	case !publishedVersion.IsNull() && publishedVersion.ValueInt64() >= int64(version):
		diags.AddWarning("Unusual entry lifecycle", fmt.Sprintf("The stored entry version is %d with published version %d; Terraform preserved the representable state.", version, publishedVersion.ValueInt64()))
	}

	return diags
}

func validateEntryPublicationResponse(sentVersion, responseVersion int, responsePublishedVersion cm.OptInt) diag.Diagnostics {
	var diags diag.Diagnostics

	if entryPublicationResponseIsExact(sentVersion, responseVersion, responsePublishedVersion) {
		return diags
	}

	publishedVersion, published := responsePublishedVersion.Get()
	switch {
	case !published:
		diags.AddError("Unexpected entry publication response", "Contentful accepted the publish request but omitted sys.publishedVersion.")
	case publishedVersion != sentVersion:
		diags.AddError("Unexpected entry publication response", fmt.Sprintf("Contentful accepted publication of version %d but returned publishedVersion %d.", sentVersion, publishedVersion))
	case responseVersion <= 0:
		diags.AddError("Unexpected entry publication response", fmt.Sprintf("Contentful accepted publication of version %d but returned nonpositive version %d.", sentVersion, responseVersion))
	case responseVersion != sentVersion+1:
		diags.AddWarning("Unusual entry version after publication", fmt.Sprintf("Contentful reported publishedVersion %d but returned current version %d; Terraform checkpointed the representable response state.", sentVersion, responseVersion))
	}

	return diags
}
