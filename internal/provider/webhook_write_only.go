package provider

import (
	"context"
	"maps"
	"slices"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const webhookPasswordWO = "http_basic_password_wo" // #nosec G101 -- Terraform attribute name.

func webhookHeaderVerifierPath(key string) writeOnlySecretPath {
	return writeOnlySecretPath{
		key:           "headers[" + strconv.Quote(key) + "].value_wo",
		attributePath: path.Root("headers").AtMapKey(key).AtName("value_wo"),
	}
}

func webhookSecretMatches(ctx context.Context, verifier secretVerifier, credentialPath writeOnlySecretPath, prior, candidate types.String, previouslyPresent bool) (bool, diag.Diagnostics) {
	if candidate.IsUnknown() || !previouslyPresent {
		return false, nil
	}

	return matchesPriorValue(ctx, verifier, credentialPath, prior, candidate)
}

// Return the verifier to publish and whether the wire value can be retained.
func prepareWebhookCredential(ctx context.Context, verifier secretVerifier, credentialPath writeOnlySecretPath, prior, ordinary, writeOnly types.String, previouslyPresent bool) (secretVerifier, bool, diag.Diagnostics) {
	candidate := ordinary
	if !writeOnly.IsNull() {
		candidate = writeOnly
	}

	if candidate.IsUnknown() {
		return "", false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(credentialPath.attributePath, "Unresolved webhook credential", "The credential must be known before the webhook can be written.")}
	}

	if candidate.IsNull() {
		return "", prior.IsNull() && verifier == "", nil
	}

	matches, diags := webhookSecretMatches(ctx, verifier, credentialPath, prior, candidate, previouslyPresent)
	if diags.HasError() {
		return "", false, diags
	}

	next, diags := prepareSecretVerifier(ctx, verifier, credentialPath, prior, writeOnly, matches)

	return next, matches, diags
}

func (s *webhookVerifiers) passwordHash() secretVerifier {
	if s.Basic == nil {
		return ""
	}

	return s.Basic.Hash
}

// Compare the effective Plan and apply-time credentials before requiring a writable
// full definition. A local update may retain an imported, unreadable credential.
func prepareWebhookWrite(ctx context.Context, state, plan, config WebhookModel, store *webhookVerifiers) (WebhookModel, bool, diag.Diagnostics) {
	diags := validateWebhookWriteOnly(plan, config)
	for _, key := range changedWebhookHeaderSecretFlags(state, plan) {
		diags.AddAttributeError(path.Root("headers").AtMapKey(key).AtName("secret"), "Cannot change webhook header secret flag",
			"Contentful keeps the original secret flag when updating a header with the same name. Remove the header in one apply and recreate it in a later apply, or use a new header name.")
	}

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

	passwordMatches, passwordDiags := prepareWebhookPassword(ctx, state, plan, config, store)
	diags.Append(passwordDiags...)

	needsPUT := !passwordMatches || webhookSettingsRequirePUT(plan, state)

	headers, headersChanged, headerDiags := prepareWebhookHeaders(ctx, state, plan, config, store)
	request.Headers = headers

	diags.Append(headerDiags...)

	needsPUT = needsPUT || headersChanged
	if state.WebhookID.IsNull() || needsPUT {
		diags.Append(validateWebhookBasicPair(plan, config, state.WebhookID.IsNull())...)
		diags.Append(validatePlannedWebhookHeaderValues(path.Root("headers"), headers.Elements(), state.Headers.Elements(), nil)...)
	}

	return request, needsPUT, diags
}

func webhookSettingsRequirePUT(plan, state WebhookModel) bool {
	return !plan.HTTPBasicUsername.Equal(state.HTTPBasicUsername) ||
		!plan.Name.Equal(state.Name) || !plan.URL.Equal(state.URL) || !plan.Active.Equal(state.Active) ||
		!plan.Topics.Equal(state.Topics) || !plan.Filters.Equal(state.Filters) || !plan.Transformation.Equal(state.Transformation)
}

