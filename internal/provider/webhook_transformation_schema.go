package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

//nolint:ireturn
func WebhookTransformationSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		Attributes: WebhookTransformationValue{}.SchemaAttributes(ctx),
		CustomType: NewTypedObjectNull[WebhookTransformationValue]().CustomType(ctx),
		Optional:   optional,
	}
}
