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
