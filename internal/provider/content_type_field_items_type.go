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

type ContentTypeFieldItemsType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = ContentTypeFieldItemsType{}

func (t ContentTypeFieldItemsType) Equal(o attr.Type) bool {
	other, ok := o.(ContentTypeFieldItemsType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ContentTypeFieldItemsType) String() string {
	return "ContentTypeFieldItemsType"
}

//nolint:ireturn
func (t ContentTypeFieldItemsType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, ContentTypeFieldItemsValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t ContentTypeFieldItemsType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewContentTypeFieldItemsValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewContentTypeFieldItemsValueNull(), nil
	}

	if !value.IsKnown() {
		return NewContentTypeFieldItemsValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create ContentTypeFieldItemsValue from Terraform: %w", err)
	}

	v, diags := NewContentTypeFieldItemsValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t ContentTypeFieldItemsType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewContentTypeFieldItemsValueNull(), nil
	case value.IsUnknown():
		return NewContentTypeFieldItemsValueUnknown(), nil
	}

	return NewContentTypeFieldItemsValueKnownFromAttributes(ctx, value.Attributes())
}
