package provider

import (
	"context"
	"maps"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Keep unknown header objects as candidates until Terraform resolves their values.
func (config *WebhookModel) writeOnlyHeaderValues() map[string]types.String {
	values := make(map[string]types.String)

	for key, header := range config.Headers.Elements() {
		if header.IsUnknown() {
			values[key] = types.StringUnknown()
		} else if value := header.Value().ValueWO; !value.IsNull() {
			values[key] = value
		}
	}

	return values
}

// Config supplies write-only inputs. All ordinary values, including header
// membership, must use the effective Plan in Create and Update.
func validateWebhookWriteOnly(plan, config WebhookModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if !config.HTTPBasicPasswordWO.IsNull() {
		if !plan.HTTPBasicPassword.IsNull() && !plan.HTTPBasicPassword.IsUnknown() {
			diags.AddAttributeError(path.Root(webhookPasswordWO), "Conflicting webhook Basic passwords", "Configure either http_basic_password or http_basic_password_wo. Remove an ignore_changes rule that retains the ordinary password before migrating.")
		}

		if plan.HTTPBasicUsername.IsNull() {
			diags.AddAttributeError(path.Root("http_basic_username"), "Missing webhook Basic username", "Configure http_basic_username together with http_basic_password_wo.")
		}
	}

	if plan.Headers.IsUnknown() || config.Headers.IsUnknown() {
		return diags
	}

	for _, key := range slices.Sorted(maps.Keys(config.writeOnlyHeaderValues())) {
		credentialPath := path.Root("headers").AtMapKey(key).AtName("value_wo")

		header, ok := plan.Headers.Elements()[key]
		switch {
		case !ok || header.IsNull():
			diags.AddAttributeError(credentialPath, "Missing webhook header", "The effective plan must contain this header to apply its write-only value. Check ignore_changes for this key.")
		case header.IsUnknown():
			continue
		default:
			value := header.Value()
			if !value.Secret.IsUnknown() && !value.Secret.ValueBool() {
				diags.AddAttributeError(credentialPath, "Write-only webhook header must be secret", "Set secret = true on the matching header.")
			}

			if !value.Value.IsNull() && !value.Value.IsUnknown() {
				diags.AddAttributeError(credentialPath, "Conflicting webhook header values", "Configure either an ordinary value or a write-only value for each header. Remove an ignore_changes rule that retains the ordinary value before migrating.")
			}

			for other := range plan.Headers.Elements() {
				if other != key && strings.EqualFold(other, key) {
					diags.AddAttributeError(credentialPath, "Ambiguous webhook header name", "Use one spelling of each write-only header name. HTTP header names differing only in case cannot identify independent secrets.")

					break
				}
			}
		}
	}

	return diags
}

func (r *webhookResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config WebhookModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(validateWebhookWriteOnly(config, config)...)
	}
}

func (r *webhookResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var state, config WebhookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// A full PUT can advance sys.version. Ignore only version when deciding
	// whether an otherwise visible update already leaves room for this.
	candidate := req.Plan
	resp.Diagnostics.Append(candidate.SetAttribute(ctx, path.Root("version"), state.Version)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !candidate.Raw.Equal(req.State.Raw) || config.Headers.IsUnknown() || config.HTTPBasicPasswordWO.IsUnknown() {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("version"), types.Int64Unknown())...)

		return
	}

	headerValues := config.writeOnlyHeaderValues()
	if config.HTTPBasicPasswordWO.IsNull() && len(headerValues) == 0 {
		resp.Plan = candidate

		return
	}

	store, hashDiags := readWebhookVerifiers(ctx, req.Private, state)
	for _, d := range hashDiags {
		if invalid, ok := d.(invalidWriteOnlySecretHashesDiagnostic); ok {
			resp.Diagnostics.AddWarning(d.Summary(), invalid.reason+" An in-place update will fail before writing the webhook. Restore valid Terraform state before updating, or explicitly replace or destroy the resource.")
		} else {
			resp.Diagnostics.Append(d)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	changed := hashDiags.HasError()
	if !changed && !config.HTTPBasicPasswordWO.IsNull() {
		equal, d := webhookSecretMatches(ctx, &store, path.Root(webhookPasswordWO), state.HTTPBasicPassword, config.HTTPBasicPasswordWO, !state.HTTPBasicUsername.IsNull())
		resp.Diagnostics.Append(d...)

		changed = !equal
	}

	for _, key := range slices.Sorted(maps.Keys(headerValues)) {
		if changed || resp.Diagnostics.HasError() {
			break
		}

		prior, known := state.Headers.Elements()[key].GetValue()
		equal, d := webhookSecretMatches(ctx, &store, webhookHeaderVerifierPath(key), prior.Value, headerValues[key], known && prior.Secret.ValueBool())
		resp.Diagnostics.Append(d...)

		changed = !equal
	}

	if changed {
		resp.Diagnostics.Append(candidate.SetAttribute(ctx, path.Root("version"), types.Int64Unknown())...)
	}

	resp.Plan = candidate
}
