package provider

import (
	"bytes"
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const webhookPasswordWO = "http_basic_password_wo" // #nosec G101 -- durable attribute name.

func (r *webhookResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var (
		state      WebhookModel
		configured types.String
	)

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(webhookPasswordWO), &configured)...)

	if resp.Diagnostics.HasError() {
		return
	}

	versionPath := path.Root("version")
	candidate := req.Plan
	resp.Diagnostics.Append(candidate.SetAttribute(ctx, versionPath, state.Version)...)
	unchanged, diags := webhookRemoteValuesUnchanged(ctx, candidate, req.State)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	scheduled := !candidate.Raw.Equal(req.State.Raw)
	if !unchanged || configured.IsUnknown() || (scheduled && !configured.IsNull()) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, versionPath, types.Int64Unknown())...)

		return
	}

	possible := false

	if !configured.IsNull() || (scheduled && webhookBasicPasswordRedacted(state)) {
		verifier, hashDiags := readWriteOnlySecretVerifier(ctx, req.Private, webhookPasswordWO, state.HTTPBasicPassword)
		for _, d := range hashDiags {
			if invalid, ok := d.(invalidWriteOnlySecretHashesDiagnostic); ok {
				resp.Diagnostics.AddWarning(invalid.Summary(), invalid.reason+" Restore valid Terraform state before updating the webhook, or explicitly replace or destroy it.")
			} else {
				resp.Diagnostics.Append(d)
			}
		}

		if resp.Diagnostics.HasError() {
			return
		}

		// Core can hide an ignored write-only password in planning Config and
		// restore its apply-time value on a separately scheduled update.
		possible = hashDiags.HasError() || (scheduled && verifier.key != nil)
		if !possible && !configured.IsNull() {
			equal, comparisonDiags := writeOnlySecretValueMatches(ctx, &verifier, state.HTTPBasicPassword, configured)
			resp.Diagnostics.Append(comparisonDiags...)

			possible = !equal
		}
	}

	if possible {
		resp.Diagnostics.Append(candidate.SetAttribute(ctx, versionPath, types.Int64Unknown())...)
	}

	resp.Plan = candidate
}

// Prepare verifier changes before the write. Keep the public plan separate
// from the full request, which needs plaintext on every PUT retaining Basic auth.
func prepareWebhookBasicPassword(ctx context.Context, state, plan, config WebhookModel, verifier *writeOnlySecretVerifier) (WebhookModel, bool, diag.Diagnostics) {
	request := plan

	writeOnly := config.HTTPBasicPasswordWO
	if !writeOnly.IsNull() {
		if !plan.HTTPBasicPassword.IsNull() {
			return request, false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root(webhookPasswordWO), "Conflicting webhook Basic passwords", "Configure either http_basic_password or http_basic_password_wo. Remove an ignore_changes rule that retains the ordinary password before migrating.")}
		}

		if plan.HTTPBasicUsername.IsNull() {
			return request, false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root("http_basic_username"), "Missing webhook Basic username", "Configure a username with the write-only password.")}
		}

		request.HTTPBasicPassword = writeOnly
	}

	if request.HTTPBasicPassword.IsUnknown() {
		return request, false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root("http_basic_password"), "Unresolved webhook Basic password", "The password must be known before the webhook can be written.")}
	}

	if request.HTTPBasicPassword.IsNull() {
		if plan.HTTPBasicUsername.IsNull() {
			verifier.clear()
		}

		return request, state.HTTPBasicPassword.IsNull(), nil
	}

	matches, diags := prepareWriteOnlySecretValue(ctx, verifier, state.HTTPBasicPassword, request.HTTPBasicPassword.ValueString(), writeOnly)

	return request, matches, diags
}

func webhookBasicPasswordRedacted(state WebhookModel) bool {
	return state.HTTPBasicPassword.IsNull() && !state.HTTPBasicUsername.IsNull()
}

func checkWebhookBasicPasswordResponse(response *cm.WebhookDefinition, ordinary types.String) diag.Diagnostics {
	returned := types.StringPointerValue(response.HttpBasicPassword.ValueStringPointer())
	if !returned.IsNull() && !returned.Equal(ordinary) {
		response.HttpBasicPassword.Reset()

		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root("http_basic_password"), "Unexpected webhook Basic password in response", "Contentful returned a readable password at a redacted path. The provider discarded the response password.")}
	}

	return nil
}

func publishWebhookBasicResponse(ctx context.Context, private PrivateProviderData, verifier *writeOnlySecretVerifier, identity *tfsdk.ResourceIdentity, target *tfsdk.State, response cm.WebhookDefinition, state, plan WebhookModel, protected bool) diag.Diagnostics {
	if state.WebhookID.IsNull() && response.Sys.ID == "" {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Missing created webhook ID", "Contentful accepted the create request without returning a webhook ID. Inspect the space for the new webhook before retrying.")}
	}

	var secretDiags diag.Diagnostics
	if protected || (webhookBasicPasswordRedacted(state) && plan.HTTPBasicPassword.IsNull()) {
		secretDiags.Append(checkWebhookBasicPasswordResponse(&response, plan.HTTPBasicPassword)...)

		if protected && plan.HTTPBasicPassword.IsNull() && !plan.HTTPBasicUsername.IsNull() && response.HttpBasicPassword.IsNull() {
			secretDiags.AddAttributeError(path.Root("http_basic_password"), "Unconfirmed webhook Basic password", "Contentful returned an explicit null password for the configured Basic credentials.")
		}
	}

	data, diags, consistency, confirmed := ReconcileWebhookMutationResponse(ctx, response, plan, state.WebhookID)
	if diags.HasError() {
		return diags
	}

	data.Timeouts = plan.Timeouts

	consistency.Append(secretDiags...)

	if consistency.HasError() && protected {
		data.HTTPBasicPassword = types.StringNull()
		if data.HTTPBasicUsername.IsNull() {
			data.HTTPBasicUsername = plan.HTTPBasicUsername
			if data.HTTPBasicUsername.IsNull() {
				data.HTTPBasicUsername = state.HTTPBasicUsername
			}
		}
	}

	if consistency.HasError() || bytes.Equal(verifier.original, verifier.prepared) {
		diags.Append(setResourceIdentityStateAndMutationVersion(ctx, identity, target, private, webhookIdentityAttributeNames(), &data, response.Sys.Version, confirmed)...)
	} else {
		diags.Append(publishStateAfterPrivate(ctx, identity, target, webhookIdentityAttributeNames(), &data, func() diag.Diagnostics {
			d := setResourceMutationVersion(ctx, private, response.Sys.Version, confirmed)
			if !d.HasError() {
				d.Append(verifier.publish(ctx, private)...)
			}

			return d
		})...)
	}

	diags.Append(consistency...)

	return diags
}
