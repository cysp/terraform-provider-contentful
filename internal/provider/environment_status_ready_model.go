package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnvironmentStatusReadyModel struct {
	IDIdentityModel
	EnvironmentIdentityModel

	Status types.String `tfsdk:"status"`
}
