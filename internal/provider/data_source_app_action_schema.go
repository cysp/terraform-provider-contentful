package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func appActionDataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"app_action_id": schema.StringAttribute{
			Description: "System ID of the App Action.",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "Name of the App Action.",
			Computed:    true,
		},
		"category": schema.StringAttribute{
			Description: "Category of the App Action.",
			Computed:    true,
		},
		"type": schema.StringAttribute{
			Description: "Action type, such as `endpoint` or `function-invocation`.",
			Computed:    true,
		},
		"description": schema.StringAttribute{
			Description: "Description of the App Action.",
			Computed:    true,
		},
		"url": schema.StringAttribute{
			Description: "HTTPS endpoint URL for an `endpoint` action; otherwise `null`.",
			Computed:    true,
		},
		"function_id": schema.StringAttribute{
			Description: "Linked Function ID for a `function-invocation` action; otherwise `null`.",
			Computed:    true,
		},
		"parameters": schema.StringAttribute{
			Description: "Legacy or built-in category parameter definitions encoded as a JSON array, or `null` when absent.",
			Computed:    true,
			CustomType:  jsontypes.NormalizedType{},
		},
		"parameters_schema": schema.StringAttribute{
			Description: "Input JSON Schema encoded as JSON, or `null` when absent.",
			Computed:    true,
			CustomType:  jsontypes.NormalizedType{},
		},
		"result_schema": schema.StringAttribute{
			Description: "Result JSON Schema encoded as JSON, or `null` when absent.",
			Computed:    true,
			CustomType:  jsontypes.NormalizedType{},
		},
	}
}

func AppActionDataSourceSchema(ctx context.Context) schema.Schema {
	attributes := appActionDataSourceAttributes()
	attributes["organization_id"] = schema.StringAttribute{
		Description: "ID of the organization.",
		Required:    true,
		Validators:  []validator.String{discoveryIDValidator{}},
	}
	attributes["app_definition_id"] = schema.StringAttribute{
		Description: "ID of the App Definition.",
		Required:    true,
		Validators:  []validator.String{discoveryIDValidator{}},
	}
	attributes["app_action_id"] = schema.StringAttribute{
		Description: "System ID of the App Action.",
		Required:    true,
		Validators:  []validator.String{discoveryIDValidator{}},
	}
	attributes["id"] = schema.StringAttribute{
		Description: "Composite Terraform identifier in `organization_id/app_definition_id/app_action_id` form.",
		Computed:    true,
	}
	attributes["timeouts"] = timeouts.Attributes(ctx)

	return schema.Schema{
		Description: "Reads an existing Contentful App Action, including its target and parameter definitions.",
		Attributes:  attributes,
	}
}

func AppActionsDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Retrieves all Contentful App Actions in an App Definition.",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				Description: "ID of the organization.",
				Required:    true,
				Validators:  []validator.String{discoveryIDValidator{}},
			},
			"app_definition_id": schema.StringAttribute{
				Description: "ID of the App Definition.",
				Required:    true,
				Validators:  []validator.String{discoveryIDValidator{}},
			},
			"id": schema.StringAttribute{
				Description: "Composite Terraform identifier in `organization_id/app_definition_id` form.",
				Computed:    true,
			},
			"app_actions": schema.ListNestedAttribute{
				Description: "App Actions in the App Definition, ordered lexicographically by `app_action_id`. An App Definition with no actions returns an empty list.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: appActionDataSourceAttributes(),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}
