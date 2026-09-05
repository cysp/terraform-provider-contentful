package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func WebhookResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful Webhook.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite Terraform resource identifier in space_id/webhook_id form.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space containing the webhook.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"webhook_id": schema.StringAttribute{
				Description: "System ID of the webhook.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"active": schema.BoolAttribute{
				Description: "Whether the webhook is active. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"name": schema.StringAttribute{
				Description: "Name of the webhook.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "Preconfigured HTTP endpoint that is called when content has changed.",
				Required:    true,
			},
			"topics": schema.ListAttribute{
				Description: "List of one or more event topics to which the webhook subscribes.",
				ElementType: types.StringType,
				CustomType:  TypedList[types.String]{}.CustomType(ctx),
				Required:    true,
				Validators: []validator.List{
					listvalidator.NoNullValues(),
					listvalidator.SizeAtLeast(1),
				},
			},
			"filters": WebhookFiltersSchema(ctx, true),
			"http_basic_password": schema.StringAttribute{
				Description: "HTTP Basic authentication password. Contentful does not return this value, so Terraform preserves a previously managed value during refresh but cannot detect changes made outside Terraform; import leaves it null. Terraform marks the value sensitive, which obscures CLI output but does not encrypt or omit plan or state data.",
				Optional:    true,
				Sensitive:   true,
			},
			"http_basic_username": schema.StringAttribute{
				Description: "HTTP Basic authentication username.",
				Optional:    true,
			},
			"headers":        WebhookHeadersSchema(ctx, true),
			"transformation": WebhookTransformationSchema(ctx, true),
			"timeouts":       timeouts.AttributesAll(ctx),
		},
	}
}
