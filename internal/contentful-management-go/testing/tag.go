package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewTagFromRequest(spaceID, environmentID, tagID string, req *cm.TagData, visibility cm.OptString) cm.Tag {
	tag := cm.Tag{
		Sys: cm.NewTagSys(spaceID, environmentID, tagID),
	}

	UpdateTagFromRequest(&tag, req, visibility)

	return tag
}

func UpdateTagFromRequest(tag *cm.Tag, req *cm.TagData, visibility cm.OptString) {
	tag.Sys.Version++

	tag.Name = req.Name

	if visibility.IsSet() {
		tag.Sys.Visibility.SetTo(cm.TagSysVisibility(visibility.Value))
	}
}
