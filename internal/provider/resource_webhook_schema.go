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
		Description: "Manages a Contentful Webhook that sends HTTP requests when events match its topics and filters. Configure the receiving endpoint separately.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite Terraform resource identifier in `space_id/webhook_id` form.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space containing the webhook. Changing this value replaces the resource.",
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
			"version": schema.Int64Attribute{
				Description: "The Contentful `sys.version` last observed for this webhook. A possible write-only rotation makes the planned version unknown.",
				Computed:    true,
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
				Description: "HTTP endpoint that receives events matching the webhook topics and filters.",
				Required:    true,
			},
			"topics": schema.ListAttribute{
				Description: "One or more Contentful event topics, such as `Entry.publish`, `Entry.unpublish`, or `Asset.save`.",
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
				MarkdownDescription: "HTTP Basic authentication password. Requires a username in the effective plan. Import leaves an omitted password null. See [Secrets and Terraform state](../guides/secrets-and-state#webhook-credentials-and-secret-headers) for delegation, refresh and recovery.",
				Optional:            true,
				Sensitive:           true,
			},
			"http_basic_password_wo": schema.StringAttribute{
				MarkdownDescription: "Write-only alternative to `http_basic_password` for Terraform 1.11.1 or later. Requires a username in the effective plan to set or rotate the password. An ephemeral input keeps the password out of saved plans. See [Secrets and Terraform state](../guides/secrets-and-state#basic-authentication).",
				Optional:            true,
				Sensitive:           true,
				WriteOnly:           true,
			},
			"http_basic_username": schema.StringAttribute{
				Description: "HTTP Basic authentication username. Supply a password when setting a new username. Retain the same username with both password arguments omitted to preserve an existing password. Removing all Basic credentials from the effective plan clears authentication.",
				Optional:    true,
			},
			"headers":        WebhookHeadersSchema(ctx),
			"transformation": WebhookTransformationSchema(ctx, true),
			"timeouts":       timeouts.Attributes(ctx, crudTimeoutOptions()),
		},
	}
}
