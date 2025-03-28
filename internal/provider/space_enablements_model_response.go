package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (m *SpaceEnablementsResourceModel) ReadFromResponse(_ context.Context, response cm.SpaceEnablement) diag.Diagnostics {
	spaceID := response.Sys.Space.Sys.ID

	m.ID = types.StringValue(spaceID)
	m.SpaceID = types.StringValue(spaceID)

	m.CrossSpaceLinks = boolValueFromOptSpaceEnablementField(response.CrossSpaceLinks)
	m.SpaceTemplates = boolValueFromOptSpaceEnablementField(response.SpaceTemplates)
	m.StudioExperiences = boolValueFromOptSpaceEnablementField(response.StudioExperiences)
	m.SuggestConcepts = boolValueFromOptSpaceEnablementField(response.SuggestConcepts)

	return nil
}

func boolValueFromOptSpaceEnablementField(field cm.OptSpaceEnablementField) types.Bool {
	if field, ok := field.Get(); ok {
		return types.BoolValue(field.Enabled)
	}

	return types.BoolNull()
}
