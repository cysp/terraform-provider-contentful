package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebhookHeaderValue struct {
	Value   types.String `tfsdk:"value"`
	ValueWO types.String `tfsdk:"value_wo"`
	Secret  types.Bool   `tfsdk:"secret"`
}
