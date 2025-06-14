package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PersonalAccessTokenModel struct {
	ID        types.String            `tfsdk:"id"`
	Name      types.String            `tfsdk:"name"`
	ExpiresIn types.Int64             `tfsdk:"expires_in"`
	ExpiresAt timetypes.RFC3339       `tfsdk:"expires_at"`
	RevokedAt timetypes.RFC3339       `tfsdk:"revoked_at"`
	Scopes    TypedList[types.String] `tfsdk:"scopes"`
	Token     types.String            `tfsdk:"token"`
}
