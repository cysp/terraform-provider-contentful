package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewEntryFromRequest(spaceID, environmentID, contentTypeID, entryID string, req *cm.EntryRequest) cm.Entry {
	entry := cm.Entry{
		Sys: cm.NewEntrySys(spaceID, environmentID, contentTypeID, entryID),
	}

	UpdateEntryFromRequest(&entry, req)

	return entry
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
