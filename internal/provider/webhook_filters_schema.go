package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

//nolint:ireturn
func WebhookFiltersSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.ListNestedAttribute{
		NestedObject: schema.NestedAttributeObject{
			Attributes: WebhookFilterValue{}.SchemaAttributes(ctx),
			CustomType: WebhookFilterValue{}.CustomType(ctx),
		},
		CustomType: NewTypedListNull[WebhookFilterValue](ctx).CustomType(ctx),
		Optional:   optional,
	}
}
