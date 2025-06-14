package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebhookModel struct {
	ID                types.String                  `tfsdk:"id"`
	SpaceID           types.String                  `tfsdk:"space_id"`
	WebhookID         types.String                  `tfsdk:"webhook_id"`
	Name              types.String                  `tfsdk:"name"`
	URL               types.String                  `tfsdk:"url"`
	Topics            TypedList[types.String]       `tfsdk:"topics"`
	Filters           TypedList[WebhookFilterValue] `tfsdk:"filters"`
	HTTPBasicPassword types.String                  `tfsdk:"http_basic_password"`
	HTTPBasicUsername types.String                  `tfsdk:"http_basic_username"`
	Headers           TypedMap[WebhookHeaderValue]  `tfsdk:"headers"`
	Transformation    WebhookTransformationValue    `tfsdk:"transformation"`
	Active            types.Bool                    `tfsdk:"active"`
}
