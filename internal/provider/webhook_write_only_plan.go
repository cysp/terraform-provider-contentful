package provider

import (
	"context"
	"maps"
	"slices"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
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

	if knownWebhookPassword(config.HTTPBasicPasswordWO) && knownWebhookPassword(plan.HTTPBasicPassword) {
		diags.AddAttributeError(path.Root(webhookPasswordWO), "Conflicting webhook Basic passwords", "Configure either http_basic_password or http_basic_password_wo. Remove an ignore_changes rule that retains the ordinary password before migrating.")
	}

	if plan.Headers.IsUnknown() || config.Headers.IsUnknown() {
		return diags
	}

	values := config.writeOnlyHeaderValues()
	for _, key := range slices.Sorted(maps.Keys(values)) {
		if values[key].IsUnknown() {
			continue
		}

		credentialPath := webhookHeaderVerifierPath(key)

		header, ok := plan.Headers.Elements()[key]
		switch {
		case !ok || header.IsNull():
			diags.AddAttributeError(credentialPath.attributePath, "Missing webhook header", "The effective plan must contain this header to apply its write-only value. Check ignore_changes for this key.")
		case header.IsUnknown():
			continue
		default:
			value := header.Value()
			if !value.Secret.IsUnknown() && !value.Secret.ValueBool() {
				diags.AddAttributeError(credentialPath.attributePath, "Write-only webhook header must be secret", "Set secret = true on the matching header.")
			}

			if !value.Value.IsNull() && !value.Value.IsUnknown() {
				diags.AddAttributeError(credentialPath.attributePath, "Conflicting webhook header values", "Configure either an ordinary value or a write-only value for each header. Remove an ignore_changes rule that retains the ordinary value before migrating.")
			}

			for other := range plan.Headers.Elements() {
				if other != key && strings.EqualFold(other, key) {
					diags.AddAttributeError(credentialPath.attributePath, "Ambiguous webhook header name", "Use one spelling of each write-only header name. HTTP header names differing only in case cannot identify independent secrets.")

					break
				}
			}
		}
	}

	return diags
}

func knownWebhookPassword(password types.String) bool {
	return !password.IsUnknown() && !password.IsNull()
}

// An ordinary password in the effective plan still applies when the write-only
// alternative is unknown. A known conflict is diagnosed separately.
func webhookPlannedBasicPassword(plan, config WebhookModel) (types.String, path.Path) {
	if !knownWebhookPassword(plan.HTTPBasicPassword) && !config.HTTPBasicPasswordWO.IsNull() {
		return config.HTTPBasicPasswordWO, path.Root(webhookPasswordWO)
	}

	return plan.HTTPBasicPassword, path.Root("http_basic_password")
}

func validateWebhookBasicPair(plan, config WebhookModel, creating bool) diag.Diagnostics {
	password, passwordPath := webhookPlannedBasicPassword(plan, config)
	username := plan.HTTPBasicUsername

	var diags diag.Diagnostics
	// A PUT with the same username and no password omits both fields.
	if creating || !password.IsNull() {
		diags.Append(validateWebhookUsernameLength(username)...)
	}

	diags.Append(validateWebhookPasswordLength(password, passwordPath)...)

	diags.Append(validateWebhookBasicPairValues(username, password, passwordPath, creating)...)

	return diags
}

func validateWebhookBasicPairValues(username, password types.String, passwordPath path.Path, requirePair bool) diag.Diagnostics {
	var diags diag.Diagnostics
	if username.IsUnknown() || password.IsUnknown() {
		return diags
	}

	switch {
	case requirePair && !username.IsNull() && password.IsNull():
		diags.AddAttributeError(passwordPath, "Missing webhook Basic password", "Supply a password when creating or replacing a webhook with Basic authentication or changing its username.")
	case requirePair && username.IsNull() && !password.IsNull():
		diags.AddAttributeError(path.Root("http_basic_username"), "Missing webhook Basic username", "A password requires a username in the effective plan.")
	case username.IsNull() || password.IsNull():
		// Request conversion distinguishes delegation from a changed username.
	case username.ValueString() == "" && password.ValueString() != "":
		diags.AddAttributeError(path.Root("http_basic_username"), "Incomplete webhook Basic credentials", "Supply a nonempty username with a nonempty password.")
	case username.ValueString() != "" && password.ValueString() == "":
		diags.AddAttributeError(passwordPath, "Incomplete webhook Basic credentials", "Supply a nonempty password with a nonempty username.")
	}

	return diags
}

