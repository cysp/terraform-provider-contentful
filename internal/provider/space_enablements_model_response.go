package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSpaceEnablementsResourceModelFromResponse(_ context.Context, response cm.SpaceEnablement) (SpaceEnablementsModel, diag.Diagnostics) {
	spaceID := response.Sys.Space.Sys.ID

	model := SpaceEnablementsModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID([]string{spaceID}),
		SpaceEnablementsIdentityModel: SpaceEnablementsIdentityModel{
			SpaceID: types.StringValue(spaceID),
		},
	}

	model.CrossSpaceLinks = boolValueFromOptSpaceEnablementField(response.CrossSpaceLinks)
	model.SpaceTemplates = boolValueFromOptSpaceEnablementField(response.SpaceTemplates)
	model.StudioExperiences = boolValueFromOptSpaceEnablementField(response.StudioExperiences)
	model.SuggestConcepts = boolValueFromOptSpaceEnablementField(response.SuggestConcepts)

	return model, nil
}

func boolValueFromOptSpaceEnablementField(field cm.OptSpaceEnablementField) types.Bool {
	if field, ok := field.Get(); ok {
		return types.BoolValue(field.Enabled)
	}

	return types.BoolNull()
}
