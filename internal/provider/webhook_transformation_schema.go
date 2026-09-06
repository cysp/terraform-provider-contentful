package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

//nolint:ireturn
func WebhookTransformationSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: "Customizes the outgoing webhook HTTP method, Content-Type, Content-Length header, and request body. When omitted, Contentful uses its standard webhook request behavior.",
		Attributes:  WebhookTransformationValue{}.SchemaAttributes(ctx),
		CustomType:  NewTypedObjectNull[WebhookTransformationValue]().CustomType(ctx),
		Optional:    optional,
	}
}
