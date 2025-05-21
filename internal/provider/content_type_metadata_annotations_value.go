package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type ContentTypeMetadataAnnotationsValue struct {
	ContentType      TypedList[ContentTypeMetadataAnnotationAssignmentValue] `tfsdk:"content_type"`
	ContentTypeField TypedMap[ContentTypeMetadataAnnotationAssignmentValue]  `tfsdk:"content_type_field"`
	state            attr.ValueState
}

var _ basetypes.ObjectValuable = ContentTypeMetadataAnnotationsValue{}

func NewContentTypeMetadataAnnotationsValueUnknown() ContentTypeMetadataAnnotationsValue {
	return ContentTypeMetadataAnnotationsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewContentTypeMetadataAnnotationsValueNull() ContentTypeMetadataAnnotationsValue {
	return ContentTypeMetadataAnnotationsValue{
		state: attr.ValueStateNull,
	}
}

func NewContentTypeMetadataAnnotationsValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (ContentTypeMetadataAnnotationsValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeMetadataAnnotationsValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func (v ContentTypeMetadataAnnotationsValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"content_type": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: ContentTypeMetadataAnnotationAssignmentValue{}.SchemaAttributes(ctx),
				CustomType: ContentTypeMetadataAnnotationAssignmentValue{}.CustomType(ctx),
			},
			CustomType: NewTypedListNull[ContentTypeMetadataAnnotationAssignmentValue](ctx).CustomType(ctx),
			Optional:   true,
		},
		"content_type_field": schema.MapNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: ContentTypeMetadataAnnotationAssignmentValue{}.SchemaAttributes(ctx),
				CustomType: ContentTypeMetadataAnnotationAssignmentValue{}.CustomType(ctx),
			},
			CustomType: NewTypedMapNull[ContentTypeMetadataAnnotationAssignmentValue](ctx).CustomType(ctx),
			Optional:   true,
		},
	}
}

//nolint:ireturn
func (v ContentTypeMetadataAnnotationsValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ContentTypeMetadataAnnotationsType{ObjectType: v.ObjectType(ctx)}
}

//nolint:ireturn
func (v ContentTypeMetadataAnnotationsValue) Type(ctx context.Context) attr.Type {
	return ContentTypeMetadataAnnotationsType{ObjectType: v.ObjectType(ctx)}
}

func (v ContentTypeMetadataAnnotationsValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v ContentTypeMetadataAnnotationsValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v ContentTypeMetadataAnnotationsValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[ContentTypeMetadataAnnotationsValue](v, o)
}

func (v ContentTypeMetadataAnnotationsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ContentTypeMetadataAnnotationsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ContentTypeMetadataAnnotationsValue) String() string {
	return "ContentTypeMetadataAnnotationsValue"
}

func (v ContentTypeMetadataAnnotationsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v ContentTypeMetadataAnnotationsValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
