package provider

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IDIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func NewIDIdentityModelFromMultipartID(id []string) IDIdentityModel {
	return IDIdentityModel{
		ID: types.StringValue(strings.Join(id, "/")),
	}
}
