package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type ContentTypeFieldValue struct {
	ID           basetypes.StringValue `tfsdk:"id"`
	Name         basetypes.StringValue `tfsdk:"name"`
	FieldType    basetypes.StringValue `tfsdk:"type"`
	LinkType     basetypes.StringValue `tfsdk:"link_type"`
	Disabled     basetypes.BoolValue   `tfsdk:"disabled"`
	Omitted      basetypes.BoolValue   `tfsdk:"omitted"`
	Required     basetypes.BoolValue   `tfsdk:"required"`
	DefaultValue jsontypes.Normalized  `tfsdk:"default_value"`
	Items        basetypes.ObjectValue `tfsdk:"items"`
	Localized    basetypes.BoolValue   `tfsdk:"localized"`
	Validations  basetypes.ListValue   `tfsdk:"validations"`
	state        attr.ValueState
}

var _ basetypes.ObjectValuable = ContentTypeFieldValue{}

func NewContentTypeFieldValueUnknown() ContentTypeFieldValue {
	return ContentTypeFieldValue{
		state: attr.ValueStateUnknown,
	}
}

func NewContentTypeFieldValueNull() ContentTypeFieldValue {
	return ContentTypeFieldValue{
		state: attr.ValueStateNull,
	}
}

func NewContentTypeFieldValueKnownFromAttributesMust(ctx context.Context, attributes map[string]attr.Value) ContentTypeFieldValue {
	value, diags := NewContentTypeFieldValueKnownFromAttributes(ctx, attributes)
	if diags.HasError() {
		panic(diags)
	}

	return value
}

func NewContentTypeFieldValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (ContentTypeFieldValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeFieldValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := setTFSDKAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func (v ContentTypeFieldValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Required: true,
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"type": schema.StringAttribute{
			Required: true,
		},
		"link_type": schema.StringAttribute{
			Optional: true,
		},
		"items": schema.SingleNestedAttribute{
			Attributes: ContentTypeFieldItemsValue{}.SchemaAttributes(ctx),
			CustomType: ContentTypeFieldItemsValue{}.CustomType(ctx),
			Optional:   true,
		},
		"default_value": schema.StringAttribute{
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
			Computed:   true,
		},
		"localized": schema.BoolAttribute{
			Required: true,
		},
		"disabled": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
		"omitted": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
		"required": schema.BoolAttribute{
			Required: true,
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
func (v ContentTypeFieldValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ContentTypeFieldType{
		v.ObjectType(ctx),
	}
}

//nolint:ireturn
func (v ContentTypeFieldValue) Type(ctx context.Context) attr.Type {
	return ContentTypeFieldType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v ContentTypeFieldValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
}

func (v ContentTypeFieldValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"id":            basetypes.StringType{},
		"name":          basetypes.StringType{},
		"type":          basetypes.StringType{},
		"link_type":     basetypes.StringType{},
		"items":         ContentTypeFieldItemsValue{}.ObjectType(ctx),
		"default_value": jsontypes.NormalizedType{},
		"localized":     basetypes.BoolType{},
		"disabled":      basetypes.BoolType{},
		"omitted":       basetypes.BoolType{},
		"required":      basetypes.BoolType{},
		"validations":   basetypes.ListType{ElemType: jsontypes.NormalizedType{}},
	}
}

func (v ContentTypeFieldValue) Equal(o attr.Value) bool {
	other, ok := o.(ContentTypeFieldValue)
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

func (v ContentTypeFieldValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ContentTypeFieldValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ContentTypeFieldValue) String() string {
	return "ContentTypeFieldValue"
}

func (v ContentTypeFieldValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return ReflectToTerraformValue(ctx, v, v.state)
}

func (v ContentTypeFieldValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return ReflectToObjectValue(ctx, v)
}
