package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func EntryResourceSchema(ctx context.Context) schema.Schema {
	return entryResourceSchema(ctx, 1, localizedEntryFieldsSchemaAttribute(ctx))
}

func EntryResourceSchemaV0(ctx context.Context) schema.Schema {
	return entryResourceSchema(ctx, 0, rawEntryFieldsSchemaAttribute(ctx))
}

func entryResourceSchema(ctx context.Context, version int64, fieldsAttribute schema.Attribute) schema.Schema {
	defaultMetadataObjectValue, _ := NewTypedObject[EntryMetadataValue](EntryMetadataValue{
		Concepts: NewTypedListFromStringSlice([]string{}),
		Tags:     NewTypedListFromStringSlice([]string{}),
	}).ToObjectValue(ctx)

	return schema.Schema{
		Version:     version,
		Description: "Manages a Contentful Entry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space containing the entry.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "ID of the environment containing the entry.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"entry_id": schema.StringAttribute{
				Description: "ID of the entry. When `entry_id` is configured, Terraform creates a new Entry with the specified ID; an existing Entry with the same ID causes an error and is not adopted.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"content_type_id": schema.StringAttribute{
				Description: "ID of the content type for this entry.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"published_version": schema.Int64Attribute{
				Description: "Contentful `sys.publishedVersion` for the Entry. Null when Contentful reports the Entry as unpublished.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"fields": fieldsAttribute,
			"metadata": schema.SingleNestedAttribute{
				Attributes:  EntryMetadataValue{}.SchemaAttributes(ctx),
				CustomType:  NewTypedObjectNull[EntryMetadataValue]().CustomType(ctx),
				Description: "Entry metadata, including assigned tags and taxonomy concepts.",
				Optional:    true,
				Computed:    true,
				Default:     objectdefault.StaticValue(defaultMetadataObjectValue),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func localizedEntryFieldsSchemaAttribute(ctx context.Context) schema.MapAttribute {
	return schema.MapAttribute{
		Description: "Fields that are custom defined by a user through the definition of content types, keyed by field ID and locale.",
		ElementType: NewTypedMapNull[jsontypes.Normalized]().CustomType(ctx),
		CustomType:  NewTypedMapNull[TypedMap[jsontypes.Normalized]]().CustomType(ctx),
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Map{
			mapplanmodifier.UseStateForUnknown(),
		},
	}
}

func rawEntryFieldsSchemaAttribute(ctx context.Context) schema.MapAttribute {
	return schema.MapAttribute{
		Description: "Fields that are custom defined by a user through the definition of content types. Fields object always includes locale.",
		ElementType: jsontypes.NormalizedType{},
		CustomType:  NewTypedMapNull[jsontypes.Normalized]().CustomType(ctx),
		Required:    true,
	}
}

func (v EntryMetadataValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	defaultConceptsListValue, _ := NewTypedListFromStringSlice([]string{}).ToListValue(ctx)
	defaultTagsListValue, _ := NewTypedListFromStringSlice([]string{}).ToListValue(ctx)

	return map[string]schema.Attribute{
		"concepts": schema.ListAttribute{
			Description: "IDs of Contentful taxonomy concepts attached to the entry. Comparison ignores ordering but preserves duplicate occurrences. Reordering alone may update Terraform state but sends no Contentful Entry PUT or Publish request.",
			ElementType: types.StringType,
			CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
			Optional:    true,
			Computed:    true,
			Default:     listdefault.StaticValue(defaultConceptsListValue),
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.List{
				listvalidator.NoNullValues(),
			},
		},
		"tags": schema.ListAttribute{
			Description: "IDs of Contentful tags attached to the entry. Comparison ignores ordering but preserves duplicate occurrences. Reordering alone may update Terraform state but sends no Contentful Entry PUT or Publish request.",
			ElementType: types.StringType,
			CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
			Optional:    true,
			Computed:    true,
			Default:     listdefault.StaticValue(defaultTagsListValue),
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.List{
				listvalidator.NoNullValues(),
			},
		},
	}
}
