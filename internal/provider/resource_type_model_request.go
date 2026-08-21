package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (r *ResourceTypeModel) ToResourceTypeData(_ context.Context, modelPath path.Path) (cm.ResourceTypeData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	name, nameDiags := requestRequiredString(r.Name, modelPath.AtName("name"))
	diags.Append(nameDiags...)

	defaultFieldMappingPath := modelPath.AtName("default_field_mapping")

	if r.DefaultFieldMapping == nil {
		diags.AddAttributeError(
			defaultFieldMappingPath,
			"Unexpected null default field mapping",
			"The default field mapping is required.",
		)

		return cm.ResourceTypeData{}, diags
	}

	title, titleDiags := requestRequiredString(r.DefaultFieldMapping.Title, defaultFieldMappingPath.AtName("title"))
	diags.Append(titleDiags...)

	subtitle, subtitleDiags := requestOmittableString(r.DefaultFieldMapping.Subtitle, defaultFieldMappingPath.AtName("subtitle"))
	diags.Append(subtitleDiags...)

	description, descriptionDiags := requestOmittableString(r.DefaultFieldMapping.Description, defaultFieldMappingPath.AtName("description"))
	diags.Append(descriptionDiags...)

	externalURL, externalURLDiags := requestOmittableString(r.DefaultFieldMapping.ExternalURL, defaultFieldMappingPath.AtName("external_url"))
	diags.Append(externalURLDiags...)

	resourceTypeFields := cm.ResourceTypeData{
		Name: name,
		DefaultFieldMapping: cm.ResourceTypeDefaultFieldMapping{
			Title:       title,
			Subtitle:    subtitle,
			Description: description,
			ExternalUrl: externalURL,
		},
	}

	if r.DefaultFieldMapping.Image != nil {
		imagePath := defaultFieldMappingPath.AtName("image")
		url, urlDiags := requestRequiredString(r.DefaultFieldMapping.Image.URL, imagePath.AtName("url"))
		diags.Append(urlDiags...)

		altText, altTextDiags := requestOmittableString(r.DefaultFieldMapping.Image.AltText, imagePath.AtName("alt_text"))
		diags.Append(altTextDiags...)

		resourceTypeFields.DefaultFieldMapping.Image.SetTo(cm.ResourceTypeDefaultFieldMappingImage{
			URL:     url,
			AltText: altText,
		})
	}

	if r.DefaultFieldMapping.Badge != nil {
		badgePath := defaultFieldMappingPath.AtName("badge")
		label, labelDiags := requestRequiredString(r.DefaultFieldMapping.Badge.Label, badgePath.AtName("label"))
		diags.Append(labelDiags...)

		variant, variantDiags := requestRequiredString(r.DefaultFieldMapping.Badge.Variant, badgePath.AtName("variant"))
		diags.Append(variantDiags...)

		resourceTypeFields.DefaultFieldMapping.Badge.SetTo(cm.ResourceTypeDefaultFieldMappingBadge{
			Label:   label,
			Variant: variant,
		})
	}

	if diags.HasError() {
		return cm.ResourceTypeData{}, diags
	}

	return resourceTypeFields, diags
}
