package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func CopyAttributeValues(ctx context.Context, dst any, src any) diag.Diagnostics {
	dstAttributeTypes := tpfr.AttributeTypesOf(ctx, dst)

	srcAttributeValues := tpfr.AttributeValuesOf(src)

	attributeValues := make(map[string]attr.Value, len(dstAttributeTypes))

	for fieldName := range dstAttributeTypes {
		attributeValues[fieldName] = srcAttributeValues[fieldName]
	}

	return tpfr.SetAttributeValues(ctx, dst, attributeValues)
}
