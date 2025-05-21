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

type ContentTypeMetadataAnnotationsType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = ContentTypeMetadataAnnotationsType{}

func (t ContentTypeMetadataAnnotationsType) Equal(o attr.Type) bool {
	other, ok := o.(ContentTypeMetadataAnnotationsType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ContentTypeMetadataAnnotationsType) String() string {
	return "ContentTypeMetadataAnnotationsType"
}

//nolint:ireturn
func (t ContentTypeMetadataAnnotationsType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, ContentTypeMetadataAnnotationsValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t ContentTypeMetadataAnnotationsType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewContentTypeMetadataAnnotationsValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewContentTypeMetadataAnnotationsValueNull(), nil
	}

	if !value.IsKnown() {
		return NewContentTypeMetadataAnnotationsValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create ContentTypeMetadataAnnotationsValue from Terraform: %w", err)
	}

	v, diags := NewContentTypeMetadataAnnotationsValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t ContentTypeMetadataAnnotationsType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewContentTypeMetadataAnnotationsValueNull(), nil
	case value.IsUnknown():
		return NewContentTypeMetadataAnnotationsValueUnknown(), nil
	}

	return NewContentTypeMetadataAnnotationsValueKnownFromAttributes(ctx, value.Attributes())
}
