package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebhookIdentityModel struct {
	SpaceID   types.String `tfsdk:"space_id"`
	WebhookID types.String `tfsdk:"webhook_id"`
}

type WebhookModel struct {
	IDIdentityModel
	WebhookIdentityModel

	Name              types.String                               `tfsdk:"name"`
	URL               types.String                               `tfsdk:"url"`
	Topics            TypedList[types.String]                    `tfsdk:"topics"`
	Filters           TypedList[TypedObject[WebhookFilterValue]] `tfsdk:"filters"`
	HTTPBasicPassword types.String                               `tfsdk:"http_basic_password"`
	HTTPBasicUsername types.String                               `tfsdk:"http_basic_username"`
	Headers           TypedMap[TypedObject[WebhookHeaderValue]]  `tfsdk:"headers"`
	Transformation    TypedObject[WebhookTransformationValue]    `tfsdk:"transformation"`
	Active            types.Bool                                 `tfsdk:"active"`
}
