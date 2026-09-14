package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func environmentDataSourceItemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"environment_id":         schema.StringAttribute{Description: "ID of the environment or environment alias.", Computed: true},
		"name":                   schema.StringAttribute{Description: "Name returned by Contentful, which can be the alias name when looking up an alias.", Computed: true},
		"status":                 schema.StringAttribute{Description: "Latest environment status returned by Contentful.", Computed: true},
		"aliased_environment_id": schema.StringAttribute{Description: "ID of the target environment when looking up an alias, or null when Contentful omits it.", Computed: true},
	}
}

func EnvironmentDataSourceSchema(ctx context.Context) schema.Schema {
	attributes := environmentDataSourceItemAttributes()
	attributes["space_id"] = schema.StringAttribute{Description: "ID of the space.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["environment_id"] = schema.StringAttribute{Description: "ID of the environment or environment alias.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["id"] = schema.StringAttribute{Description: "Composite Terraform identifier in `space_id/environment_id` form.", Computed: true}
	attributes["timeouts"] = timeouts.Attributes(ctx)

	return schema.Schema{Description: "Retrieves a Contentful Environment. Use `contentful_environment_status_ready` to wait for readiness.", Attributes: attributes}
}

func EnvironmentsDataSourceSchema(ctx context.Context) schema.Schema {
	attributes := map[string]schema.Attribute{}
	attributes["space_id"] = schema.StringAttribute{Description: "ID of the space.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["id"] = schema.StringAttribute{Description: "Terraform identifier for this lookup, equal to `space_id`.", Computed: true}
	attributes["timeouts"] = timeouts.Attributes(ctx)
	attributes["environments"] = schema.ListNestedAttribute{Description: "Environments in the space, ordered lexicographically by `environment_id`.", Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: environmentDataSourceItemAttributes()}}

	return schema.Schema{Description: "Retrieves all Contentful Environments in a space. Use `contentful_environment_status_ready` to wait for readiness.", Attributes: attributes}
}
