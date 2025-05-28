//nolint:dupl
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type ContentTypeMetadataType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = ContentTypeMetadataType{}

func (t ContentTypeMetadataType) Equal(o attr.Type) bool {
	other, ok := o.(ContentTypeMetadataType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ContentTypeMetadataType) String() string {
	return "ContentTypeMetadataType"
}

//nolint:ireturn
func (t ContentTypeMetadataType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, ContentTypeMetadataValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t ContentTypeMetadataType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewContentTypeMetadataValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewContentTypeMetadataValueNull(), nil
	}

	if !value.IsKnown() {
		return NewContentTypeMetadataValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create ContentTypeMetadataValue from Terraform: %w", err)
	}

	v, diags := NewContentTypeMetadataValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t ContentTypeMetadataType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewContentTypeMetadataValueNull(), nil
	case value.IsUnknown():
		return NewContentTypeMetadataValueUnknown(), nil
	}

	return NewContentTypeMetadataValueKnownFromAttributes(ctx, value.Attributes())
}
