package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const signingSecretWriteOnlyDescription = "Write-only signing secret, with the same format as `value`. Configure exactly one of `value` and `value_wo`. Requires Terraform 1.11.1 or later and accepts ephemeral values. The complete value is excluded from this resource's plan and state. Every update compares the supplied value, including timeout-only updates; a different value can rotate the secret even with `ignore_changes = [value_wo]`."

func resolveSigningSecretValue(value, valueWO types.String, modelPath path.Path) (types.String, path.Path, diag.Diagnostics) {
	valuePath, woPath := modelPath.AtName("value"), modelPath.AtName("value_wo")
	if !value.IsNull() && !valueWO.IsNull() {
		return types.StringNull(), woPath, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(woPath, "Conflicting signing secret values", "The effective plan retains value while value_wo is configured. Remove the conflicting ignore_changes rule before migrating to value_wo.")}
	}

	if !valueWO.IsNull() {
		return valueWO, woPath, nil
	}

	return value, valuePath, nil
}

// Signing-secret resources own only value_wo. The generic store also supports
// dynamic paths whose prior records may outlive the current configuration.
func readSigningSecretHashes(ctx context.Context, private PrivateProviderData) (writeOnlySecretHashes, diag.Diagnostics) {
	hashes, diags := readWriteOnlySecretHashes(ctx, private)
	if diags.HasError() {
		return hashes, diags
	}

	for _, record := range hashes.records {
		if record.Path != path.Root("value_wo").String() {
			return hashes, invalidWriteOnlySecretHashes("A verifier path is not supported by this signing-secret resource.")
		}
	}

	return hashes, diags
}

// A known ordinary value is the authoritative baseline. An obsolete verifier
// must never override a different known ordinary value.
func signingSecretValueMatches(ctx context.Context, hashes *writeOnlySecretHashes, priorValue, value types.String) (bool, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() || priorValue.IsUnknown() {
		return false, diag.Diagnostics{diag.NewErrorDiagnostic("Unresolved signing secret comparison", "Secret comparison requires a known non-null candidate and a known or null prior value.")}
	}

	if !priorValue.IsNull() {
		return priorValue.Equal(value), nil
	}

	return hashes.matches(ctx, path.Root("value_wo"), value)
}

// Prepare hash changes before PUT so hashing cannot fail after changing the
// remote secret. The caller saves these changes after a successful PUT,
// or when no PUT is needed.
func prepareSigningSecretValue(ctx context.Context, hashes *writeOnlySecretHashes, priorValue, value types.String, valuePath path.Path) (bool, diag.Diagnostics) {
	matches, diags := signingSecretValueMatches(ctx, hashes, priorValue, value)
	if diags.HasError() {
		return false, diags
	}

	if valuePath.Equal(path.Root("value_wo")) {
		if !matches || !priorValue.IsNull() {
			diags.Append(hashes.set(ctx, valuePath, value)...)
		}
	} else {
		diags.Append(hashes.remove(path.Root("value_wo"))...)
	}

	return matches, diags
}

// Encode public state and identity before publishing the prepared private
// value. Diagnostics do not roll back Framework response mutations.
func publishSigningSecretState(ctx context.Context, private PrivateProviderData, hashes *writeOnlySecretHashes, identity *tfsdk.ResourceIdentity, state *tfsdk.State, identityNames []string, model any) diag.Diagnostics {
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

	diags.Append(hashes.publish(ctx, private)...)

	if diags.HasError() {
		return diags
	}

	if identity != nil {
		*identity = *stagedIdentity
	}

	*state = stagedState

	return diags
}

func maskSigningSecretValues(ctx context.Context, values ...types.String) context.Context {
	knownValues := make([]string, 0, len(values))

	for _, value := range values {
		if !value.IsNull() && !value.IsUnknown() && value.ValueString() != "" {
			knownValues = append(knownValues, value.ValueString())
		}
	}

	return tflog.MaskLogStrings(ctx, knownValues...)
}
