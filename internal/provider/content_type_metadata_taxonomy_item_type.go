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

type ContentTypeMetadataTaxonomyItemType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = ContentTypeMetadataTaxonomyItemType{}

func (t ContentTypeMetadataTaxonomyItemType) Equal(o attr.Type) bool {
	other, ok := o.(ContentTypeMetadataTaxonomyItemType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ContentTypeMetadataTaxonomyItemType) String() string {
	return "ContentTypeMetadataTaxonomyItemType"
}

//nolint:ireturn
func (t ContentTypeMetadataTaxonomyItemType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, ContentTypeMetadataTaxonomyItemValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t ContentTypeMetadataTaxonomyItemType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewContentTypeMetadataTaxonomyItemValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewContentTypeMetadataTaxonomyItemValueNull(), nil
	}

	if !value.IsKnown() {
		return NewContentTypeMetadataTaxonomyItemValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create ContentTypeMetadataTaxonomyItemValue from Terraform: %w", err)
	}

	v, diags := NewContentTypeMetadataTaxonomyItemValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t ContentTypeMetadataTaxonomyItemType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewContentTypeMetadataTaxonomyItemValueNull(), nil
	case value.IsUnknown():
		return NewContentTypeMetadataTaxonomyItemValueUnknown(), nil
	}

	return NewContentTypeMetadataTaxonomyItemValueKnownFromAttributes(ctx, value.Attributes())
}
