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

type ContentTypeFieldAllowedResourceItemExternalType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = ContentTypeFieldAllowedResourceItemExternalType{}

func (t ContentTypeFieldAllowedResourceItemExternalType) Equal(o attr.Type) bool {
	other, ok := o.(ContentTypeFieldAllowedResourceItemExternalType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ContentTypeFieldAllowedResourceItemExternalType) String() string {
	return "ContentTypeFieldAllowedResourceItemExternalType"
}

//nolint:ireturn
func (t ContentTypeFieldAllowedResourceItemExternalType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, ContentTypeFieldAllowedResourceItemExternalValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t ContentTypeFieldAllowedResourceItemExternalType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewContentTypeFieldAllowedResourceItemExternalValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewContentTypeFieldAllowedResourceItemExternalValueNull(), nil
	}

	if !value.IsKnown() {
		return NewContentTypeFieldAllowedResourceItemExternalValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create EditorInterfaceEditorLayoutItemGroupItemExternalValue from Terraform: %w", err)
	}

	v, diags := NewContentTypeFieldAllowedResourceItemExternalValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t ContentTypeFieldAllowedResourceItemExternalType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewContentTypeFieldAllowedResourceItemExternalValueNull(), nil
	case value.IsUnknown():
		return NewContentTypeFieldAllowedResourceItemExternalValueUnknown(), nil
	}

	return NewContentTypeFieldAllowedResourceItemExternalValueKnownFromAttributes(ctx, value.Attributes())
}