func validateWebhookUsernameLength(username types.String) diag.Diagnostics {
	if cm.WebhookBasicUsernameTooLong(username.ValueString()) {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root("http_basic_username"), "Webhook Basic credential too long", "The username must fit within 50 UTF-16 code units.")}
	}

	return nil
}

func validateWebhookPasswordLength(password types.String, passwordPath path.Path) diag.Diagnostics {
	if cm.WebhookBasicPasswordTooLong(password.ValueString()) {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(passwordPath, "Webhook Basic credential too long", "The password must fit within 255 UTF-16 code units.")}
	}

	return nil
}

func (r *webhookResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config WebhookModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(validateWebhookWriteOnly(config, config)...)
	}
}

func (r *webhookResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan, config WebhookModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	var state WebhookModel
	if !req.State.Raw.IsNull() {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	headerValues := config.writeOnlyHeaderValues()
	resp.Diagnostics.Append(validatePreliminaryWebhookHeaders(state, plan, config, headerValues)...)

	// Core may preview an update with ignored prior values before planning
	// replacement from configuration. Validate that pair on the create plan.
	if req.State.Raw.IsNull() {
		resp.Diagnostics.Append(validateWebhookBasicPair(plan, config, true)...)

		return
	}

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

	resp.Diagnostics.Append(validatePlannedWebhookUpdateCredentials(ctx, req.Private, plan, state, config, candidate.Raw.IsFullyKnown())...)

	hasWriteOnly := len(headerValues) != 0 || !config.HTTPBasicPasswordWO.IsNull()
	unchanged, compareDiags := webhookRemoteValuesUnchanged(ctx, candidate, req.State)
	resp.Diagnostics.Append(compareDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	scheduled := !candidate.Raw.Equal(req.State.Raw)
	if !unchanged || (hasWriteOnly && scheduled) || config.Headers.IsUnknown() || config.HTTPBasicPasswordWO.IsUnknown() {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("version"), types.Int64Unknown())...)

		return
	}

	if !hasWriteOnly && (!scheduled || !webhookReadRedaction(state).any()) {
		resp.Plan = candidate

		return
	}

	changed, diags := webhookWriteOnlyChangePlanned(ctx, req.Private, state, config, headerValues, scheduled)
	resp.Diagnostics.Append(diags...)

	if changed {
		resp.Diagnostics.Append(candidate.SetAttribute(ctx, path.Root("version"), types.Int64Unknown())...)
	}

	resp.Plan = candidate
}

func validatePlannedWebhookUpdateCredentials(ctx context.Context, private PrivateProviderData, plan, state, config WebhookModel, fullyKnown bool) diag.Diagnostics {
	var diags diag.Diagnostics

	usernameChanged := !plan.HTTPBasicUsername.IsUnknown() && !plan.HTTPBasicUsername.Equal(state.HTTPBasicUsername)
	if usernameChanged && plan.HTTPBasicUsername.Equal(config.HTTPBasicUsername) {
		diags.Append(validateWebhookUsernameLength(plan.HTTPBasicUsername)...)
	}

	password, passwordPath := webhookPlannedBasicPassword(plan, config)

	diags.Append(validatePreliminaryWebhookBasicPair(plan, state, config, password, passwordPath, usernameChanged)...)

	lengthDiags := validateWebhookPasswordLength(password, passwordPath)
	retainedUsernameTooLong := !usernameChanged && cm.WebhookBasicUsernameTooLong(plan.HTTPBasicUsername.ValueString()) && knownWebhookPassword(password)

	// In the preliminary update plan, even Config can contain an ignored
	// ordinary value that Core will discard when planning a replacement.
	if knownWebhookPassword(plan.HTTPBasicPassword) &&
		(password.Equal(state.HTTPBasicPassword) || !password.Equal(config.HTTPBasicPassword)) {
		lengthDiags = nil
	}

	if len(lengthDiags) == 0 && !retainedUsernameTooLong {
		return diags
	}

	if usernameChanged || (fullyKnown && webhookSettingsRequirePUT(plan, state)) ||
		webhookPasswordNeedsWrite(ctx, private, state, plan, config, password, fullyKnown) {
		diags.Append(lengthDiags...)

		if retainedUsernameTooLong {
			diags.AddWarning("Basic username too long for an in-place webhook update", "Terraform’s normalized username exceeds Contentful’s 50 UTF-16 code unit limit. An in-place update that sends a password will fail before writing the webhook. Other resources may change first. Supply a shorter username and its password, or delegate the existing pair. Terraform checks the configured pair separately for a replacement.")
		}
	}

	return diags
}

