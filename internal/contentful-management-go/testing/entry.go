package cmtesting

import (
	"encoding/json"

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

func projectEntryResponse(entry cm.Entry) cm.Entry {
	fields, ok := entry.Fields.Get()
	if !ok {
		return entry
	}

	responseFields := make(cm.EntryFields, len(fields))
	for fieldID, value := range fields {
		if entryFieldIsRawJSONNull(value) || entryFieldHasOnlyEmptyArrays(value) {
			continue
		}

		responseFields[fieldID] = value
	}

	if len(responseFields) == 0 {
		entry.Fields.Reset()
	} else {
		entry.Fields.SetTo(responseFields)
	}

	return entry
}

func entryFieldIsRawJSONNull(value []byte) bool {
	var decoded any

	return json.Unmarshal(value, &decoded) == nil && decoded == nil
}

func entryFieldHasOnlyEmptyArrays(value []byte) bool {
	var localized map[string]json.RawMessage

	err := json.Unmarshal(value, &localized)
	if err != nil || len(localized) == 0 {
		return false
	}

	for _, localeValue := range localized {
		var decoded any

		err = json.Unmarshal(localeValue, &decoded)
		if err != nil {
			return false
		}

		items, ok := decoded.([]any)
		if !ok || len(items) != 0 {
			return false
		}
	}

	return true
}
