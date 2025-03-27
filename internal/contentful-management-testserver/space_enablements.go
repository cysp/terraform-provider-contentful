package contentfulmanagementtestserver

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewSpaceEnablement(spaceID string) cm.SpaceEnablement {
	return cm.SpaceEnablement{
		Sys: cm.SpaceEnablementSys{
			Type: cm.SpaceEnablementSysTypeSpaceEnablement,
			Space: cm.SpaceLink{
				Sys: cm.SpaceLinkSys{
					Type:     cm.SpaceLinkSysTypeLink,
					LinkType: cm.SpaceLinkSysLinkTypeSpace,
					ID:       spaceID,
				},
			},
			Version: 1,
		},
	}
}

func UpdateSpaceEnablementFromFields(spaceEnablement *cm.SpaceEnablement, spaceEnablementFields cm.SpaceEnablementFields) {
	spaceEnablement.Sys.Version++

	spaceEnablement.CrossSpaceLinks = spaceEnablementFields.CrossSpaceLinks
	spaceEnablement.SpaceTemplates = spaceEnablementFields.SpaceTemplates
	spaceEnablement.StudioExperiences = spaceEnablementFields.StudioExperiences
	spaceEnablement.SuggestConcepts = spaceEnablementFields.SuggestConcepts
}
