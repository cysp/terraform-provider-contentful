package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebhookSigningSecretIdentityModel struct {
	SpaceID types.String `tfsdk:"space_id"`
}

type WebhookSigningSecretModel struct {
	IDIdentityModel
	WebhookSigningSecretIdentityModel

	Value     types.String      `tfsdk:"value"`
	ValueWO   types.String      `tfsdk:"value_wo"`
	CreatedAt timetypes.RFC3339 `tfsdk:"created_at"`
	UpdatedAt timetypes.RFC3339 `tfsdk:"updated_at"`
	Timeouts  timeouts.Value    `tfsdk:"timeouts"`
}
