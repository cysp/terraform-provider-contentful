package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func environmentAliasDataSourceItemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"environment_alias_id":  schema.StringAttribute{Description: "ID of the environment alias.", Computed: true},
		"target_environment_id": schema.StringAttribute{Description: "ID of the environment reached through this alias.", Computed: true},
	}
}

func EnvironmentAliasDataSourceSchema(ctx context.Context) schema.Schema {
	attributes := environmentAliasDataSourceItemAttributes()
	attributes["space_id"] = schema.StringAttribute{Description: "ID of the space.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["environment_alias_id"] = schema.StringAttribute{Description: "ID of the environment alias.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["id"] = schema.StringAttribute{Description: "Composite Terraform identifier in `space_id/environment_alias_id` form.", Computed: true}
	attributes["timeouts"] = timeouts.Attributes(ctx)

	return schema.Schema{Description: "Retrieves a Contentful Environment Alias.", Attributes: attributes}
}

func EnvironmentAliasesDataSourceSchema(ctx context.Context) schema.Schema {
	attributes := map[string]schema.Attribute{}
	attributes["space_id"] = schema.StringAttribute{Description: "ID of the space.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["id"] = schema.StringAttribute{Description: "Terraform identifier for this lookup, equal to `space_id`.", Computed: true}
	attributes["timeouts"] = timeouts.Attributes(ctx)
	attributes["environment_aliases"] = schema.ListNestedAttribute{Description: "Environment aliases in the space, ordered lexicographically by `environment_alias_id`.", Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: environmentAliasDataSourceItemAttributes()}}

	return schema.Schema{Description: "Retrieves all Contentful Environment Aliases in a space.", Attributes: attributes}
}
