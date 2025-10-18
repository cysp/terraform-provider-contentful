package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewEntryFromRequest(spaceID, environmentID, contentTypeID, entryID string, req *cm.EntryRequest) cm.Entry {
	entry := cm.Entry{
		Sys: NewEntrySys(spaceID, environmentID, contentTypeID, entryID),
	}

	UpdateEntryFromRequest(&entry, req)

	return entry
}

func NewEntrySys(spaceID, environmentID, contentTypeID, entryID string) cm.EntrySys {
	return cm.EntrySys{
		Type: cm.EntrySysTypeEntry,
		Space: cm.SpaceLink{
			Sys: cm.SpaceLinkSys{
				Type:     cm.SpaceLinkSysTypeLink,
				LinkType: cm.SpaceLinkSysLinkTypeSpace,
				ID:       spaceID,
			},
		},
		Environment: cm.EnvironmentLink{
			Sys: cm.EnvironmentLinkSys{
				Type:     cm.EnvironmentLinkSysTypeLink,
				LinkType: cm.EnvironmentLinkSysLinkTypeEnvironment,
				ID:       environmentID,
			},
		},
		ContentType: cm.ContentTypeLink{
			Sys: cm.ContentTypeLinkSys{
				Type:     cm.ContentTypeLinkSysTypeLink,
				LinkType: cm.ContentTypeLinkSysLinkTypeContentType,
				ID:       contentTypeID,
			},
		},
		ID: entryID,
	}
}

func UpdateEntryFromRequest(entry *cm.Entry, req *cm.EntryRequest) {
	entry.Sys.Version++

	entry.Fields = req.Fields
	entry.Metadata = req.Metadata
}

func publishEntry(entry *cm.Entry) {
	entry.Sys.PublishedVersion.SetTo(entry.Sys.Version)

	entry.Sys.Version++
}