func prepareWebhookPassword(ctx context.Context, state, plan, config WebhookModel, store *webhookVerifiers) (bool, diag.Diagnostics) {
	sameUser := !state.HTTPBasicUsername.IsNull() && state.HTTPBasicUsername.Equal(plan.HTTPBasicUsername)
	if sameUser && plan.HTTPBasicPassword.IsNull() && config.HTTPBasicPasswordWO.IsNull() {
		return true, nil
	}

	prior := store.passwordHash()

	next, matches, diags := prepareWebhookCredential(ctx, prior, rootWriteOnlySecretPath(webhookPasswordWO),
		state.HTTPBasicPassword, plan.HTTPBasicPassword, config.HTTPBasicPasswordWO, sameUser)
	if diags.HasError() || prior == next {
		return matches, diags
	}

	store.changed = true
	if next == "" {
		store.Basic = nil
	} else {
		store.Basic = &webhookBasicVerifier{Username: plan.HTTPBasicUsername.ValueString(), Hash: next}
	}

	return matches, diags
}

func readWebhookVerifiers(ctx context.Context, private PrivateProviderData, state WebhookModel) (webhookVerifiers, diag.Diagnostics) {
	return readWebhookVerifierRecords(ctx, private, webhookReadRedaction(state), state.HTTPBasicUsername, true)
}

func prepareWebhookHeaders(ctx context.Context, state, plan, config WebhookModel, store *webhookVerifiers) (TypedMap[TypedObject[WebhookHeaderValue]], bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	needsPUT := plan.Headers.IsUnknown() || len(plan.Headers.Elements()) != len(state.Headers.Elements())

	headers := maps.Clone(plan.Headers.Elements())

	next := make(map[string]secretVerifier, len(store.Headers))
	for _, key := range slices.Sorted(maps.Keys(headers)) {
		header, verifier, changed, headerDiags := prepareWebhookHeader(ctx, store.Headers[key], key, state.Headers.Elements()[key], headers[key], config.Headers.Elements()[key], !plan.URL.Equal(state.URL))

		headers[key] = header
		if verifier != "" {
			next[key] = verifier
		}

		needsPUT = needsPUT || changed

		diags.Append(headerDiags...)
	}

	store.changed = store.changed || !maps.Equal(store.Headers, next)
	store.Headers = next

	if plan.Headers.IsNull() || plan.Headers.IsUnknown() {
		return plan.Headers, needsPUT, diags
	}

	return NewTypedMap(headers), needsPUT, diags
}

func prepareWebhookHeader(ctx context.Context, verifier secretVerifier, key string, prior, planned, configured TypedObject[WebhookHeaderValue], receiverChanged bool) (TypedObject[WebhookHeaderValue], secretVerifier, bool, diag.Diagnostics) {
	header, known := planned.GetValue()
	if !known {
		return planned, verifier, true, nil
	} // Request conversion reports the unresolved object.

	previous, present := prior.GetValue()
	wasSecret := previous.Secret.ValueBool()

	writeOnly := configured.Value().ValueWO
	if configured.IsUnknown() || writeOnly.IsUnknown() {
		return planned, verifier, false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(webhookHeaderVerifierPath(key).attributePath, "Unresolved webhook header credential", "The write-only value must be known before the webhook can be written.")}
	}

	changed := !present || !header.Secret.Equal(previous.Secret)
	if configured.IsNull() && wasSecret && header.Secret.ValueBool() {
		// Sending the exact key without a value preserves any external rotation.
		header.Value = types.StringNull()

		return NewTypedObject(header), verifier, changed, nil
	}

	if header.Value.IsNull() && writeOnly.IsNull() {
		// Contentful can retain an existing secret by exact name. A full PUT
		// must still validate an ordinary header whose value is unavailable.
		return planned, verifier, changed || (!previous.Value.IsNull() && !header.Secret.ValueBool()), nil
	}

	hash, equal, diags := prepareWebhookCredential(ctx, verifier, webhookHeaderVerifierPath(key), previous.Value, header.Value, writeOnly, !previous.Value.IsNull() || wasSecret)
	if !writeOnly.IsNull() && (!equal || receiverChanged) {
		header.Value = writeOnly
	}

	return NewTypedObject(header), hash, changed || !equal, diags
}
