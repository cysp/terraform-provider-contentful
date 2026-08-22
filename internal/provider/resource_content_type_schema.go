package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
		Description: "Manages a Contentful Content Type in the activated state. External deactivation is activation drift. If an observed newer unpublished draft otherwise matches modeled configuration, Terraform activates that exact draft without replacing it, preserving its unmodeled values. Modeled drift instead sends a complete draft update for provider-modeled fields, so unmodeled values are not guaranteed to survive. During Update, if activation fails after that draft update, Terraform checkpoints the returned draft and an unchanged retry activates it without repeating the draft update. During Create, the provider also checkpoints the draft before returning an activation error, but Terraform Core taints any resource whose Create returns an error; the next normal plan therefore replaces the resource and repeats the draft PUT. Every activation uses its exact observed or returned version and fails on concurrent change. After an ambiguous activation outcome, refresh before retrying. Resources created before published_version was tracked must refresh normally after upgrading; with -refresh=false, missing publication state conservatively may schedule activation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "The ID of the space this content type belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "The ID of the environment this content type belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content_type_id": schema.StringAttribute{
				Description: "The unique identifier for this content type.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the content type.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the content type.",
				Required:    true,
			},
			"display_field": schema.StringAttribute{
				Description: "Field ID to use as the display field for entries of this content type.",
				Required:    true,
			},
			"published_version": schema.Int64Attribute{
				Description: "The Contentful version most recently activated, or null when the content type is not activated.",
				Computed:    true,
			},
			"fields": schema.ListNestedAttribute{
				Description: "List of fields that belong to this content type.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: ContentTypeFieldValue{}.SchemaAttributes(ctx),
					CustomType: NewTypedObjectNull[ContentTypeFieldValue]().CustomType(ctx),
				},
				CustomType: NewTypedListUnknown[TypedObject[ContentTypeFieldValue]]().CustomType(ctx),
				Required:   true,
				Validators: []validator.List{
					listvalidator.NoNullValues(),
				},
			},
			"metadata": schema.SingleNestedAttribute{
				Attributes:  ContentTypeMetadataValue{}.SchemaAttributes(ctx),
				CustomType:  NewTypedObjectNull[ContentTypeMetadataValue]().CustomType(ctx),
				Description: `Metadata for the content type. Omitting metadata removes annotations but preserves taxonomy items. To remove taxonomy items, configure taxonomy as an empty list`,
				Optional:    true,
				Computed:    true,
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"source": schema.StringAttribute{
			Description: "The source for allowed Contentful entry resources.",
			Required:    true,
		},
		"content_types": schema.ListAttribute{
			Description: "List of content type IDs that are allowed to be linked.",
			ElementType: types.StringType,
			CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
			Required:    true,
			Validators: []validator.List{
				listvalidator.NoNullValues(),
			},
		},
	}
}

func (v ContentTypeFieldAllowedResourceItemExternalValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"type": schema.StringAttribute{
			Description: "The type of external resource.",
			Required:    true,
		},
	}
}

//nolint:dupl
func (v ContentTypeFieldAllowedResourceItemValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"contentful_entry": schema.SingleNestedAttribute{
			Description: "Configuration for allowing Contentful entry resources.",
			Attributes:  ContentTypeFieldAllowedResourceItemContentfulEntryValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[ContentTypeFieldAllowedResourceItemContentfulEntryValue]().CustomType(ctx),
			Optional:    true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("contentful_entry"),
					path.MatchRelative().AtParent().AtName("external"),
				),
			},
		},
		"external": schema.SingleNestedAttribute{
			Description: "Configuration for allowing external resources.",
			Attributes:  ContentTypeFieldAllowedResourceItemExternalValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[ContentTypeFieldAllowedResourceItemExternalValue]().CustomType(ctx),
			Optional:    true,
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
			Description: "The type of items in the array.",
			Required:    true,
		},
		"link_type": schema.StringAttribute{
			Description: "For arrays of Links, specifies the type of resource being linked to.",
			Optional:    true,
		},
		"validations": schema.ListAttribute{
			ElementType: jsontypes.NormalizedType{},
			CustomType:  NewTypedListNull[jsontypes.Normalized]().CustomType(ctx),
			Optional:    true,
			Computed:    true,
			Default:     listdefault.StaticValue(types.ListValueMust(jsontypes.NormalizedType{}, []attr.Value{})),
			Validators: []validator.List{
				listvalidator.NoNullValues(),
			},
		},
	}
}

