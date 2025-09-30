package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SpaceEnablementsIdentityModel struct {
	SpaceID types.String `tfsdk:"space_id"`
}

type SpaceEnablementsModel struct {
	SpaceEnablementsIdentityModel

	ID                types.String `tfsdk:"id"`
	CrossSpaceLinks   types.Bool   `tfsdk:"cross_space_links"`
	SpaceTemplates    types.Bool   `tfsdk:"space_templates"`
	StudioExperiences types.Bool   `tfsdk:"studio_experiences"`
	SuggestConcepts   types.Bool   `tfsdk:"suggest_concepts"`
}
