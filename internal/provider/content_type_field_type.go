package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type FieldsType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = FieldsType{}

func (t FieldsType) Equal(o attr.Type) bool {
	other, ok := o.(FieldsType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

//nolint:ireturn
func (t FieldsType) ValueType(_ context.Context) attr.Value {
	return FieldsValue{}
}

func (t FieldsType) String() string {
	return "FieldsType"
}

//nolint:ireturn
func (t FieldsType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: t.TerraformAttributeTypes(ctx),
	}
}

func (t FieldsType) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"id":            tftypes.String,
		"name":          tftypes.String,
		"type":          tftypes.String,
		"link_type":     tftypes.String,
		"disabled":      tftypes.Bool,
		"omitted":       tftypes.Bool,
		"required":      tftypes.Bool,
		"default_value": jsontypes.NormalizedType{}.TerraformType(ctx),
		"items":         ContentTypeFieldItemsType{}.TerraformType(ctx),
		"localized":     tftypes.Bool,
		"validations":   tftypes.List{ElementType: jsontypes.NormalizedType{}.TerraformType(ctx)},
	}
}

//nolint:ireturn
func (t FieldsType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewFieldsValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewFieldsValueNull(), nil
	}

	if !value.IsKnown() {
		return NewFieldsValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create FieldsValue from Terraform: %w", err)
	}

	v, diags := NewContentTypeFieldValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t FieldsType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewFieldsValueNull(), nil
	case value.IsUnknown():
		return NewFieldsValueUnknown(), nil
	}

	return NewContentTypeFieldValueKnownFromAttributes(ctx, value.Attributes())
}
