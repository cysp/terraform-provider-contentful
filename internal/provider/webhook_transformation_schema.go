package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

//nolint:ireturn
func WebhookTransformationSchema(ctx context.Context, optional bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: "Customizes the outgoing webhook HTTP method, Content-Type, Content-Length header, and request body. When omitted, Contentful uses its standard webhook request behavior. Unsupported methods and content types prevent writes, including values retained by ignore_changes. Read and import preserve returned values. To destroy with unsupported configuration, correct the configuration and retry destroy; no intervening apply is required.",
		Attributes:  WebhookTransformationValue{}.SchemaAttributes(ctx),
		CustomType:  NewTypedObjectNull[WebhookTransformationValue]().CustomType(ctx),
		Optional:    optional,
	}
}
