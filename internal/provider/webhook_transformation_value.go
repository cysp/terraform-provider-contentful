package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebhookTransformationValue struct {
	Method               types.String         `tfsdk:"method"`
	ContentType          types.String         `tfsdk:"content_type"`
	IncludeContentLength types.Bool           `tfsdk:"include_content_length"`
	Body                 jsontypes.Normalized `tfsdk:"body"`
}
