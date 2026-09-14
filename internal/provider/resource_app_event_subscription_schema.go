package provider

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func AppEventSubscriptionResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful App Event Subscription.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Composite Terraform resource identifier in `organization_id/app_definition_id` form.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Description:   "ID of the organization that owns the app. Changing this value replaces the resource.",
				Required:      true,
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"app_definition_id": schema.StringAttribute{
				Description:   "ID of the App Definition. Changing this value replaces the resource.",
				Required:      true,
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"topics": schema.SetAttribute{
				Description: "Nonempty set of event topics, such as `Entry.publish`. Updates replace the complete set. Contentful validates the supported topics.",
				Required:    true,
				ElementType: types.StringType,
				Validators:  []validator.Set{setvalidator.SizeAtLeast(1), setvalidator.NoNullValues(), setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))},
			},
			"target_url": schema.StringAttribute{
				Description: "HTTPS URL for event delivery. Omit when using `handler_function_id`. Marked sensitive and stored in Terraform state.",
				Optional:    true,
				Sensitive:   true,
				Validators:  []validator.String{appEventSubscriptionTargetValidator{}},
			},
			"filter_function_id": schema.StringAttribute{
				Description: "ID of an `appevent.filter` Function that decides whether an event proceeds. Omission excludes this role from the subscription request.",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"transformation_function_id": schema.StringAttribute{
				Description: "ID of an `appevent.transformation` Function that modifies the request before signing. Omission excludes this role from the subscription request.",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"handler_function_id": schema.StringAttribute{
				Description: "ID of an `appevent.handler` Function used as the event destination instead of `target_url`. Omission excludes this role from the subscription request.",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

var appEventSubscriptionTargetPattern = regexp.MustCompile(`^https://.+`)

type appEventSubscriptionTargetValidator struct{}

func (appEventSubscriptionTargetValidator) Description(context.Context) string {
	return "The target must be an HTTPS URL."
}

func (v appEventSubscriptionTargetValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v appEventSubscriptionTargetValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if !appEventSubscriptionTargetPattern.MatchString(req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid app event target URL", v.Description(ctx))
	}
}
