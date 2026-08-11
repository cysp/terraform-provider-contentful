package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebhookHeaderValue struct {
	Value  types.String `tfsdk:"value"`
	Secret types.Bool   `tfsdk:"secret"`
}
