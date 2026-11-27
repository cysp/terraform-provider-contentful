package provider

import (
	"context"
	"maps"
	"slices"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// A full PUT can include ordinary and write-only credentials together. Record
// which individual paths require redaction before staging verifier changes.
type webhookRedaction struct {
	password bool
	headers  map[string]bool
}

func (mask webhookRedaction) any() bool {
	return mask.password || len(mask.headers) != 0
}

func webhookMutationRedaction(prior webhookVerifiers, config WebhookModel) webhookRedaction {
	mask := webhookRedaction{password: !config.HTTPBasicPasswordWO.IsNull(), headers: make(map[string]bool)}
	for key := range config.writeOnlyHeaderValues() {
		mask.headers[key] = true
	}

	mask.password = mask.password || prior.Basic != nil
	for key := range prior.Headers {
		mask.headers[key] = true
	}

	return mask
}

func webhookReadRedaction(state WebhookModel) webhookRedaction {
	mask := webhookRedaction{password: state.HTTPBasicPassword.IsNull() && !state.HTTPBasicUsername.IsNull(), headers: make(map[string]bool)}
	for key, item := range state.Headers.Elements() {
		if header := item.Value(); header.Secret.ValueBool() && header.Value.IsNull() {
			mask.headers[key] = true
		}
	}

	return mask
}

// An unreadable import or failed first migration can leave a redacted path
// without a verifier. Guard an unsolicited echo while its ordinary value is
// absent from the effective plan, even when this PUT only delegates the path.
func (mask webhookRedaction) guardMutationResponse(state, plan WebhookModel) webhookRedaction {
	guard := webhookReadRedaction(state)
	guard.password = (guard.password && plan.HTTPBasicPassword.IsNull()) || mask.password

	for key := range guard.headers {
		if !plan.Headers.Elements()[key].Value().Value.IsNull() {
			delete(guard.headers, key)
		}
	}

	maps.Copy(guard.headers, mask.headers)

	return guard
}

// Only an exact ordinary value from the effective plan permits a previously
// redacted path to become readable, including after a change in header case.
func (mask webhookRedaction) redact(response cm.WebhookDefinition, plan WebhookModel) (cm.WebhookDefinition, diag.Diagnostics) {
	var diags diag.Diagnostics

	returned := types.StringPointerValue(response.HttpBasicPassword.ValueStringPointer())

	if mask.password && !returned.IsNull() && !returned.Equal(plan.HTTPBasicPassword) {
		diags.AddAttributeError(path.Root("http_basic_password"), "Unexpected webhook Basic password in response", "Contentful returned a password field for a redacted credential. The provider discarded the response password.")
		response.HttpBasicPassword.Reset()
	}

	headers, headerDiags := mask.redactHeaders(response.Headers, plan.Headers)
	response.Headers = headers

	diags.Append(headerDiags...)

	return response, diags
}

func (mask webhookRedaction) redactHeaders(headers cm.WebhookDefinitionHeaders, planned TypedMap[TypedObject[WebhookHeaderValue]]) (cm.WebhookDefinitionHeaders, diag.Diagnostics) {
	var diags diag.Diagnostics

	protected := make(map[string]bool, len(mask.headers))
	for key := range mask.headers {
		protected[strings.ToLower(key)] = true
	}

	// On import the response itself can contain both a redacted header and
	// a readable case alias. Detect duplicates before projecting into a map.
	occurrences := make(map[string]int, len(headers))
	for _, header := range headers {
		occurrences[strings.ToLower(header.Key)]++

		if header.Secret.Or(false) && !header.Value.IsSet() && planned.Elements()[header.Key].Value().Value.IsNull() {
			protected[strings.ToLower(header.Key)] = true
		}
	}

	seen := make(map[string]bool, len(headers))

	headers = slices.Clone(headers)
	for i := range headers {
		header := &headers[i]
		name := strings.ToLower(header.Key)

		secret := header.Secret.Or(false)
		ordinary := planned.Elements()[header.Key].Value().Value

		if protected[name] {
			if seen[header.Key] || (occurrences[name] > 1 && (header.Value.IsSet() || !secret)) {
				diags.AddAttributeError(path.Root("headers"), "Ambiguous webhook secret header in response", "Contentful returned multiple occurrences of a protected header. The response cannot acknowledge its value.")
				header.Value.Reset()
				header.Secret = cm.NewOptBool(true)
			}
			// Only the exact effective ordinary value is authorized for public
			// state, including when changing a former write-only header to ordinary.
			if header.Value.IsSet() || !secret {
				readable, present := header.Value.Get()
				if !present || ordinary.IsNull() || ordinary.IsUnknown() || readable != ordinary.ValueString() {
					diags.AddAttributeError(path.Root("headers").AtMapKey(header.Key), "Unexpected webhook secret header in response", "Contentful returned a readable value or an ordinary header at a redacted path. The provider discarded the response value.")
					header.Value.Reset()
					header.Secret = cm.NewOptBool(true)
				}
			}
		}

		seen[header.Key] = true
	}

	return headers, diags
}

// Keep the redacted descriptor even when the response omits or renames that
// credential. The next Read can then reject an unexpected echo. Ordinary
// siblings already have their reconciled response or effective-plan fallback.
func (mask webhookRedaction) retainUnconfirmed(data *WebhookModel, state, plan WebhookModel) {
	if mask.password {
		data.HTTPBasicPassword = types.StringNull()
		if data.HTTPBasicUsername.IsNull() {
			data.HTTPBasicUsername = plan.HTTPBasicUsername
			if data.HTTPBasicUsername.IsNull() {
				data.HTTPBasicUsername = state.HTTPBasicUsername
			}
		}
	}

	if len(mask.headers) == 0 {
		return
	}

	headers := maps.Clone(data.Headers.Elements())
	if headers == nil {
		headers = make(map[string]TypedObject[WebhookHeaderValue])
	}

	for key := range mask.headers {
		for returned := range headers {
			if returned != key && strings.EqualFold(returned, key) && !mask.headers[returned] {
				delete(headers, returned)
			}
		}

		headers[key] = NewTypedObject(WebhookHeaderValue{Secret: types.BoolValue(true)})
	}

	data.Headers = NewTypedMap(headers)
}

// A Read that witnesses a credential disappear invalidates its comparison
// baseline. A recreated same-name header may use different secret bytes.
func removedWebhookVerifiers(prior, refreshed WebhookModel) []writeOnlySecretPath {
	var removed []writeOnlySecretPath
	if !prior.HTTPBasicUsername.IsNull() && prior.HTTPBasicPassword.IsNull() && !prior.HTTPBasicUsername.Equal(refreshed.HTTPBasicUsername) {
		removed = append(removed, rootWriteOnlySecretPath(webhookPasswordWO))
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

func webhookVerifiersAfterRead(ctx context.Context, private PrivateProviderData, prior, refreshed WebhookModel) (writeOnlySecretVerifierChange, diag.Diagnostics) {
	removed := removedWebhookVerifiers(prior, refreshed)
	if len(removed) == 0 {
		return writeOnlySecretVerifierChange{}, nil
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

		return writeOnlySecretVerifierChange{}, accessDiags
	}

	store.retain(webhookReadRedaction(refreshed), refreshed.HTTPBasicUsername)

	change, encodeDiags := store.encode()
	diags.Append(encodeDiags...)

	return change, diags
}

// An inconsistent response can still identify a newly created webhook. Keep
// its ordinary recovery state and endpoint version while retaining only
// previously acknowledged credentials and verifiers.
func publishWebhookResponse(ctx context.Context, private PrivateProviderData, change writeOnlySecretVerifierChange, identity *tfsdk.ResourceIdentity, target *tfsdk.State, response cm.WebhookDefinition, state, plan WebhookModel, mask webhookRedaction) diag.Diagnostics {
	if state.WebhookID.IsNull() && response.Sys.ID == "" {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Missing created webhook ID", "Contentful returned an empty ID. The provider could not identify the created webhook or acknowledge its credentials. Inspect webhooks in the space before retrying.")}
	}

	sanitized, secretDiags := mask.guardMutationResponse(state, plan).redact(response, plan)
	if mask.password && plan.HTTPBasicPassword.IsNull() && !plan.HTTPBasicUsername.IsNull() && response.HttpBasicPassword.IsNull() {
		secretDiags.AddAttributeError(path.Root("http_basic_password"), "Unconfirmed webhook Basic password", "Contentful returned an explicit null password for the configured Basic credentials.")
	}

	data, diags, consistency, confirmed := ReconcileWebhookMutationResponse(ctx, sanitized, plan, state.WebhookID)
	if diags.HasError() {
		return diags
	}

	data.Timeouts = plan.Timeouts

	consistency.Append(secretDiags...)

	if consistency.HasError() {
		mask.retainUnconfirmed(&data, state, plan)
		diags.Append(setResourceIdentityStateAndMutationVersion(ctx, identity, target, private, webhookIdentityAttributeNames(), &data, response.Sys.Version, confirmed)...)
	} else {
		diags.Append(publishWebhookState(ctx, private, change, identity, target, data, response.Sys.Version, confirmed)...)
	}

	diags.Append(consistency...)

	return diags
}

func publishWebhookState(ctx context.Context, private PrivateProviderData, change writeOnlySecretVerifierChange, identity *tfsdk.ResourceIdentity, state *tfsdk.State, model WebhookModel, version int, confirmed bool) diag.Diagnostics {
	if !change.changed {
		// With no verifier publication pending, retain the ordinary webhook's
		// state/identity checkpoint even if recording the version fails.
		return setResourceIdentityStateAndMutationVersion(ctx, identity, state, private, webhookIdentityAttributeNames(), &model, version, confirmed)
	}

	return publishStateAfterPrivate(ctx, identity, state, webhookIdentityAttributeNames(), &model, func() diag.Diagnostics {
		// A private write failure may leave a newer lock version for a confirmed
		// endpoint. It must never acknowledge unrecorded secret bytes.
		diags := setResourceMutationVersion(ctx, private, version, confirmed)
		if !diags.HasError() {
			diags.Append(change.publish(ctx, private)...)
		}

		return diags
	})
}
