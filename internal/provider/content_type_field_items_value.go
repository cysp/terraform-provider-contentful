package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type ContentTypeFieldItemsValue struct {
	ItemsType   basetypes.StringValue `tfsdk:"type"`
	LinkType    basetypes.StringValue `tfsdk:"link_type"`
	Validations basetypes.ListValue   `tfsdk:"validations"`
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

func NewContentTypeFieldItemsValueKnownFromAttributesMust(ctx context.Context, attributes map[string]attr.Value) ContentTypeFieldItemsValue {
	value, diags := NewContentTypeFieldItemsValueKnownFromAttributes(ctx, attributes)
	if diags.HasError() {
		panic(diags)
	}

	return value
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

func (v ContentTypeFieldItemsValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"type": schema.StringAttribute{
			Required: true,
		},
		"link_type": schema.StringAttribute{
			Optional: true,
		},
		"validations": schema.ListAttribute{
			ElementType: jsontypes.NormalizedType{},
			Optional:    true,
			Computed:    true,
			Default:     listdefault.StaticValue(NewEmptyListMust(jsontypes.NormalizedType{})),
		},
	}
}

//nolint:ireturn
func (v ContentTypeFieldItemsValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ContentTypeFieldItemsType{ObjectType: v.ObjectType(ctx)}
}

//nolint:ireturn
func (v ContentTypeFieldItemsValue) Type(ctx context.Context) attr.Type {
	return ContentTypeFieldItemsType{ObjectType: v.ObjectType(ctx)}
}

func (v ContentTypeFieldItemsValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v ContentTypeFieldItemsValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v ContentTypeFieldItemsValue) Equal(o attr.Value) bool {
	other, ok := o.(ContentTypeFieldItemsValue)
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state == attr.ValueStateKnown {
		return compareTFSDKAttributesEqual(v, other)
	}

	return true
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
	return ReflectToTerraformValue(ctx, v, v.state)
}

func (v ContentTypeFieldItemsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return ReflectToObjectValue(ctx, v)
}