func (v ContentTypeFieldValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The unique ID of this field within the content type.",
			Required:    true,
		},
		"name": schema.StringAttribute{
			Description: "The human-readable name of the field.",
			Required:    true,
		},
		"type": schema.StringAttribute{
			Description: "The field's data type.",
			Required:    true,
		},
		"link_type": schema.StringAttribute{
			Description: "For Link or Array of Links fields, specifies the type of resource being linked to (e.g., Entry, Asset).",
			Optional:    true,
		},
		"items": schema.SingleNestedAttribute{
			Description: "For Array fields, defines the type of items in the array.",
			Attributes:  ContentTypeFieldItemsValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[ContentTypeFieldItemsValue]().CustomType(ctx),
			Optional:    true,
		},
		"default_value": schema.StringAttribute{
			Description: "Default value for the field in JSON format.",
			CustomType:  jsontypes.NormalizedType{},
			Optional:    true,
		},
		"localized": schema.BoolAttribute{
			Description: "Whether the field can have different values for different locales.",
			Required:    true,
		},
		"disabled": schema.BoolAttribute{
			Description: "Whether the field is disabled (not editable in the UI).",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
		"omitted": schema.BoolAttribute{
			Description: "Whether the field is omitted from API responses. Before removing a field while the content type is activated, set omitted to true and apply so Contentful activates that change, then remove the field in a later apply.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
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
			Validators: []validator.List{
				listvalidator.NoNullValues(),
			},
		},
		"allowed_resources": schema.ListNestedAttribute{
			Description: "For Resource Link fields, defines the allowed resource types that can be linked.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: ContentTypeFieldAllowedResourceItemValue{}.SchemaAttributes(ctx),
				CustomType: NewTypedObjectNull[ContentTypeFieldAllowedResourceItemValue]().CustomType(ctx),
			},
			CustomType: NewTypedListNull[TypedObject[ContentTypeFieldAllowedResourceItemValue]]().CustomType(ctx),
			Optional:   true,
			Validators: []validator.List{
				listvalidator.NoNullValues(),
			},
		},
	}
}

func (v ContentTypeMetadataTaxonomyItemConceptSchemeValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "ID of the taxonomy concept scheme.",
			Required:    true,
		},
		"required": schema.BoolAttribute{
			Description: "Whether this taxonomy concept scheme is required.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
	}
}

func (v ContentTypeMetadataTaxonomyItemConceptValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "ID of the taxonomy concept.",
			Required:    true,
		},
		"required": schema.BoolAttribute{
			Description: "Whether this taxonomy concept is required.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
	}
}

//nolint:dupl
func (v ContentTypeMetadataTaxonomyItemValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"taxonomy_concept": schema.SingleNestedAttribute{
			Description: "A specific taxonomy concept to associate with this content type.",
			Attributes:  ContentTypeMetadataTaxonomyItemConceptValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptValue]().CustomType(ctx),
			Optional:    true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("taxonomy_concept"),
					path.MatchRelative().AtParent().AtName("taxonomy_concept_scheme"),
				),
			},
		},
		"taxonomy_concept_scheme": schema.SingleNestedAttribute{
			Description: "A taxonomy concept scheme to associate with this content type.",
			Attributes:  ContentTypeMetadataTaxonomyItemConceptSchemeValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue]().CustomType(ctx),
			Optional:    true,
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
			Computed:    true,
			Validators: []validator.List{
				listvalidator.NoNullValues(),
				listvalidator.AtLeastOneOf(
					path.MatchRelative().AtParent().AtName("annotations"),
					path.MatchRelative().AtParent().AtName("taxonomy"),
				),
			},
		},
	}
}
