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
		Description: "Manages the App Event Subscription for an organization and App Definition. Create and update upsert the complete subscription, including an existing subscription at that address. Manage each subscription with one resource. Installation determines the space and environment that trigger events; this resource configures routing and does not install the app or establish delivery readiness.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Composite Terraform resource identifier in `organization_id/app_definition_id` form; Contentful does not assign a subscription ID.",
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
				Description:   "ID of the existing App Definition. Changing this value replaces the resource.",
				Required:      true,
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"topics": schema.SetAttribute{
				Description: "Complete nonempty set of event topics, such as `Entry.publish`. Order is insignificant and duplicate configuration values collapse to one set member. Contentful validates its current topic vocabulary; webhook wildcard syntax is not implied.",
				Required:    true,
				ElementType: types.StringType,
				Validators:  []validator.Set{setvalidator.SizeAtLeast(1), setvalidator.NoNullValues(), setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))},
			},
			"target_url": schema.StringAttribute{
				Description: "HTTPS endpoint for HTTP delivery. Omit when using a handler Function. Omission excludes the target from the complete applied subscription. Sensitive because URLs can contain credentials; stored in Terraform state.",
				Optional:    true,
				Sensitive:   true,
				Validators:  []validator.String{appEventSubscriptionTargetValidator{}},
			},
			"filter_function_id": schema.StringAttribute{
				Description: "ID of a deployed `appevent.filter` Function that decides whether an event proceeds. Omission excludes this role from the applied subscription. Function availability and compatibility are validated by Contentful.",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"transformation_function_id": schema.StringAttribute{
				Description: "ID of a deployed `appevent.transformation` Function that modifies the request before signing. Omission excludes this role from the applied subscription. Function availability and compatibility are validated by Contentful.",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"handler_function_id": schema.StringAttribute{
				Description: "ID of a deployed `appevent.handler` Function used as the event destination in place of an HTTP target. Omission excludes this role from the applied subscription. Function availability and compatibility are validated by Contentful.",
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
