package provider

import (
	"context"
	"maps"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//nolint:ireturn
func WebhookHeadersSchema(ctx context.Context) schema.Attribute {
	return schema.MapNestedAttribute{
		Description: "HTTP headers sent by the webhook, keyed by header name. Omitting the map or setting `headers = {}` removes all headers unless `ignore_changes` retains the map or a key. To keep an externally managed header, omit its key and ignore that exact key. See [Secrets and Terraform state](../guides/secrets-and-state#secret-custom-headers).",
		NestedObject: schema.NestedAttributeObject{
			Attributes: WebhookHeaderValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookHeaderValue]().CustomType(ctx),
		},
		CustomType: TypedMap[TypedObject[WebhookHeaderValue]]{}.CustomType(ctx),
		Optional:   true,
		Validators: []validator.Map{
			mapvalidator.NoNullValues(),
		},
	}
}

func validatePlannedWebhookHeaderValues(headersPath path.Path, headers map[string]TypedObject[WebhookHeaderValue], writeOnly map[string]types.String, creating bool) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, key := range slices.Sorted(maps.Keys(headers)) {
		header, known := headers[key].GetValue()
		if !known {
			continue
		}

		headerPath := headersPath.AtMapKey(key)
		if creating && header.Value.IsNull() && writeOnly[key].IsNull() {
			diags.AddAttributeError(headerPath.AtName("value"), "Missing initial webhook header value",
				"Supply an initial value for this header when creating or replacing a webhook.")

			continue
		}

		diags.Append(header.validateValueOmission(headerPath)...)
	}

	return diags
}

// Contentful retains the secret flag for the lifetime of an exact header name.
func changedWebhookHeaderSecretFlags(state, plan WebhookModel) []string {
	// A different space replaces the webhook. None of its header flags carry over.
	if !state.SpaceID.Equal(plan.SpaceID) {
		return nil
	}

	var changed []string

	for _, key := range slices.Sorted(maps.Keys(plan.Headers.Elements())) {
		before := state.Headers.Elements()[key].Value().Secret
		after := plan.Headers.Elements()[key].Value().Secret

		if before.IsNull() || before.IsUnknown() || after.IsNull() || after.IsUnknown() || before.Equal(after) {
			continue
		}

		changed = append(changed, key)
	}

	return changed
}

func validatePreliminaryWebhookHeaders(state, plan, config WebhookModel, headerValues map[string]types.String) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, key := range changedWebhookHeaderSecretFlags(state, plan) {
		diags.AddWarning("Header flag change requires Webhook replacement or removal", path.Root("headers").AtMapKey(key).AtName("secret").String()+": An in-place update cannot change an existing header’s secret flag. Other resources may change before the update fails. A replacement creates new headers. To keep this webhook, remove the header in one apply and recreate it in another, or choose a new header name.")
	}

	if !config.Headers.IsUnknown() && !plan.Headers.Equal(state.Headers) {
		diags.Append(validatePlannedWebhookHeaderValues(path.Root("headers"), plan.Headers.Elements(), headerValues, state.WebhookID.IsNull())...)
	}

	return diags
}
