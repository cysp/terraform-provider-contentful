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

type ContentTypeFieldAllowedResourceItemContentfulEntryType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = ContentTypeFieldAllowedResourceItemContentfulEntryType{}

func (t ContentTypeFieldAllowedResourceItemContentfulEntryType) Equal(o attr.Type) bool {
	other, ok := o.(ContentTypeFieldAllowedResourceItemContentfulEntryType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ContentTypeFieldAllowedResourceItemContentfulEntryType) String() string {
	return "ContentTypeFieldAllowedResourceItemContentfulEntryType"
}

//nolint:ireturn
func (t ContentTypeFieldAllowedResourceItemContentfulEntryType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, ContentTypeFieldAllowedResourceItemContentfulEntryValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t ContentTypeFieldAllowedResourceItemContentfulEntryType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewContentTypeFieldAllowedResourceItemContentfulEntryValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewContentTypeFieldAllowedResourceItemContentfulEntryValueNull(), nil
	}

	if !value.IsKnown() {
		return NewContentTypeFieldAllowedResourceItemContentfulEntryValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create EditorInterfaceEditorLayoutItemGroupItemContentfulEntryValue from Terraform: %w", err)
	}

	v, diags := NewContentTypeFieldAllowedResourceItemContentfulEntryValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t ContentTypeFieldAllowedResourceItemContentfulEntryType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewContentTypeFieldAllowedResourceItemContentfulEntryValueNull(), nil
	case value.IsUnknown():
		return NewContentTypeFieldAllowedResourceItemContentfulEntryValueUnknown(), nil
	}

	return NewContentTypeFieldAllowedResourceItemContentfulEntryValueKnownFromAttributes(ctx, value.Attributes())
}
