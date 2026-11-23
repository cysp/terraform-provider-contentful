package provider

import (
	"context"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const webhookPasswordWO = "http_basic_password_wo" // #nosec G101 -- Terraform attribute name.

func webhookHeaderVerifierPath(key string) writeOnlySecretPath {
	return writeOnlySecretPath{
		key:       "headers[" + strconv.Quote(key) + "].value_wo",
		attribute: path.Root("headers").AtMapKey(key).AtName("value_wo"),
	}
}

func webhookHeaderVerifierKey(encoded string) (string, bool) {
	quoted, ok := strings.CutPrefix(encoded, "headers[")
	if !ok {
		return "", false
	}

	quoted, ok = strings.CutSuffix(quoted, "].value_wo")
	if !ok {
		return "", false
	}

	key, err := strconv.Unquote(quoted)

	return key, err == nil && webhookHeaderVerifierPath(key).key == encoded
}

func webhookVerifierPath(encoded string) bool {
	if encoded == webhookPasswordWO {
		return true
	}

	_, ok := webhookHeaderVerifierKey(encoded)

	return ok
}

func webhookSecretMatches(ctx context.Context, store *writeOnlySecretVerifiers, credentialPath writeOnlySecretPath, prior, candidate types.String, previouslyPresent bool) (bool, diag.Diagnostics) {
	if candidate.IsUnknown() || !previouslyPresent {
		return false, nil
	}

	return writeOnlySecretValueMatches(ctx, store, credentialPath, prior, candidate)
}

// Compare actual apply-time bytes before any mutation. An acknowledged ordinary
// value takes precedence over an obsolete verifier for the same path.
func prepareWebhookCredential(ctx context.Context, store *writeOnlySecretVerifiers, credentialPath writeOnlySecretPath, prior, ordinary, writeOnly types.String, previouslyPresent bool) (bool, diag.Diagnostics) {
	candidate := ordinary
	if !writeOnly.IsNull() {
		candidate = writeOnly
	}

	if candidate.IsUnknown() {
		return false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(credentialPath.attribute, "Unresolved webhook credential", "The credential must be known before the webhook can be written.")}
	}

	if candidate.IsNull() {
		changed := !prior.IsNull() || store.has(credentialPath)

		return !changed, store.remove(credentialPath)
	}

	matches, diags := webhookSecretMatches(ctx, store, credentialPath, prior, candidate, previouslyPresent)
	if diags.HasError() {
		return false, diags
	}

	diags.Append(store.prepareValue(ctx, credentialPath, prior, writeOnly, matches)...)

	return matches, diags
}

// Compare the effective Plan and apply-time credentials before requiring a writable
// full definition. A local update may retain an imported, unreadable credential.
func prepareWebhookWrite(ctx context.Context, state, plan, config WebhookModel, store *writeOnlySecretVerifiers) (WebhookModel, bool, diag.Diagnostics) {
	diags := validateWebhookWriteOnly(plan, config)
	if config.Headers.IsUnknown() {
		diags.AddAttributeError(path.Root("headers"), "Unresolved webhook headers", "Configured headers must be known before the webhook can be written.")
	}

	if diags.HasError() {
		return WebhookModel{}, false, diags
	}

	request := plan
	if !config.HTTPBasicPasswordWO.IsNull() {
		request.HTTPBasicPassword = config.HTTPBasicPasswordWO
	}

	matches := true

	if !state.HTTPBasicPassword.IsNull() || !request.HTTPBasicPassword.IsNull() || plan.HTTPBasicUsername.IsNull() {
		var passwordDiags diag.Diagnostics

		matches, passwordDiags = prepareWebhookCredential(ctx, store, rootWriteOnlySecretPath(webhookPasswordWO), state.HTTPBasicPassword, plan.HTTPBasicPassword, config.HTTPBasicPasswordWO, !state.HTTPBasicUsername.IsNull())
		diags.Append(passwordDiags...)
	}

	needsPUT := !matches || !plan.HTTPBasicUsername.Equal(state.HTTPBasicUsername) ||
		!plan.Name.Equal(state.Name) || !plan.URL.Equal(state.URL) || !plan.Active.Equal(state.Active) ||
		!plan.Topics.Equal(state.Topics) || !plan.Filters.Equal(state.Filters) || !plan.Transformation.Equal(state.Transformation)

	headers, headersChanged, headerDiags := prepareWebhookHeaders(ctx, state, plan, config, store)
	request.Headers = headers

	diags.Append(headerDiags...)

	return request, needsPUT || headersChanged, diags
}

