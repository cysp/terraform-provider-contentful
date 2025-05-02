package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type ContentTypeFieldValue struct {
	ID               types.String                                        `tfsdk:"id"`
	Name             types.String                                        `tfsdk:"name"`
	FieldType        types.String                                        `tfsdk:"type"`
	LinkType         types.String                                        `tfsdk:"link_type"`
	Disabled         types.Bool                                          `tfsdk:"disabled"`
	Omitted          types.Bool                                          `tfsdk:"omitted"`
	Required         types.Bool                                          `tfsdk:"required"`
	DefaultValue     jsontypes.Normalized                                `tfsdk:"default_value"`
	Items            ContentTypeFieldItemsValue                          `tfsdk:"items"`
	Localized        types.Bool                                          `tfsdk:"localized"`
	Validations      TypedList[jsontypes.Normalized]                     `tfsdk:"validations"`
	AllowedResources TypedList[ContentTypeFieldAllowedResourceItemValue] `tfsdk:"allowed_resources"`
	state            attr.ValueState
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

func NewContentTypeFieldValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (ContentTypeFieldValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeFieldValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
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
			CustomType:  NewTypedListNull[jsontypes.Normalized](ctx).CustomType(ctx),
			Optional:    true,
			Computed:    true,
			Default:     listdefault.StaticValue(types.ListValueMust(jsontypes.NormalizedType{}, []attr.Value{})),
		},
		"allowed_resources": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: NewContentTypeFieldAllowedResourceItemValueNull().SchemaAttributes(ctx),
				CustomType: NewContentTypeFieldAllowedResourceItemValueNull().CustomType(ctx),
			},
			CustomType: NewTypedListNull[ContentTypeFieldAllowedResourceItemValue](ctx).CustomType(ctx),
			Optional:   true,
		},
	}
}

//nolint:ireturn
func (v ContentTypeFieldValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ContentTypeFieldType{ObjectType: v.ObjectType(ctx)}
}

//nolint:ireturn
func (v ContentTypeFieldValue) Type(ctx context.Context) attr.Type {
	return ContentTypeFieldType{ObjectType: v.ObjectType(ctx)}
}

func (v ContentTypeFieldValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v ContentTypeFieldValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v ContentTypeFieldValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[ContentTypeFieldValue](v, o)
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
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v ContentTypeFieldValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
