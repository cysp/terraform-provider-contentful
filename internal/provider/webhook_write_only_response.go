package provider

import (
	"bytes"
	"context"
	"slices"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CMA's redaction contract lets ordinary state retain a prior value. For a
// write-only value an unexpected echo must never reach that fallback, even
// if the response changed the header's case or its secret flag.
func redactWebhookResponse(response cm.WebhookDefinition, state, plan, config WebhookModel) (cm.WebhookDefinition, diag.Diagnostics) {
	var diags diag.Diagnostics

	returned := types.StringPointerValue(response.HttpBasicPassword.ValueStringPointer())

	redactedPassword := state.HTTPBasicPassword.IsNull() && (!state.HTTPBasicUsername.IsNull() || plan.HTTPBasicPassword.IsNull())
	if !response.HttpBasicPassword.IsEmpty() && (!config.HTTPBasicPasswordWO.IsNull() || (redactedPassword && !returned.IsNull() && !returned.Equal(plan.HTTPBasicPassword))) {
		diags.AddAttributeError(path.Root("http_basic_password"), "Unexpected webhook Basic password in response", "Contentful returned a password field for a redacted credential. The provider discarded the response password.")
		response.HttpBasicPassword.Reset()
	}

	protected := protectedWebhookHeaderNames(response.Headers, state, plan, config)
	seen := make(map[string]bool, len(response.Headers))

	response.Headers = slices.Clone(response.Headers)
	for i := range response.Headers {
		header := &response.Headers[i]
		name := strings.ToLower(header.Key)

		secret := header.Secret.Or(false)
		ordinary := plan.Headers.Elements()[header.Key].Value().Value

		if protected[name] {
			if seen[name] {
				diags.AddAttributeError(path.Root("headers"), "Ambiguous webhook secret header in response", "Contentful returned multiple occurrences of a protected header. The response cannot acknowledge its value.")
			}
			// Only the exact effective ordinary value is authorized for public
			// state, including when changing a former write-only header to ordinary.
			if header.Value.IsSet() || !secret {
				readable, present := header.Value.Get()
				if !present || ordinary.IsNull() || ordinary.IsUnknown() || readable != ordinary.ValueString() {
					diags.AddAttributeError(path.Root("headers").AtMapKey(header.Key), "Unexpected webhook secret header in response", "Contentful returned a readable value or an ordinary header at a redacted path. The provider discarded the response value.")
					header.Value.Reset()
				}
			}
		}

		seen[name] = true
	}

	return response, diags
}

func protectedWebhookHeaderNames(headers cm.WebhookDefinitionHeaders, state, plan, config WebhookModel) map[string]bool {
	protected := make(map[string]bool)
	for key := range config.writeOnlyHeaderValues() {
		protected[strings.ToLower(key)] = true
	}

	for key, item := range state.Headers.Elements() {
		if header := item.Value(); header.Secret.ValueBool() && header.Value.IsNull() {
			protected[strings.ToLower(key)] = true
		}
	}

	// On import only the response identifies redacted keys. A later readable
	// duplicate cannot override that evidence when the headers become a map.
	for _, header := range headers {
		if header.Secret.Or(false) && plan.Headers.Elements()[header.Key].Value().Value.IsNull() && state.Headers.Elements()[header.Key].Value().Value.IsNull() {
			protected[strings.ToLower(header.Key)] = true
		}
	}

	return protected
}

func retainUnconfirmedWebhookCredentials(data *WebhookModel, state, config WebhookModel, response cm.WebhookDefinition) {
	if !config.HTTPBasicPasswordWO.IsNull() || state.HTTPBasicPassword.IsNull() {
		data.HTTPBasicPassword = types.StringNull()
	} else if response.HttpBasicPassword.IsEmpty() {
		data.HTTPBasicPassword = state.HTTPBasicPassword
	}

	retained := make(map[string]TypedObject[WebhookHeaderValue])
	protected := make(map[string]bool)

	for key, item := range state.Headers.Elements() {
		prior := item.Value()
		if prior.Secret.ValueBool() && prior.Value.IsNull() {
			retained[key] = item
			protected[strings.ToLower(key)] = true
		}
	}

	for key := range config.writeOnlyHeaderValues() {
		name := strings.ToLower(key)
		if !protected[name] {
			// An ambiguous ordinary-to-write-only migration can no longer use
			// its old public value without losing this path's redaction on Read.
			retained[key] = NewTypedObject(WebhookHeaderValue{Secret: types.BoolValue(true), Value: types.StringNull()})
			protected[name] = true
		}
	}

	redacted := make(map[string]bool)

	for _, header := range response.Headers {
		if header.Secret.Or(false) && !header.Value.IsSet() {
			redacted[header.Key] = true
		}
	}

	for key, item := range data.Headers.Elements() {
		if !protected[strings.ToLower(key)] {
			header := item.Value()
			if redacted[key] {
				// A different redacted ordinary header may also have been changed in
				// the same unconfirmed write. Preserve only its previous public bytes.
				prior := state.Headers.Elements()[key].Value()
				header.Value = prior.Value
			}

			retained[key] = NewTypedObject(header)
		}
	}

	data.Headers = NewTypedMap(retained)
}

// A Read that witnesses a credential disappear invalidates its comparison
// baseline. A recreated same-name header may use different secret bytes.
func removedWebhookVerifiers(prior, refreshed WebhookModel) []path.Path {
	var removed []path.Path
	if !prior.HTTPBasicUsername.IsNull() && prior.HTTPBasicPassword.IsNull() && !prior.HTTPBasicUsername.Equal(refreshed.HTTPBasicUsername) {
		removed = append(removed, path.Root(webhookPasswordWO))
	}

	for key, item := range prior.Headers.Elements() {
		header := item.Value()

		next, present := refreshed.Headers.Elements()[key].GetValue()
		if header.Secret.ValueBool() && header.Value.IsNull() && (!present || !next.Secret.ValueBool()) {
			removed = append(removed, webhookHeaderVerifierPath(key))
		}
	}

	return removed
}

func webhookVerifiersAfterRead(ctx context.Context, private PrivateProviderData, prior, refreshed WebhookModel) (writeOnlySecretVerifiers, diag.Diagnostics) {
	removed := removedWebhookVerifiers(prior, refreshed)
	if len(removed) == 0 {
		return writeOnlySecretVerifiers{}, nil
	}

	store, diags := readWebhookVerifiers(ctx, private, prior)
	if diags.HasError() {
		// Keep malformed contents in private state. Refresh and destruction
		// remain available; subsequent Update fails closed without a baseline.
		var accessDiags diag.Diagnostics

		for _, diagnostic := range diags {
			if _, invalid := diagnostic.(invalidWriteOnlySecretHashesDiagnostic); !invalid {
				accessDiags.Append(diagnostic)
			}
		}

		return writeOnlySecretVerifiers{}, accessDiags
	}

	for _, credentialPath := range removed {
		diags.Append(store.remove(credentialPath)...)
	}

	return store, diags
}

// An inconsistent response can still identify a newly created webhook. Keep
// its ordinary recovery state and endpoint version while retaining only
// previously acknowledged credentials and verifiers.
func publishWebhookResponse(ctx context.Context, private PrivateProviderData, store *writeOnlySecretVerifiers, identity *tfsdk.ResourceIdentity, target *tfsdk.State, response cm.WebhookDefinition, state, plan, config WebhookModel, protected bool) diag.Diagnostics {
	if state.WebhookID.IsNull() && response.Sys.ID == "" {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Missing created webhook ID", "Contentful returned an empty ID. The provider could not identify the created webhook or acknowledge its credentials. Inspect webhooks in the space before retrying.")}
	}

	sanitized, secretDiags := redactWebhookResponse(response, state, plan, config)

	data, diags, consistency, confirmed := ReconcileWebhookMutationResponse(ctx, sanitized, plan, state.WebhookID)
	if diags.HasError() {
		return diags
	}

	data.Timeouts = plan.Timeouts

	consistency.Append(secretDiags...)

	if !confirmed {
		data.Version = types.Int64Null()
	}

	if consistency.HasError() && (protected || secretDiags.HasError()) {
		retainUnconfirmedWebhookCredentials(&data, state, config, sanitized)
		diags.Append(setResourceIdentityStateAndMutationVersion(ctx, identity, target, private, webhookIdentityAttributeNames(), &data, response.Sys.Version, confirmed)...)
	} else {
		diags.Append(publishWebhookState(ctx, private, store, identity, target, data, response.Sys.Version, confirmed)...)
	}

	diags.Append(consistency...)

	return diags
}

func publishWebhookState(ctx context.Context, private PrivateProviderData, store *writeOnlySecretVerifiers, identity *tfsdk.ResourceIdentity, state *tfsdk.State, model WebhookModel, version int, confirmed bool) diag.Diagnostics {
	if bytes.Equal(store.original, store.prepared) {
		// With no verifier publication pending, retain the ordinary webhook's
		// state/identity checkpoint even if recording the version fails.
		return setResourceIdentityStateAndMutationVersion(ctx, identity, state, private, webhookIdentityAttributeNames(), &model, version, confirmed)
	}

	return publishStateAfterPrivate(ctx, identity, state, webhookIdentityAttributeNames(), &model, func() diag.Diagnostics {
		// A private write failure may leave a newer lock version for a confirmed
		// endpoint. It must never acknowledge unrecorded secret bytes.
		diags := setResourceMutationVersion(ctx, private, version, confirmed)
		if !diags.HasError() {
			diags.Append(store.publish(ctx, private)...)
		}

		return diags
	})
}
