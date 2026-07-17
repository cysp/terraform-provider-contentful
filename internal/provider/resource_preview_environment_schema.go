package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func PreviewEnvironmentResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful content preview platform. This space-level resource is not a Contentful environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space containing the content preview platform.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"preview_environment_id": schema.StringAttribute{
				Description: "System ID of the content preview platform. When omitted, Contentful generates an ID. When forcing replacement of a configured platform with a caller-selected ID, choose a different ID because Contentful may prevent immediate recreation under the previous value.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the content preview platform.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the content preview platform.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"content_type_configurations": schema.MapNestedAttribute{
				Description: "Active preview URL configurations keyed by content type ID. Removing a key disables that configuration in Contentful.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: PreviewEnvironmentContentTypeConfigurationValue{}.SchemaAttributes(ctx),
					CustomType: NewTypedObjectNull[PreviewEnvironmentContentTypeConfigurationValue]().CustomType(ctx),
				},
				CustomType: TypedMap[TypedObject[PreviewEnvironmentContentTypeConfigurationValue]]{}.CustomType(ctx),
				Required:   true,
				Validators: []validator.Map{
					mapvalidator.KeysAre(stringvalidator.LengthAtLeast(1)),
				},
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (PreviewEnvironmentContentTypeConfigurationValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"url": schema.StringAttribute{
			Description: "Preview URL template. Do not include access tokens.",
			Required:    true,
		},
	}
}
