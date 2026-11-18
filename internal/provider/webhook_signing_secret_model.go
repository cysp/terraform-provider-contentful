package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebhookSigningSecretIdentityModel struct {
	SpaceID types.String `tfsdk:"space_id"`
}

type WebhookSigningSecretModel struct {
	IDIdentityModel
	WebhookSigningSecretIdentityModel

	Value     types.String   `tfsdk:"value"`
	ValueWO   types.String   `tfsdk:"value_wo"`
	CreatedAt types.String   `tfsdk:"created_at"`
	UpdatedAt types.String   `tfsdk:"updated_at"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}
