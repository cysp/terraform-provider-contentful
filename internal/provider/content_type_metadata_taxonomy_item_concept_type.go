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

type ContentTypeMetadataTaxonomyItemConceptType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = ContentTypeMetadataTaxonomyItemConceptType{}

func (t ContentTypeMetadataTaxonomyItemConceptType) Equal(o attr.Type) bool {
	other, ok := o.(ContentTypeMetadataTaxonomyItemConceptType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ContentTypeMetadataTaxonomyItemConceptType) String() string {
	return "ContentTypeMetadataTaxonomyItemConceptType"
}

//nolint:ireturn
func (t ContentTypeMetadataTaxonomyItemConceptType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, ContentTypeMetadataTaxonomyItemConceptValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t ContentTypeMetadataTaxonomyItemConceptType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewContentTypeMetadataTaxonomyItemConceptValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewContentTypeMetadataTaxonomyItemConceptValueNull(), nil
	}

	if !value.IsKnown() {
		return NewContentTypeMetadataTaxonomyItemConceptValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create ContentTypeMetadataTaxonomyItemConceptValue from Terraform: %w", err)
	}

	v, diags := NewContentTypeMetadataTaxonomyItemConceptValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t ContentTypeMetadataTaxonomyItemConceptType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewContentTypeMetadataTaxonomyItemConceptValueNull(), nil
	case value.IsUnknown():
		return NewContentTypeMetadataTaxonomyItemConceptValueUnknown(), nil
	}

	return NewContentTypeMetadataTaxonomyItemConceptValueKnownFromAttributes(ctx, value.Attributes())
}
