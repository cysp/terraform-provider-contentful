package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (m *SpaceEnablementsModel) ToSpaceEnablementData(_ context.Context, config SpaceEnablementsModel) (cm.SpaceEnablementData, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	diags.Append(rejectUnknownConfigurationOwnedRequestValue(m.CrossSpaceLinks, config.CrossSpaceLinks, path.Root("cross_space_links"))...)
	diags.Append(rejectUnknownConfigurationOwnedRequestValue(m.SpaceTemplates, config.SpaceTemplates, path.Root("space_templates"))...)
	diags.Append(rejectUnknownConfigurationOwnedRequestValue(m.StudioExperiences, config.StudioExperiences, path.Root("studio_experiences"))...)
	diags.Append(rejectUnknownConfigurationOwnedRequestValue(m.SuggestConcepts, config.SuggestConcepts, path.Root("suggest_concepts"))...)

	if diags.HasError() {
		return cm.SpaceEnablementData{}, diags
	}

	crossSpaceLinks, crossSpaceLinksDiags := requestRequiredBool(m.CrossSpaceLinks, path.Root("cross_space_links"))
	diags.Append(crossSpaceLinksDiags...)

	spaceTemplates, spaceTemplatesDiags := requestRequiredBool(m.SpaceTemplates, path.Root("space_templates"))
	diags.Append(spaceTemplatesDiags...)

	if !crossSpaceLinksDiags.HasError() && !spaceTemplatesDiags.HasError() && crossSpaceLinks != spaceTemplates {
		diags.AddAttributeError(
			path.Root("space_templates"),
			"Unequal coupled space enablements",
			"cross_space_links and space_templates must have the same value before they can be sent to Contentful.",
		)
	}

	if diags.HasError() {
		return cm.SpaceEnablementData{}, diags
	}

	fields := cm.SpaceEnablementData{}

	fields.CrossSpaceLinks.SetTo(cm.SpaceEnablementField{Enabled: crossSpaceLinks})
	fields.SpaceTemplates.SetTo(cm.SpaceEnablementField{Enabled: spaceTemplates})
	setOptSpaceEnablementFieldFromBoolValue(&fields.StudioExperiences, m.StudioExperiences)
	setOptSpaceEnablementFieldFromBoolValue(&fields.SuggestConcepts, m.SuggestConcepts)

	return fields, diags
}

func setOptSpaceEnablementFieldFromBoolValue(field *cm.OptSpaceEnablementField, value types.Bool) {
	switch {
	case !value.IsUnknown() && !value.IsNull():
		field.SetTo(cm.SpaceEnablementField{
			Enabled: value.ValueBool(),
		})
	default:
		// Unknown response-owned Optional+Computed enablements intentionally stay
		// absent; configuration-owned unknown plans were rejected above.
		field.Reset()
	}
}
