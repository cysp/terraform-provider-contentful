package provider

import (
	"context"
	"strconv"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	webhookPasswordWO     = "http_basic_password_wo" // #nosec G101 -- Terraform attribute name.
	webhookHeaderValuesWO = "header_values_wo"
)

func webhookHeaderVerifierPath(key string) path.Path {
	return path.Root(webhookHeaderValuesWO).AtMapKey(key)
}

func webhookHeaderVerifierKey(encoded string) (string, bool) {
	quoted, ok := strings.CutPrefix(encoded, webhookHeaderValuesWO+"[")
	if !ok {
		return "", false
	}

	quoted, ok = strings.CutSuffix(quoted, "]")
	if !ok {
		return "", false
	}

	key, err := strconv.Unquote(quoted)

	return key, err == nil && webhookHeaderVerifierPath(key).String() == encoded
}

func webhookVerifierPath(encoded string) bool {
	if encoded == webhookPasswordWO {
		return true
	}

	_, ok := webhookHeaderVerifierKey(encoded)

	return ok
}

func webhookSecretMatches(ctx context.Context, store *writeOnlySecretVerifiers, credentialPath path.Path, prior, candidate types.String, previouslyPresent bool) (bool, diag.Diagnostics) {
	if candidate.IsUnknown() || !previouslyPresent {
		return false, nil
	}

	if !prior.IsNull() {
		return prior.Equal(candidate), nil
	}

	return store.matches(ctx, credentialPath, candidate)
}

// Compare actual apply-time bytes before any mutation. An acknowledged ordinary
// value takes precedence over an obsolete verifier for the same path.
func prepareWebhookCredential(ctx context.Context, store *writeOnlySecretVerifiers, credentialPath path.Path, prior, ordinary, writeOnly types.String, previouslyPresent bool) (bool, diag.Diagnostics) {
	candidate := ordinary
	if !writeOnly.IsNull() {
		candidate = writeOnly
	}

	if candidate.IsUnknown() {
		return false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(credentialPath, "Unresolved webhook credential", "The credential must be known before the webhook can be written.")}
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

// A request owns its own header slice. Only that HTTP model ever receives
// write-only plaintext; response reconciliation receives the untouched Plan.
func prepareWebhookWrite(ctx context.Context, request *cm.WebhookDefinitionData, state, plan, config WebhookModel, store *writeOnlySecretVerifiers) (bool, diag.Diagnostics) {
	diags := validateWebhookWriteOnly(plan, config)
	if config.HeaderValuesWO.IsUnknown() {
		diags.AddAttributeError(path.Root(webhookHeaderValuesWO), "Unresolved webhook header values", "The write-only header map must be known before the webhook can be written.")
	}

	if diags.HasError() {
		return false, diags
	}

	passwordChanged, passwordDiags := prepareWebhookBasicAuth(ctx, store, request, state, plan, config.HTTPBasicPasswordWO)
	diags.Append(passwordDiags...)

	if diags.HasError() {
		return false, diags
	}

	needsPUT := passwordChanged || plan.Headers.IsUnknown() ||
		!plan.Name.Equal(state.Name) || !plan.URL.Equal(state.URL) || !plan.Active.Equal(state.Active) ||
		!plan.Topics.Equal(state.Topics) || !plan.Filters.Equal(state.Filters) || !plan.Transformation.Equal(state.Transformation) ||
		len(request.Headers) != len(state.Headers.Elements())

	for i := range request.Headers {
		header := &request.Headers[i]
		changed, headerDiags := prepareWebhookHeader(ctx, store, header, state, plan, config)
		diags.Append(headerDiags...)

		needsPUT = needsPUT || changed
	}

	for _, encoded := range store.paths() {
		key, ok := webhookHeaderVerifierKey(encoded)
		if ok {
			if _, present := plan.Headers.Elements()[key]; !present {
				diags.Append(store.remove(webhookHeaderVerifierPath(key))...)
			}
		}
	}

	return needsPUT, diags
}

func prepareWebhookBasicAuth(ctx context.Context, store *writeOnlySecretVerifiers, request *cm.WebhookDefinitionData, state, plan WebhookModel, writeOnly types.String) (bool, diag.Diagnostics) {
	passwordPath := path.Root(webhookPasswordWO)
	if plan.HTTPBasicPassword.IsNull() && writeOnly.IsNull() && !plan.HTTPBasicUsername.IsNull() {
		return false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(passwordPath, "Missing webhook Basic password", "Supply a password with the username, or omit both credentials to clear Basic authentication.")}
	}

	if !plan.HTTPBasicPassword.IsNull() && plan.HTTPBasicUsername.IsNull() {
		return false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root("http_basic_username"), "Missing webhook Basic username", "Supply a username together with the password, or omit both credentials to clear Basic authentication.")}
	}

	equal, diags := prepareWebhookCredential(ctx, store, passwordPath, state.HTTPBasicPassword, plan.HTTPBasicPassword, writeOnly, !state.HTTPBasicUsername.IsNull())
	if !writeOnly.IsNull() {
		request.HttpBasicPassword = cm.NewOptNilString(writeOnly.ValueString())
	}

	return !equal || !plan.HTTPBasicUsername.Equal(state.HTTPBasicUsername), diags
}

func prepareWebhookHeader(ctx context.Context, store *writeOnlySecretVerifiers, wire *cm.WebhookDefinitionHeader, state, plan, config WebhookModel) (bool, diag.Diagnostics) {
	credentialPath := webhookHeaderVerifierPath(wire.Key)
	prior, present := state.Headers.Elements()[wire.Key].GetValue()
	planned := plan.Headers.Elements()[wire.Key].Value()
	writeOnly, supplied := config.HeaderValuesWO.Elements()[wire.Key]
	wasSecret := present && prior.Secret.ValueBool()
	changed := !present || !planned.Secret.Equal(prior.Secret)

	if supplied && writeOnly.IsNull() {
		return false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(credentialPath, "Null write-only webhook header value", "Supply a non-null value, or omit the map entry to retain an existing secret header.")}
	}

	if !supplied && planned.Value.IsNull() {
		if !wasSecret {
			return false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root("headers").AtMapKey(wire.Key).AtName("value"), "Missing webhook header value", "Supply a value for a new header. A value-less secret header can retain only an existing secret with the same name.")}
		}

		return changed, nil
	}

	equal, diags := prepareWebhookCredential(ctx, store, credentialPath, prior.Value, planned.Value, writeOnly, present && (!prior.Value.IsNull() || wasSecret))
	if supplied && (!equal || !wasSecret) {
		wire.Value = cm.NewOptString(writeOnly.ValueString())
	}

	return changed || !equal, diags
}

func (model *WebhookModel) hasRedactedCredentials() bool {
	if !model.HTTPBasicUsername.IsNull() && model.HTTPBasicPassword.IsNull() {
		return true
	}

	for _, item := range model.Headers.Elements() {
		if header, ok := item.GetValue(); ok && header.Secret.ValueBool() && header.Value.IsNull() {
			return true
		}
	}

	return false
}

func readWebhookVerifiers(ctx context.Context, private PrivateProviderData, state WebhookModel) (writeOnlySecretVerifiers, diag.Diagnostics) {
	if !state.hasRedactedCredentials() {
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
