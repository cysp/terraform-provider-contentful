package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type ContentTypeMetadataAnnotationAssignmentValue struct {
	ID           types.String         `tfsdk:"id"`
	DefaultValue jsontypes.Normalized `tfsdk:"parameters"`
	state        attr.ValueState
}

var _ basetypes.ObjectValuable = ContentTypeMetadataAnnotationAssignmentValue{}

func NewContentTypeMetadataAnnotationAssignmentValueUnknown() ContentTypeMetadataAnnotationAssignmentValue {
	return ContentTypeMetadataAnnotationAssignmentValue{
		state: attr.ValueStateUnknown,
	}
}

func NewContentTypeMetadataAnnotationAssignmentValueNull() ContentTypeMetadataAnnotationAssignmentValue {
	return ContentTypeMetadataAnnotationAssignmentValue{
		state: attr.ValueStateNull,
	}
}

func NewContentTypeMetadataAnnotationAssignmentValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (ContentTypeMetadataAnnotationAssignmentValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeMetadataAnnotationAssignmentValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func (v ContentTypeMetadataAnnotationAssignmentValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Required: true,
		},
		"parameters": schema.StringAttribute{
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
			Computed:   true,
		},
	}
}

//nolint:ireturn
func (v ContentTypeMetadataAnnotationAssignmentValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ContentTypeMetadataAnnotationAssignmentType{ObjectType: v.ObjectType(ctx)}
}

//nolint:ireturn
func (v ContentTypeMetadataAnnotationAssignmentValue) Type(ctx context.Context) attr.Type {
	return ContentTypeMetadataAnnotationAssignmentType{ObjectType: v.ObjectType(ctx)}
}

func (v ContentTypeMetadataAnnotationAssignmentValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v ContentTypeMetadataAnnotationAssignmentValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v ContentTypeMetadataAnnotationAssignmentValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[ContentTypeMetadataAnnotationAssignmentValue](v, o)
}

func (v ContentTypeMetadataAnnotationAssignmentValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ContentTypeMetadataAnnotationAssignmentValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ContentTypeMetadataAnnotationAssignmentValue) String() string {
	return "ContentTypeMetadataAnnotationAssignmentValue"
}

func (v ContentTypeMetadataAnnotationAssignmentValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v ContentTypeMetadataAnnotationAssignmentValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
