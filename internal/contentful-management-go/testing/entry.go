package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewEntryFromRequest(spaceID, environmentID, entryID string, req *cm.EntryRequest) cm.Entry {
	entry := cm.Entry{
		Sys: NewEntrySys(spaceID, environmentID, entryID),
	}

	UpdateEntryFromRequest(&entry, req)

	return entry
}

func NewEntrySys(spaceID, environmentID, entryID string) cm.EntrySys {
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
