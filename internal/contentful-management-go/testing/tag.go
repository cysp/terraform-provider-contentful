package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewTagFromRequest(spaceID, environmentID, tagID string, req *cm.TagRequest) cm.Tag {
	tag := cm.Tag{
		Sys: cm.NewTagSys(spaceID, environmentID, tagID),
	}

	if visibility, ok := req.Sys.Visibility.Get(); ok {
		tag.Sys.Visibility = visibility
	}

	UpdateTagFromRequest(&tag, req)

	return tag
}

func UpdateTagFromRequest(tag *cm.Tag, req *cm.TagRequest) {
	tag.Sys.Version++

	tag.Name = req.Name
}