func validatePreliminaryWebhookBasicPair(plan, state, config WebhookModel, password types.String, passwordPath path.Path, usernameChanged bool) diag.Diagnostics {
	var diags diag.Diagnostics

	// A selected write-only password cannot come from prior state. An ordinary
	// pair in Core's preliminary replacement plan can include an ignored value.
	selectedWriteOnly := passwordPath.Equal(path.Root(webhookPasswordWO)) && knownWebhookPassword(password)

	changedOrdinary := plan.HTTPBasicUsername.Equal(config.HTTPBasicUsername) &&
		!plan.HTTPBasicPassword.Equal(state.HTTPBasicPassword) && plan.HTTPBasicPassword.Equal(config.HTTPBasicPassword)

	pairDiags := validateWebhookBasicPairValues(plan.HTTPBasicUsername, password, passwordPath, true)
	switch {
	case usernameChanged && (selectedWriteOnly || changedOrdinary):
		diags.Append(pairDiags...)
	case pairDiags.HasError() && (usernameChanged || (!password.IsNull() && (selectedWriteOnly || changedOrdinary))):
		// Core also sends this preliminary plan for a valid replacement with
		// an ignored counterpart. Only the create plan can settle that case.
		diags.AddWarning("Incomplete Basic credentials for an in-place webhook update",
			"An in-place update with this Basic username and password would fail. Other resources may already have changed by then. Supply the missing or nonempty counterpart to update the webhook. Terraform checks the configured credentials separately when planning a replacement.")
	}

	return diags
}

// Both password forms can match the last acknowledged value and change only
// Terraform storage. Compare only credentials affected by a length limit here so ordinary plans
// keep their existing verifier cost and replacement can recover corrupt state.
func webhookPasswordNeedsWrite(ctx context.Context, private PrivateProviderData, state, plan, config WebhookModel, password types.String, fullyKnown bool) bool {
	store, diags := readWebhookVerifiers(ctx, private, state)
	if diags.HasError() {
		return false
	}

	matches, diags := webhookSecretMatches(ctx, store.passwordHash(), rootWriteOnlySecretPath(webhookPasswordWO), state.HTTPBasicPassword, password, !state.HTTPBasicUsername.IsNull())
	if diags.HasError() {
		return false
	}

	if !matches {
		return true
	}

	if !fullyKnown || (plan.Headers.Equal(state.Headers) && len(config.writeOnlyHeaderValues()) == 0) {
		return false
	}

	// A matching legacy password will still be sent on a header update.
	_, headersChanged, diags := prepareWebhookHeaders(ctx, state, plan, config, &store)

	return !diags.HasError() && headersChanged
}

func webhookWriteOnlyChangePlanned(ctx context.Context, private PrivateProviderData, state, config WebhookModel, headerValues map[string]types.String, scheduled bool) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	store, hashDiags := readWebhookVerifiers(ctx, private, state)
	for _, d := range hashDiags {
		if invalid, ok := d.(invalidWriteOnlySecretHashesDiagnostic); ok {
			diags.AddWarning(d.Summary(), invalid.reason+" An in-place update will fail before writing the webhook. Restore valid Terraform state before updating, or explicitly replace or destroy the resource.")
		} else {
			diags.Append(d)
		}
	}

	if diags.HasError() {
		return false, diags
	}

	// Core can hide an ignored write-only argument during planning and supply it again at apply.
	changed := hashDiags.HasError() || (scheduled && (store.Basic != nil || len(store.Headers) != 0))
	if !changed && !config.HTTPBasicPasswordWO.IsNull() {
		equal, d := webhookSecretMatches(ctx, store.passwordHash(), rootWriteOnlySecretPath(webhookPasswordWO), state.HTTPBasicPassword, config.HTTPBasicPasswordWO, !state.HTTPBasicUsername.IsNull())
		diags.Append(d...)

		changed = !equal
	}

	for _, key := range slices.Sorted(maps.Keys(headerValues)) {
		if changed || diags.HasError() {
			break
		}

		prior, known := state.Headers.Elements()[key].GetValue()
		equal, d := webhookSecretMatches(ctx, store.Headers[key], webhookHeaderVerifierPath(key), prior.Value, headerValues[key], known && prior.Secret.ValueBool())
		diags.Append(d...)

		changed = !equal
	}

	return changed, diags
}
