package contentfulmanagementtestserver

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewPreviewAPIKeySys(spaceID, previewAPIKeyID string) cm.PreviewApiKeySys {
	return cm.PreviewApiKeySys{
		Type: cm.PreviewApiKeySysTypePreviewApiKey,
		ID:   previewAPIKeyID,
		Space: cm.SpaceLink{
			Sys: cm.SpaceLinkSys{
				Type:     cm.SpaceLinkSysTypeLink,
				LinkType: cm.SpaceLinkSysLinkTypeSpace,
				ID:       spaceID,
			},
		},
	}
}
