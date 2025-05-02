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

type ContentTypeFieldAllowedResourceItemType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = ContentTypeFieldAllowedResourceItemType{}

func (t ContentTypeFieldAllowedResourceItemType) Equal(o attr.Type) bool {
	other, ok := o.(ContentTypeFieldAllowedResourceItemType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ContentTypeFieldAllowedResourceItemType) String() string {
	return "ContentTypeFieldAllowedResourceItemType"
}

//nolint:ireturn
func (t ContentTypeFieldAllowedResourceItemType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, ContentTypeFieldAllowedResourceItemValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t ContentTypeFieldAllowedResourceItemType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewContentTypeFieldAllowedResourceItemValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewContentTypeFieldAllowedResourceItemValueNull(), nil
	}

	if !value.IsKnown() {
		return NewContentTypeFieldAllowedResourceItemValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create ContentTypeFieldAllowedResourceItemValue from Terraform: %w", err)
	}

	v, diags := NewContentTypeFieldAllowedResourceItemValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t ContentTypeFieldAllowedResourceItemType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewContentTypeFieldAllowedResourceItemValueNull(), nil
	case value.IsUnknown():
		return NewContentTypeFieldAllowedResourceItemValueUnknown(), nil
	}

	return NewContentTypeFieldAllowedResourceItemValueKnownFromAttributes(ctx, value.Attributes())
}
