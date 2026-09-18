package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func AppActionResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful App Action in an existing App Definition. Function deployment and action invocation are handled separately.\n\n" +
			"Create, Update, and Delete are not automatically retried, including after rate limiting. " +
			"If creation fails, check the App Definition for an action created without saved Terraform state and import it before applying again to avoid duplicates. " +
			"After a failed update or delete, refresh state and review the plan before retrying.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Composite Terraform resource identifier in `organization_id/app_definition_id/app_action_id` form.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Description:   "ID of the organization that owns the app. Changing this value replaces the resource.",
				Required:      true,
				Validators:    []validator.String{discoveryIDValidator{}},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"app_definition_id": schema.StringAttribute{
				Description:   "ID of the App Definition. Changing this value replaces the resource.",
				Required:      true,
				Validators:    []validator.String{discoveryIDValidator{}},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"app_action_id": schema.StringAttribute{
				Description:   "System ID of the App Action, allocated by Contentful.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Name of the App Action.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"category": schema.StringAttribute{
				Description: "Action category. Supported categories are `Custom`, `Entries.v1.0`, and `Notification.v1.0`. For `Custom`, configure exactly one of `parameters` and `parameters_schema`.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf("Custom", "Entries.v1.0", "Notification.v1.0")},
			},
			"type": schema.StringAttribute{
				Description: "Action type: `endpoint` or `function-invocation`.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf("endpoint", "function-invocation")},
			},
			"description": schema.StringAttribute{
				Description: "Description of the App Action. Omission removes an existing description on update.",
				Optional:    true,
			},
			"url": schema.StringAttribute{
				Description: "HTTPS endpoint URL. Required when `type` is `endpoint`; omit it for `function-invocation`.",
				Optional:    true,
			},
			"function_id": schema.StringAttribute{
				Description: "ID of an `appaction.call` Function. Required when `type` is `function-invocation`; omit it for `endpoint`.",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"parameters": schema.StringAttribute{
				Description: "Legacy parameter definitions for `Custom` actions, encoded as a JSON array. Each definition requires `id`, `name`, and `type` (`Boolean`, `Symbol`, `Number`, or `Enum`). `[]` defines no arguments. Built-in parameter definitions are available through the App Action data sources.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators:  []validator.String{appActionJSONValidator{array: true}},
			},
			"parameters_schema": schema.StringAttribute{
				Description: "Input JSON Schema (draft 4) object encoded with `jsonencode(...)`. Omission removes the schema on update.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators:  []validator.String{appActionJSONValidator{}},
			},
			"result_schema": schema.StringAttribute{
				Description: "JSON Schema (draft 4) object for successful action results, encoded with `jsonencode(...)`. Omission removes the schema on update.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators:  []validator.String{appActionJSONValidator{}},
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}
