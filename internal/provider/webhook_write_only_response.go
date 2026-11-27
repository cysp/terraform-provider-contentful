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
	guard.password = mask.password || (plan.HTTPBasicPassword.IsNull() && (guard.password || !plan.HTTPBasicUsername.IsNull()))

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
func (mask webhookRedaction) retainUnconfirmed(data *WebhookModel, state, plan WebhookModel, sanitized cm.WebhookDefinition, confirmed bool) {
	// Only an explicit ordinary echo from the target can release a guard.
	// Reconciled state can also contain planned fallbacks for omitted fields.
	password, present := sanitized.HttpBasicPassword.Get()

	ordinaryPassword := confirmed && present && data.HTTPBasicUsername.Equal(plan.HTTPBasicUsername) &&
		plan.HTTPBasicPassword.Equal(types.StringValue(password))
	if mask.password && !ordinaryPassword {
		data.HTTPBasicPassword = types.StringNull()
		if !confirmed || data.HTTPBasicUsername.IsNull() {
			data.HTTPBasicUsername = state.HTTPBasicUsername
			if data.HTTPBasicUsername.IsNull() {
				data.HTTPBasicUsername = plan.HTTPBasicUsername
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

	var ordinary map[string]bool
	if confirmed {
		ordinary = matchingOrdinaryWebhookHeaders(sanitized.Headers, plan.Headers)
	}

	for key := range mask.headers {
		if ordinary[key] {
			continue
		}

		for returned := range headers {
			if returned != key && strings.EqualFold(returned, key) && !mask.headers[returned] && !ordinary[returned] {
				delete(headers, returned)
			}
		}

		headers[key] = NewTypedObject(WebhookHeaderValue{Secret: types.BoolValue(true)})
	}

	data.Headers = NewTypedMap(headers)
}

// Redaction has already discarded ambiguous or unauthorized echoes. A
// fallback from the plan cannot establish that the endpoint returned a value.
func matchingOrdinaryWebhookHeaders(headers cm.WebhookDefinitionHeaders, plan TypedMap[TypedObject[WebhookHeaderValue]]) map[string]bool {
	matching := make(map[string]bool)

	for _, header := range headers {
		planned := plan.Elements()[header.Key].Value()

		value, present := header.Value.Get()
		if present && planned.Value.Equal(types.StringValue(value)) && planned.Secret.Equal(types.BoolValue(header.Secret.Or(false))) {
			matching[header.Key] = true
		}
	}

	return matching
}

func webhookVerifiersAfterRead(ctx context.Context, private PrivateProviderData, prior, refreshed WebhookModel, passwordNull bool) (writeOnlySecretVerifierChange, diag.Diagnostics) {
	before, after := webhookReadRedaction(prior), webhookReadRedaction(refreshed)

	sameUser := prior.HTTPBasicUsername.Equal(refreshed.HTTPBasicUsername)
	if maps.Equal(before.headers, after.headers) && before.password == after.password &&
		(!before.password || (sameUser && !passwordNull)) {
		return writeOnlySecretVerifierChange{}, nil
	}

	after.password = before.password && after.password && sameUser && !passwordNull
	maps.DeleteFunc(after.headers, func(key string, _ bool) bool { return !before.headers[key] })
	store, diags := readWebhookVerifierRecords(ctx, private, after, refreshed.HTTPBasicUsername, false)

	if diags.HasError() {
		// An unreadable envelope remains unusable for comparison. Read and
		// Delete remain available, but private access failures are still errors.
		var accessDiags diag.Diagnostics

		for _, diagnostic := range diags {
			if _, invalid := diagnostic.(invalidWriteOnlySecretHashesDiagnostic); !invalid {
				accessDiags.Append(diagnostic)
			}
		}

		return writeOnlySecretVerifierChange{}, accessDiags
	}

	change, encodeDiags := store.encode()
	diags.Append(encodeDiags...)

	return change, diags
}

// An inconsistent response can still identify a newly created webhook. Keep
// its ordinary recovery state and endpoint version while retaining only
// previously acknowledged credentials and verifiers.
func publishWebhookResponse(ctx context.Context, private PrivateProviderData, priorVerifiers webhookVerifiers, change writeOnlySecretVerifierChange, identity *tfsdk.ResourceIdentity, target *tfsdk.State, response cm.WebhookDefinition, state, plan WebhookModel, mask webhookRedaction, sentPassword bool) diag.Diagnostics {
	guard := mask.guardMutationResponse(state, plan)
	sanitized, secretDiags := guard.redact(response, plan)

	if sentPassword && plan.HTTPBasicPassword.IsNull() && response.HttpBasicPassword.IsNull() {
		secretDiags.AddAttributeError(path.Root("http_basic_password"), "Unconfirmed webhook Basic password", "Contentful returned an explicit null password for the configured Basic credentials.")
	}

	data, diags, consistency, confirmed := ReconcileWebhookMutationResponse(ctx, sanitized, plan, state.WebhookID)
	if diags.HasError() {
		return diags
	}

	data.Timeouts = plan.Timeouts

	consistency.Append(secretDiags...)

	if consistency.HasError() {
		// Use the original password field: redaction can discard a returned value.
		if confirmed && response.HttpBasicUsername.IsNull() && response.HttpBasicPassword.ValueStringPointer() == nil {
			guard.password = false
		}

		if confirmed && response.Headers != nil {
			priorVerifiers.discardAbsentHeaders(data.Headers)
			maps.DeleteFunc(guard.headers, func(key string, _ bool) bool {
				_, planned := plan.Headers.Elements()[key]
				_, returned := data.Headers.Elements()[key]

				return !planned && !returned
			})
		}

		guard.retainUnconfirmed(&data, state, plan, sanitized, confirmed)
		priorVerifiers.retain(webhookReadRedaction(data), data.HTTPBasicUsername)

		var encodeDiags diag.Diagnostics

		change, encodeDiags = priorVerifiers.encode()
		diags.Append(encodeDiags...)

		if diags.HasError() {
			diags.Append(consistency...)

			return diags
		}
	}

	diags.Append(publishWebhookState(ctx, private, change, identity, target, data, response.Sys.Version, confirmed)...)

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
