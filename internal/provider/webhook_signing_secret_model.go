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

	Value    types.String   `tfsdk:"value"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
