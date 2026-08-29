package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type pendingLifecycleFailure struct {
	summary string
	detail  string
}

type pendingLifecycleAuthorityOutcome uint8

const (
	pendingLifecycleAuthorityRevoked pendingLifecycleAuthorityOutcome = iota
	pendingLifecycleAuthorityRetained
	pendingLifecycleAuthorityConfirmed
)

func pendingLifecycleFailureFromDiagnostics(summary, context string, diagnostics diag.Diagnostics) *pendingLifecycleFailure {
	details := []string{context}

	for _, diagnostic := range diagnostics {
		if diagnostic.Severity() == diag.SeverityError {
			details = append(details, diagnostic.Summary()+": "+diagnostic.Detail())
		}
	}

	return &pendingLifecycleFailure{summary: summary, detail: strings.Join(details, "\n\n")}
}

func appendNonErrorDiagnostics(destination *diag.Diagnostics, source diag.Diagnostics) {
	for _, diagnostic := range source {
		if diagnostic.Severity() != diag.SeverityError {
			destination.Append(diagnostic)
		}
	}
}

func appendDiagnosticsWithErrorsAsWarnings(destination *diag.Diagnostics, source diag.Diagnostics, suffix string) {
	for _, diagnostic := range source {
		if diagnostic.Severity() == diag.SeverityError {
			if diagnosticWithPath, ok := diagnostic.(diag.DiagnosticWithPath); ok {
				destination.AddAttributeWarning(
					diagnosticWithPath.Path(), diagnostic.Summary(), diagnostic.Detail()+" "+suffix,
				)
			} else {
				destination.AddWarning(diagnostic.Summary(), diagnostic.Detail()+" "+suffix)
			}

			continue
		}

		destination.Append(diagnostic)
	}
}

func publishedVersionMatchesCheckpoint(observed cm.OptInt, checkpointed types.Int64) bool {
	if checkpointed.IsUnknown() {
		return false
	}

	value, present := observed.Get()
	if checkpointed.IsNull() {
		return !present
	}

	return present && checkpointed.ValueInt64() >= 0 && int64(value) == checkpointed.ValueInt64()
}

func pendingLifecycleDraftIsValid(version int, publishedVersion types.Int64) bool {
	if version <= 0 || publishedVersion.IsUnknown() {
		return false
	}

	return publishedVersion.IsNull() ||
		(publishedVersion.ValueInt64() >= 0 && publishedVersion.ValueInt64() < int64(version))
}

func pendingLifecycleDraftMatchesCheckpoint(
	markedVersion int,
	version int,
	publishedVersion cm.OptInt,
	checkpointedPublishedVersion types.Int64,
) bool {
	return version == markedVersion &&
		pendingLifecycleDraftIsValid(version, checkpointedPublishedVersion) &&
		publishedVersionMatchesCheckpoint(publishedVersion, checkpointedPublishedVersion)
}

func applyPendingLifecycleAuthorityOutcome(
	ctx context.Context,
	providerData PrivateProviderData,
	key string,
	outcome pendingLifecycleAuthorityOutcome,
) diag.Diagnostics {
	if outcome == pendingLifecycleAuthorityRetained {
		return nil
	}

	return providerData.SetKey(ctx, key, nil)
}

func reconcilePendingLifecycleAuthorityCheckpoint(
	ctx context.Context,
	providerData PrivateProviderData,
	key string,
	pendingVersion int,
) (bool, diag.Diagnostics) {
	checkpointVersion, present, diagnostics := optionalPrivateVersion(ctx, providerData)
	if diagnostics.HasError() || !present || checkpointVersion != pendingVersion {
		diagnostics.Append(providerData.SetKey(ctx, key, nil)...)

		return false, diagnostics
	}

	return true, diagnostics
}

func optionalPendingLifecycleVersion(
	ctx context.Context,
	providerData PrivateProviderData,
	key string,
	lifecycle string,
) (int, bool, diag.Diagnostics) {
	value, diagnostics := providerData.GetKey(ctx, key)
	if diagnostics.HasError() || len(value) == 0 {
		return 0, false, diagnostics
	}

	var version int

	err := json.Unmarshal(value, &version)
	if err != nil {
		diagnostics.AddError("Failed to decode pending "+lifecycle, err.Error())

		return 0, true, diagnostics
	}

	if version <= 0 {
		diagnostics.AddError(
			"Invalid pending "+lifecycle,
			fmt.Sprintf("Terraform private state contains nonpositive Contentful version %d.", version),
		)
	}

	return version, true, diagnostics
}

func contentfulResponseIsVersionMismatch(response any) bool {
	errorResponse, ok := response.(cm.ErrorStatusCodeResponse)
	if !ok {
		return false
	}

	responseError, ok := errorResponse.GetError()

	return ok && responseError.Sys.ID == cm.ErrorSysIDVersionMismatch
}
