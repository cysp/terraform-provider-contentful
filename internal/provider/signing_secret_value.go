package provider

import (
	"context"
	"encoding/json/v2"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const signingSecretWriteOnlyDescription = "Write-only signing secret, with the same format as `value`. Configure exactly one of `value` and `value_wo`. Requires Terraform 1.11.1 or later and accepts ephemeral values. Terraform omits the write-only input from planned and applied state. A saved migration plan includes prior state and can contain an earlier ordinary `value`. Every update compares the supplied value, including timeout-only updates; a different value can rotate the secret even with `ignore_changes = [value_wo]`."

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

type signingSecretVerifier struct {
	hash       secretVerifier
	superseded bool
}

func readSigningSecretVerifier(ctx context.Context, private PrivateProviderData, priorValue types.String) (signingSecretVerifier, diag.Diagnostics) {
	data, diags := readVerifierData(ctx, private)
	if diags.HasError() || len(data) == 0 {
		return signingSecretVerifier{}, diags
	}

	// The known ordinary value supersedes even unreadable verifier contents.
	if !priorValue.IsNull() && !priorValue.IsUnknown() {
		return signingSecretVerifier{superseded: true}, diags
	}

	diags.Append(validateVerifierData(data)...)

	if diags.HasError() {
		return signingSecretVerifier{}, diags
	}

	var hash secretVerifier

	err := json.Unmarshal(data, &hash)
	if err != nil || hash == "" {
		return signingSecretVerifier{}, invalidWriteOnlySecretHashes("Expected a non-empty JSON string containing the signing secret verifier.")
	}

	if _, _, reason := parseWriteOnlySecretHash(string(hash)); reason != "" {
		return signingSecretVerifier{}, invalidWriteOnlySecretHashes(reason)
	}

	return signingSecretVerifier{hash: hash}, diags
}

func (v signingSecretVerifier) changeTo(next secretVerifier) (writeOnlySecretVerifierChange, diag.Diagnostics) {
	if v.hash == next && !v.superseded {
		return writeOnlySecretVerifierChange{}, nil
	}

	if next == "" {
		return writeOnlySecretVerifierChange{changed: true}, nil
	}

	return encodeVerifierChange(next)
}

// Compare the selected bytes and finish encoding their representation before PUT.
// Create supplies an empty verifier and always writes, regardless of the match.
func prepareSigningSecretValue(ctx context.Context, verifier signingSecretVerifier, prior types.String, requestValue string, valueWO types.String) (bool, writeOnlySecretVerifierChange, diag.Diagnostics) {
	credential := rootWriteOnlySecretPath(signingSecretWO)

	matches, diags := matchesPriorValue(ctx, verifier.hash, credential, prior, types.StringValue(requestValue))
	if diags.HasError() {
		return false, writeOnlySecretVerifierChange{}, diags
	}

	next, prepareDiags := prepareSecretVerifier(ctx, verifier.hash, credential, prior, valueWO, matches)
	diags.Append(prepareDiags...)

	if diags.HasError() {
		return false, writeOnlySecretVerifierChange{}, diags
	}

	change, encodeDiags := verifier.changeTo(next)
	diags.Append(encodeDiags...)

	return matches, change, diags
}

// Encode public state and identity before publishing the prepared private
// value. Diagnostics do not roll back Framework response mutations.
func publishSigningSecretState(ctx context.Context, private PrivateProviderData, change writeOnlySecretVerifierChange, identity *tfsdk.ResourceIdentity, state *tfsdk.State, identityNames []string, model any) diag.Diagnostics {
	return publishStateAfterPrivate(ctx, identity, state, identityNames, model, func() diag.Diagnostics {
		return change.publish(ctx, private)
	})
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
