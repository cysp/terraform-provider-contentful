package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ContentTypeResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"space_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content_type_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Required: true,
			},
			"display_field": schema.StringAttribute{
				Required: true,
			},
			"fields": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: ContentTypeFieldValue{}.SchemaAttributes(ctx),
					CustomType: NewTypedObjectNull[ContentTypeFieldValue]().CustomType(ctx),
				},
				CustomType: NewTypedListUnknown[TypedObject[ContentTypeFieldValue]]().CustomType(ctx),
				Required:   true,
			},
			"metadata": schema.SingleNestedAttribute{
				Attributes:  ContentTypeMetadataValue{}.SchemaAttributes(ctx),
				CustomType:  NewTypedObjectNull[ContentTypeMetadataValue]().CustomType(ctx),
				Description: `Metadata for the content type. Once set, metadata properties may not be removed, but the list of taxonomy items may be reduced to the empty list`,
				Optional:    true,
			},
		},
	}
}

func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"source": schema.StringAttribute{
			Required: true,
		},
		"content_types": schema.ListAttribute{
			ElementType: types.StringType,
			CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
			Required:    true,
		},
	}
}

func (v ContentTypeFieldAllowedResourceItemExternalValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"type": schema.StringAttribute{
			Required: true,
		},
	}
}

func (v ContentTypeFieldAllowedResourceItemValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"contentful_entry": schema.SingleNestedAttribute{
			Attributes: ContentTypeFieldAllowedResourceItemContentfulEntryValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[ContentTypeFieldAllowedResourceItemContentfulEntryValue]().CustomType(ctx),
			Optional:   true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("contentful_entry"),
					path.MatchRelative().AtParent().AtName("external"),
				),
			},
		},
		"external": schema.SingleNestedAttribute{
			Attributes: ContentTypeFieldAllowedResourceItemExternalValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[ContentTypeFieldAllowedResourceItemExternalValue]().CustomType(ctx),
			Optional:   true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("contentful_entry"),
					path.MatchRelative().AtParent().AtName("external"),
				),
			},
		},
	}
}

func (v ContentTypeFieldItemsValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"type": schema.StringAttribute{
			Required: true,
		},
		"link_type": schema.StringAttribute{
			Optional: true,
		},
		"validations": schema.ListAttribute{
			ElementType: jsontypes.NormalizedType{},
			CustomType:  NewTypedListNull[jsontypes.Normalized]().CustomType(ctx),
			Optional:    true,
			Computed:    true,
			Default:     listdefault.StaticValue(types.ListValueMust(jsontypes.NormalizedType{}, []attr.Value{})),
		},
	}
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
			CustomType: NewTypedObjectNull[ContentTypeFieldItemsValue]().CustomType(ctx),
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
			CustomType:  NewTypedListNull[jsontypes.Normalized]().CustomType(ctx),
			Optional:    true,
			Computed:    true,
			Default:     listdefault.StaticValue(types.ListValueMust(jsontypes.NormalizedType{}, []attr.Value{})),
		},
		"allowed_resources": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: ContentTypeFieldAllowedResourceItemValue{}.SchemaAttributes(ctx),
				CustomType: NewTypedObjectNull[ContentTypeFieldAllowedResourceItemValue]().CustomType(ctx),
			},
			CustomType: NewTypedListNull[TypedObject[ContentTypeFieldAllowedResourceItemValue]]().CustomType(ctx),
			Optional:   true,
		},
	}
}

func (v ContentTypeMetadataTaxonomyItemConceptSchemeValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Required: true,
		},
		"required": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
	}
}

func (v ContentTypeMetadataTaxonomyItemConceptValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Required: true,
		},
		"required": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
	}
}

func (v ContentTypeMetadataTaxonomyItemValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"taxonomy_concept": schema.SingleNestedAttribute{
			Attributes: ContentTypeMetadataTaxonomyItemConceptValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptValue]().CustomType(ctx),
			Optional:   true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("taxonomy_concept"),
					path.MatchRelative().AtParent().AtName("taxonomy_concept_scheme"),
				),
			},
		},
		"taxonomy_concept_scheme": schema.SingleNestedAttribute{
			Attributes: ContentTypeMetadataTaxonomyItemConceptSchemeValue{}.SchemaAttributes(ctx),
			CustomType: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue]().CustomType(ctx),
			Optional:   true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("taxonomy_concept"),
					path.MatchRelative().AtParent().AtName("taxonomy_concept_scheme"),
				),
			},
		},
	}
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
				CustomType: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemValue]().CustomType(ctx),
			},
			CustomType:  NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]]().CustomType(ctx),
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
