package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

//nolint:ireturn
func WebhookHeadersSchema(ctx context.Context) schema.Attribute {
	return schema.MapNestedAttribute{
		Description: "HTTP headers sent by the webhook, keyed by header name. Omitting headers or setting `{}` removes all managed headers on apply. To keep externally managed headers, use `lifecycle.ignore_changes` on their keys or on the whole map. See [Secrets and Terraform state](../guides/secrets-and-state#secret-custom-headers).",
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
