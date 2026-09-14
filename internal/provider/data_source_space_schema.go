package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func spaceDataSourceItemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"space_id":        schema.StringAttribute{Description: "System ID of the space.", Computed: true},
		"name":            schema.StringAttribute{Description: "Name of the space.", Computed: true},
		"organization_id": schema.StringAttribute{Description: "ID of the organization containing the space.", Computed: true},
	}
}

func SpaceDataSourceSchema(ctx context.Context) schema.Schema {
	attributes := spaceDataSourceItemAttributes()
	attributes["space_id"] = schema.StringAttribute{Description: "ID of the space.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["id"] = schema.StringAttribute{Description: "Terraform identifier for this lookup, equal to `space_id`.", Computed: true}
	attributes["timeouts"] = timeouts.Attributes(ctx)

	return schema.Schema{Description: "Retrieves a Contentful Space.", Attributes: attributes}
}

func SpacesDataSourceSchema(ctx context.Context) schema.Schema {
	attributes := map[string]schema.Attribute{}
	attributes["organization_id"] = schema.StringAttribute{Description: "ID of the organization to list spaces from. Omitted lists spaces from all accessible organizations.", Optional: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["id"] = schema.StringAttribute{Description: "Terraform lookup identifier: `spaces` without organization scope, otherwise `organizations/organization_id/spaces`.", Computed: true}
	attributes["timeouts"] = timeouts.Attributes(ctx)
	attributes["spaces"] = schema.ListNestedAttribute{Description: "Spaces accessible to the account, ordered lexicographically by `space_id`.", Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: spaceDataSourceItemAttributes()}}

	return schema.Schema{Description: "Retrieves all Contentful Spaces accessible to the account.", Attributes: attributes}
}
