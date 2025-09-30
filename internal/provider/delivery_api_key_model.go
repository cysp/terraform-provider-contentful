package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DeliveryAPIKeyIdentityModel struct {
	SpaceID  types.String `tfsdk:"space_id"`
	APIKeyID types.String `tfsdk:"api_key_id"`
}

type DeliveryAPIKeyModel struct {
	DeliveryAPIKeyIdentityModel

	ID              types.String            `tfsdk:"id"`
	Name            types.String            `tfsdk:"name"`
	Description     types.String            `tfsdk:"description"`
	Environments    TypedList[types.String] `tfsdk:"environments"`
	AccessToken     types.String            `tfsdk:"access_token"`
	PreviewAPIKeyID types.String            `tfsdk:"preview_api_key_id"`
}
