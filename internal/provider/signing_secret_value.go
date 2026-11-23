package provider

import (
	"context"

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

// Encode public state and identity before publishing the prepared private
// value. Diagnostics do not roll back Framework response mutations.
func publishSigningSecretState(ctx context.Context, private PrivateProviderData, verifier *writeOnlySecretVerifier, identity *tfsdk.ResourceIdentity, state *tfsdk.State, identityNames []string, model any) diag.Diagnostics {
	return publishStateAfterPrivate(ctx, identity, state, identityNames, model, func() diag.Diagnostics {
		return verifier.publish(ctx, private)
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
