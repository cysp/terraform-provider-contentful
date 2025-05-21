package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type ContentTypeMetadataValue struct {
	Annotations jsontypes.Normalized                            `tfsdk:"annotations"`
	Taxonomy    TypedList[ContentTypeMetadataTaxonomyItemValue] `tfsdk:"taxonomy"`
	state       attr.ValueState
}

var _ basetypes.ObjectValuable = ContentTypeMetadataValue{}

func NewContentTypeMetadataValueUnknown() ContentTypeMetadataValue {
	return ContentTypeMetadataValue{
		state: attr.ValueStateUnknown,
	}
}

func NewContentTypeMetadataValueNull() ContentTypeMetadataValue {
	return ContentTypeMetadataValue{
		state: attr.ValueStateNull,
	}
}

func NewContentTypeMetadataValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (ContentTypeMetadataValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeMetadataValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func (v ContentTypeMetadataValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"annotations": schema.StringAttribute{
			CustomType:  jsontypes.NormalizedType{},
			Description: "Annotations for this content type, represented as a JSON object fragment.",
			Optional:    true,
			Validators: []validator.String{
				stringvalidator.AtLeastOneOf(
					path.MatchRelative().AtParent().AtName("annotations"),
					path.MatchRelative().AtParent().AtName("taxonomy"),
				),
			},
		},
		"taxonomy": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: ContentTypeMetadataTaxonomyItemValue{}.SchemaAttributes(ctx),
				CustomType: ContentTypeMetadataTaxonomyItemValue{}.CustomType(ctx),
			},
			CustomType:  NewTypedListNull[ContentTypeMetadataTaxonomyItemValue](ctx).CustomType(ctx),
			Description: "List of taxonomy items for this content type. Each item represents a taxonomy term that may be associated with the content type.",
			Optional:    true,
			Validators: []validator.List{
				listvalidator.AtLeastOneOf(
					path.MatchRelative().AtParent().AtName("annotations"),
					path.MatchRelative().AtParent().AtName("taxonomy"),
				),
			},
		},
	}
}

//nolint:ireturn
func (v ContentTypeMetadataValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ContentTypeMetadataType{ObjectType: v.ObjectType(ctx)}
}

//nolint:ireturn
func (v ContentTypeMetadataValue) Type(ctx context.Context) attr.Type {
	return ContentTypeMetadataType{ObjectType: v.ObjectType(ctx)}
}

func (v ContentTypeMetadataValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v ContentTypeMetadataValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v ContentTypeMetadataValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[ContentTypeMetadataValue](v, o)
}

func (v ContentTypeMetadataValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ContentTypeMetadataValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ContentTypeMetadataValue) String() string {
	return "ContentTypeMetadataValue"
}

func (v ContentTypeMetadataValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v ContentTypeMetadataValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
