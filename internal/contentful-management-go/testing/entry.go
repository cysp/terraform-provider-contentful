package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewEntryFromFields(spaceID, environmentID, entryID string, fields cm.EntryFields) cm.Entry {
	entry := cm.Entry{
		Sys: NewEntrySys(spaceID, environmentID, entryID),
	}

	UpdateEntryFromFields(&entry, fields)

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

func UpdateEntryFromFields(entry *cm.Entry, fields cm.EntryFields) {
	entry.Fields = fields
}

func publishEntry(entry *cm.Entry) {
	entry.Sys.PublishedVersion.SetTo(entry.Sys.Version)

	entry.Sys.Version++
}
