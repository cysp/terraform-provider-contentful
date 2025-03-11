package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func ObjectAttrTypesToTerraformTypes(ctx context.Context, attrTypes map[string]attr.Type) map[string]tftypes.Type {
	tftyp := make(map[string]tftypes.Type, len(attrTypes))

	for key, attrType := range attrTypes {
		tftyp[key] = attrType.TerraformType(ctx)
	}

	return tftyp
}