func readWebhookVerifiers(ctx context.Context, private PrivateProviderData, state WebhookModel) (writeOnlySecretVerifiers, diag.Diagnostics) {
	if !webhookReadRedaction(state).any() {
		return readSupersededWriteOnlySecretVerifiers(ctx, private)
	}

	return readWriteOnlySecretVerifiers(ctx, private, webhookVerifierPath, func(encoded string) bool {
		if encoded == webhookPasswordWO {
			return !state.HTTPBasicPassword.IsNull()
		}

		key, _ := webhookHeaderVerifierKey(encoded)
		header, present := state.Headers.Elements()[key].GetValue()

		return present && !header.Value.IsNull()
	})
}

func prepareWebhookHeaders(ctx context.Context, state, plan, config WebhookModel, store *writeOnlySecretVerifiers) (TypedMap[TypedObject[WebhookHeaderValue]], bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	needsPUT := plan.Headers.IsUnknown() || len(plan.Headers.Elements()) != len(state.Headers.Elements())

	headers := maps.Clone(plan.Headers.Elements())
	for _, key := range slices.Sorted(maps.Keys(headers)) {
		header, changed, headerDiags := prepareWebhookHeader(ctx, store, key, state.Headers.Elements()[key], headers[key], config.Headers.Elements()[key])
		headers[key] = header
		needsPUT = needsPUT || changed

		diags.Append(headerDiags...)
	}

	for _, record := range store.records {
		key, ok := webhookHeaderVerifierKey(record.Path)
		if ok {
			if _, present := headers[key]; !present {
				diags.Append(store.remove(webhookHeaderVerifierPath(key))...)
			}
		}
	}

	if plan.Headers.IsNull() || plan.Headers.IsUnknown() {
		return plan.Headers, needsPUT, diags
	}

	return NewTypedMap(headers), needsPUT, diags
}

func prepareWebhookHeader(ctx context.Context, store *writeOnlySecretVerifiers, key string, prior, planned, configured TypedObject[WebhookHeaderValue]) (TypedObject[WebhookHeaderValue], bool, diag.Diagnostics) {
	header, known := planned.GetValue()
	if !known {
		return planned, true, nil
	} // Request conversion reports the unresolved object.

	previous, present := prior.GetValue()

	writeOnly := configured.Value().ValueWO
	if configured.IsUnknown() || writeOnly.IsUnknown() {
		return planned, false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(webhookHeaderVerifierPath(key).attribute, "Unresolved webhook header credential", "The write-only value must be known before the webhook can be written.")}
	}

	changed := !present || !header.Secret.Equal(previous.Secret)
	if configured.IsNull() && present && previous.Secret.ValueBool() && header.Secret.ValueBool() {
		// Sending the exact key without a value preserves any external rotation.
		header.Value = types.StringNull()

		return NewTypedObject(header), changed, nil
	}

	if header.Value.IsNull() && writeOnly.IsNull() {
		// Contentful can retain an existing secret by exact name. A full PUT
		// must still validate an ordinary header whose value is unavailable.
		return planned, changed || (!previous.Value.IsNull() && !header.Secret.ValueBool()), nil
	}

	wasSecret := present && previous.Secret.ValueBool()

	equal, diags := prepareWebhookCredential(ctx, store, webhookHeaderVerifierPath(key), previous.Value, header.Value, writeOnly, present && (!previous.Value.IsNull() || wasSecret))
	if !writeOnly.IsNull() && (!equal || !wasSecret) {
		header.Value = writeOnly
	}

	return NewTypedObject(header), changed || !equal, diags
}
