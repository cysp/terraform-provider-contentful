package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebhookModel struct {
	Active            types.Bool                 `tfsdk:"active"`
	Filters           types.List                 `tfsdk:"filters"`
	Headers           types.Map                  `tfsdk:"headers"`
	HTTPBasicPassword types.String               `tfsdk:"http_basic_password"`
	HTTPBasicUsername types.String               `tfsdk:"http_basic_username"`
	Name              types.String               `tfsdk:"name"`
	SpaceID           types.String               `tfsdk:"space_id"`
	Topics            types.List                 `tfsdk:"topics"`
	Transformation    WebhookTransformationValue `tfsdk:"transformation"`
	URL               types.String               `tfsdk:"url"`
	WebhookID         types.String               `tfsdk:"webhook_id"`
}
