package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
					CustomType: ContentTypeFieldValue{}.CustomType(ctx),
				},
				CustomType: NewTypedListUnknown[ContentTypeFieldValue](ctx).CustomType(ctx),
				Required:   true,
			},
			"metadata": schema.SingleNestedAttribute{
				Attributes:  ContentTypeMetadataValue{}.SchemaAttributes(ctx),
				CustomType:  ContentTypeMetadataValue{}.CustomType(ctx),
				Description: `Metadata for the content type. Once set, metadata properties may not be removed, but the list of taxonomy items may be reduced to the empty list`,
				Optional:    true,
			},
		},
	}
}
