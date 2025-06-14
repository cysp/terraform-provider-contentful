package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type ContentTypeFieldItemsValue struct {
	ItemsType   types.String                    `tfsdk:"type"`
	LinkType    types.String                    `tfsdk:"link_type"`
	Validations TypedList[jsontypes.Normalized] `tfsdk:"validations"`
	state       attr.ValueState
}

var _ basetypes.ObjectValuable = ContentTypeFieldItemsValue{}

func NewContentTypeFieldItemsValueUnknown() ContentTypeFieldItemsValue {
	return ContentTypeFieldItemsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewContentTypeFieldItemsValueNull() ContentTypeFieldItemsValue {
	return ContentTypeFieldItemsValue{
		state: attr.ValueStateNull,
	}
}

func NewContentTypeFieldItemsValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (ContentTypeFieldItemsValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeFieldItemsValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

//nolint:ireturn
func (v ContentTypeFieldItemsValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ContentTypeFieldItemsType{ObjectType: v.ObjectType(ctx)}
}

//nolint:ireturn
func (v ContentTypeFieldItemsValue) Type(ctx context.Context) attr.Type {
	return ContentTypeFieldItemsType{ObjectType: v.ObjectType(ctx)}
}

func (v ContentTypeFieldItemsValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v ContentTypeFieldItemsValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v ContentTypeFieldItemsValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[ContentTypeFieldItemsValue](v, o)
}

func (v ContentTypeFieldItemsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ContentTypeFieldItemsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ContentTypeFieldItemsValue) String() string {
	return "ContentTypeFieldItemsValue"
}

func (v ContentTypeFieldItemsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v ContentTypeFieldItemsValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
