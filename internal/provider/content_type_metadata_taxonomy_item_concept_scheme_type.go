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

type ContentTypeMetadataTaxonomyItemConceptSchemeType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = ContentTypeMetadataTaxonomyItemConceptSchemeType{}

func (t ContentTypeMetadataTaxonomyItemConceptSchemeType) Equal(o attr.Type) bool {
	other, ok := o.(ContentTypeMetadataTaxonomyItemConceptSchemeType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ContentTypeMetadataTaxonomyItemConceptSchemeType) String() string {
	return "ContentTypeMetadataTaxonomyItemConceptSchemeType"
}

//nolint:ireturn
func (t ContentTypeMetadataTaxonomyItemConceptSchemeType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, ContentTypeMetadataTaxonomyItemConceptSchemeValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t ContentTypeMetadataTaxonomyItemConceptSchemeType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewContentTypeMetadataTaxonomyItemConceptSchemeValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewContentTypeMetadataTaxonomyItemConceptSchemeValueNull(), nil
	}

	if !value.IsKnown() {
		return NewContentTypeMetadataTaxonomyItemConceptSchemeValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create ContentTypeMetadataTaxonomyItemConceptSchemeValue from Terraform: %w", err)
	}

	v, diags := NewContentTypeMetadataTaxonomyItemConceptSchemeValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t ContentTypeMetadataTaxonomyItemConceptSchemeType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewContentTypeMetadataTaxonomyItemConceptSchemeValueNull(), nil
	case value.IsUnknown():
		return NewContentTypeMetadataTaxonomyItemConceptSchemeValueUnknown(), nil
	}

	return NewContentTypeMetadataTaxonomyItemConceptSchemeValueKnownFromAttributes(ctx, value.Attributes())
}
