package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IDIdentityModel struct {
	ID types.String `tfsdk:"id"`
}
