package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func EntryResourceSchema(ctx context.Context) schema.Schema {
	defaultMetadataObjectValue, _ := NewTypedObject[EntryMetadataValue](EntryMetadataValue{
		Tags: NewTypedListFromStringSlice([]string{}),
	}).ToObjectValue(ctx)

	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					UseStateForUnknown(),
				},
			},
			"entry_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					UseStateForUnknown(),
				},
			},
			"content_type_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					UseStateForUnknown(),
				},
			},
			"fields": schema.MapAttribute{
				ElementType: jsontypes.NormalizedType{},
				CustomType:  NewTypedMapNull[jsontypes.Normalized]().CustomType(ctx),
				Required:    true,
			},
			"metadata": schema.SingleNestedAttribute{
				Attributes:  EntryMetadataValue{}.SchemaAttributes(ctx),
				CustomType:  NewTypedObjectNull[EntryMetadataValue]().CustomType(ctx),
				Description: "Metadata for the entry. Once set, metadata properties may not be removed, but the list of tags may be reduced to the empty list",
				Optional:    true,
				Computed:    true,
				Default:     objectdefault.StaticValue(defaultMetadataObjectValue),
				PlanModifiers: []planmodifier.Object{
					UseStateForUnknown(),
				},
			},
		},
	}
}

func (v EntryMetadataValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	defaultTagsListValue, _ := NewTypedListFromStringSlice([]string{}).ToListValue(ctx)

	return map[string]schema.Attribute{
		"tags": schema.ListAttribute{
			ElementType: types.StringType,
			CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
			Optional:    true,
			Computed:    true,
			Default:     listdefault.StaticValue(defaultTagsListValue),
			PlanModifiers: []planmodifier.List{
				UseStateForUnknown(),
			},
		},
	}
}
