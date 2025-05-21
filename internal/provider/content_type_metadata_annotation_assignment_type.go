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

type ContentTypeMetadataAnnotationAssignmentType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = ContentTypeMetadataAnnotationAssignmentType{}

func (t ContentTypeMetadataAnnotationAssignmentType) Equal(o attr.Type) bool {
	other, ok := o.(ContentTypeMetadataAnnotationAssignmentType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ContentTypeMetadataAnnotationAssignmentType) String() string {
	return "ContentTypeMetadataAnnotationAssignmentType"
}

//nolint:ireturn
func (t ContentTypeMetadataAnnotationAssignmentType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, ContentTypeMetadataAnnotationAssignmentValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t ContentTypeMetadataAnnotationAssignmentType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewContentTypeMetadataAnnotationAssignmentValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewContentTypeMetadataAnnotationAssignmentValueNull(), nil
	}

	if !value.IsKnown() {
		return NewContentTypeMetadataAnnotationAssignmentValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create ContentTypeMetadataAnnotationAssignmentValue from Terraform: %w", err)
	}

	v, diags := NewContentTypeMetadataAnnotationAssignmentValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t ContentTypeMetadataAnnotationAssignmentType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewContentTypeMetadataAnnotationAssignmentValueNull(), nil
	case value.IsUnknown():
		return NewContentTypeMetadataAnnotationAssignmentValueUnknown(), nil
	}

	return NewContentTypeMetadataAnnotationAssignmentValueKnownFromAttributes(ctx, value.Attributes())
}
