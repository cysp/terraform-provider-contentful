package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

//nolint:ireturn
func WebhookHeadersSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.MapNestedAttribute{
		NestedObject: schema.NestedAttributeObject{
			Attributes: WebhookHeaderValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[WebhookHeaderValue]().CustomType(ctx),
		},
		CustomType: TypedMap[TypedObject[WebhookHeaderValue]]{}.CustomType(ctx),
		Optional:   optional,
	}
}
