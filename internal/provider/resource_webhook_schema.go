package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
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
			"version": schema.Int64Attribute{
				Description: "Contentful version of this webhook definition. A possible update makes the planned version unknown, including a change to a write-only credential.",
				Computed:    true,
			},
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
				MarkdownDescription: "HTTP Basic authentication password; configure it together with http_basic_username. Contentful does not return this value, so Terraform preserves a previously managed value during refresh but cannot detect changes made outside Terraform; import leaves it null. See [Secrets and Terraform state](../guides/secrets-and-state) for credential and state-handling guidance.",
				Optional:            true,
				Sensitive:           true,
			},
			"http_basic_password_wo": schema.StringAttribute{
				MarkdownDescription: "Write-only alternative to `http_basic_password` for Terraform 1.11.1 or later. Configure it together with `http_basic_username`. An ephemeral input keeps the password out of saved plans. Use a randomly generated password: the stored verifier permits offline guesses. See [Secrets and Terraform state](../guides/secrets-and-state).",
				Optional:            true, Sensitive: true, WriteOnly: true,
			},
			"header_values_wo": schema.MapAttribute{
				MarkdownDescription: "Write-only secret header values for Terraform 1.11.1 or later. Each key must match a header with `secret = true` and without an ordinary `value`. Omit both value arguments for an existing secret header to retain its remote value. Use ephemeral inputs and randomly generated secrets. See [Secrets and Terraform state](../guides/secrets-and-state).",
				ElementType:         types.StringType, CustomType: TypedMap[types.String]{}.CustomType(ctx),
				Optional: true, Sensitive: true, WriteOnly: true,
				Validators: []validator.Map{mapvalidator.NoNullValues()},
			},
			"http_basic_username": schema.StringAttribute{
				Description: "HTTP Basic authentication username. Configure username and password together. Omitting both clears Basic authentication on an update unless ignore_changes retains previously managed credentials.",
				Optional:    true,
			},
			"headers":        WebhookHeadersSchema(ctx, true),
			"transformation": WebhookTransformationSchema(ctx, true),
			"timeouts":       timeouts.Attributes(ctx, crudTimeoutOptions()),
		},
	}
}
