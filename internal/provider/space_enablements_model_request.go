package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (m *SpaceEnablementsModel) ToSpaceEnablementData(_ context.Context) (cm.SpaceEnablementData, diag.Diagnostics) {
	fields := cm.SpaceEnablementData{}

	setOptSpaceEnablementFieldFromBoolValue(&fields.CrossSpaceLinks, m.CrossSpaceLinks)
	setOptSpaceEnablementFieldFromBoolValue(&fields.SpaceTemplates, m.SpaceTemplates)
	setOptSpaceEnablementFieldFromBoolValue(&fields.StudioExperiences, m.StudioExperiences)
	setOptSpaceEnablementFieldFromBoolValue(&fields.SuggestConcepts, m.SuggestConcepts)

	return fields, nil
}

func setOptSpaceEnablementFieldFromBoolValue(field *cm.OptSpaceEnablementField, value types.Bool) {
	switch {
	case !value.IsUnknown() && !value.IsNull():
		field.SetTo(cm.SpaceEnablementField{
			Enabled: value.ValueBool(),
		})
	default:
		field.Reset()
	}
}
