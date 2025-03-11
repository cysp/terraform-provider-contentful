package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type ContentTypeFieldType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = ContentTypeFieldType{}

func (t ContentTypeFieldType) Equal(o attr.Type) bool {
	other, ok := o.(ContentTypeFieldType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

//nolint:ireturn
func (t ContentTypeFieldType) ValueType(_ context.Context) attr.Value {
	return ContentTypeFieldValue{}
}

func (t ContentTypeFieldType) String() string {
	return "ContentTypeFieldType"
}

//nolint:ireturn
func (t ContentTypeFieldType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, ContentTypeFieldValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t ContentTypeFieldType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewContentTypeFieldValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewContentTypeFieldValueNull(), nil
	}

	if !value.IsKnown() {
		return NewContentTypeFieldValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create ContentTypeFieldValue from Terraform: %w", err)
	}

	v, diags := NewContentTypeFieldValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t ContentTypeFieldType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewContentTypeFieldValueNull(), nil
	case value.IsUnknown():
		return NewContentTypeFieldValueUnknown(), nil
	}

	return NewContentTypeFieldValueKnownFromAttributes(ctx, value.Attributes())
}
