package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewSpaceEnablementFromRequestFields(spaceID string, spaceEnablementFields cm.SpaceEnablementData) cm.SpaceEnablement {
	spaceEnablement := NewSpaceEnablement(spaceID)

	UpdateSpaceEnablementFromRequestFields(&spaceEnablement, spaceEnablementFields)

	return spaceEnablement
}

func NewSpaceEnablement(spaceID string) cm.SpaceEnablement {
	return cm.SpaceEnablement{
		Sys: NewSpaceEnablementSys(spaceID),
	}
}

func NewSpaceEnablementSys(spaceID string) cm.SpaceEnablementSys {
	return cm.SpaceEnablementSys{
		Type: cm.SpaceEnablementSysTypeSpaceEnablement,
		Space: cm.SpaceLink{
			Sys: cm.SpaceLinkSys{
				Type:     cm.SpaceLinkSysTypeLink,
				LinkType: cm.SpaceLinkSysLinkTypeSpace,
				ID:       spaceID,
			},
		},
	}
}

func UpdateSpaceEnablementFromRequestFields(spaceEnablement *cm.SpaceEnablement, spaceEnablementFields cm.SpaceEnablementData) {
	spaceEnablement.Sys.Version++

	spaceEnablement.CrossSpaceLinks = spaceEnablementFields.CrossSpaceLinks
	spaceEnablement.SpaceTemplates = spaceEnablementFields.SpaceTemplates
	spaceEnablement.StudioExperiences = spaceEnablementFields.StudioExperiences
	spaceEnablement.SuggestConcepts = spaceEnablementFields.SuggestConcepts
}
