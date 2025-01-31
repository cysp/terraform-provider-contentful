package provider

import (
	"context"

	webhookfilter "github.com/cysp/terraform-provider-contentful/internal/provider/webhook_filter"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

//nolint:ireturn
func WebhookFiltersSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.ListNestedAttribute{
		NestedObject: schema.NestedAttributeObject{
			Attributes: webhookfilter.WebhookFilterValue{}.SchemaAttributes(ctx),
			CustomType: webhookfilter.WebhookFilterValue{}.CustomType(ctx),
		},
		Optional: optional,
	}
}
