package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func localeDataSourceItemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"locale_id":              schema.StringAttribute{Description: "System ID of the locale, distinct from its code.", Computed: true},
		"name":                   schema.StringAttribute{Description: "Name of the locale.", Computed: true},
		"code":                   schema.StringAttribute{Description: "Locale code used for localized content, distinct from `locale_id`.", Computed: true},
		"default":                schema.BoolAttribute{Description: "Whether this is the default locale.", Computed: true},
		"fallback_code":          schema.StringAttribute{Description: "Locale code used as fallback, or null when no fallback is returned.", Computed: true},
		"optional":               schema.BoolAttribute{Description: "Whether required localized fields may be empty for this locale when publishing all locales together. This does not relax required fields in the default locale.", Computed: true},
		"content_management_api": schema.BoolAttribute{Description: "Whether localized content is available through the Content Management API and for editing in the web app. Disabled locales remain discoverable as Locale entities.", Computed: true},
		"content_delivery_api":   schema.BoolAttribute{Description: "Whether localized content is included in Content Delivery API and Content Preview API responses.", Computed: true},
	}
}

func LocaleDataSourceSchema(ctx context.Context) schema.Schema {
	attributes := localeDataSourceItemAttributes()
	attributes["space_id"] = schema.StringAttribute{Description: "ID of the space.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["environment_id"] = schema.StringAttribute{Description: "ID of the environment or environment alias.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["locale_id"] = schema.StringAttribute{Description: "System ID of the locale, distinct from its code.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["id"] = schema.StringAttribute{Description: "Composite Terraform identifier in `space_id/environment_id/locale_id` form.", Computed: true}
	attributes["timeouts"] = timeouts.Attributes(ctx)

	return schema.Schema{Description: "Retrieves a Contentful Locale. See [Using environment aliases](../guides/existing-configuration#use-environment-aliases) for alias lookups.", Attributes: attributes}
}

func LocalesDataSourceSchema(ctx context.Context) schema.Schema {
	attributes := map[string]schema.Attribute{}
	attributes["space_id"] = schema.StringAttribute{Description: "ID of the space.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["environment_id"] = schema.StringAttribute{Description: "ID of the environment or environment alias.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["id"] = schema.StringAttribute{Description: "Composite Terraform identifier in `space_id/environment_id` form.", Computed: true}
	attributes["timeouts"] = timeouts.Attributes(ctx)
	attributes["locales"] = schema.ListNestedAttribute{Description: "Locales in the environment, ordered lexicographically by `locale_id`.", Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: localeDataSourceItemAttributes()}}

	return schema.Schema{Description: "Retrieves all Contentful Locales in an environment. See [Using environment aliases](../guides/existing-configuration#use-environment-aliases) for alias lookups.", Attributes: attributes}
}
