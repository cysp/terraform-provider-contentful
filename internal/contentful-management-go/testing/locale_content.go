package cmtesting

import (
	"encoding/json"
	"maps"
	"net/http"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
)

// Localized maps are stored in full. Editing flags affect response projection;
// code changes and deletion change the stored keys without advancing content versions.
func (ts *Handler) changeLocaleContent(spaceID, environmentID, oldCode, newCode string) {
	for _, entry := range ts.entries.List(spaceID, environmentID) {
		if fields, ok := entry.Fields.Get(); ok {
			fields = maps.Clone(fields)
			for fieldID, value := range fields {
				fields[fieldID] = changeLocaleKey(value, oldCode, newCode)
			}

			entry.Fields.SetTo(fields)
		}
	}

	for _, contentType := range ts.contentTypes.List(spaceID, environmentID) {
		for index := range contentType.Fields {
			field := &contentType.Fields[index]
			field.DefaultValue = changeLocaleKey(field.DefaultValue, oldCode, newCode)
		}
	}
}

func changeLocaleKey(value jx.Raw, oldCode, newCode string) jx.Raw {
	var localized map[string]json.RawMessage
	if json.Unmarshal(value, &localized) != nil {
		return value
	}

	old, found := localized[oldCode]
	if !found {
		return value
	}

	delete(localized, oldCode)

	if newCode != "" {
		localized[newCode] = old
	}

	// Values came from successful JSON decoding; only their string keys changed.
	result, _ := json.Marshal(localized) //nolint:errchkjson // All RawMessages were decoded above.

	return result
}

func projectLocaleValue(value jx.Raw, locales []*cm.Locale) jx.Raw {
	var localized map[string]json.RawMessage
	if json.Unmarshal(value, &localized) != nil {
		return value
	}

	changed := false

	for _, locale := range locales {
		if _, found := localized[locale.Code]; found && !locale.ContentManagementApi {
			delete(localized, locale.Code)

			changed = true
		}
	}

	if !changed {
		return value
	}

	// Values came from successful JSON decoding; only their string keys changed.
	result, _ := json.Marshal(localized) //nolint:errchkjson // All RawMessages were decoded above.

	return result
}

func (ts *Handler) projectEntryResponse(entry cm.Entry) cm.Entry {
	if fields, ok := entry.Fields.Get(); ok {
		fields = maps.Clone(fields)

		locales := ts.locales.List(entry.Sys.Space.Sys.ID, entry.Sys.Environment.Sys.ID)
		for id, value := range fields {
			fields[id] = projectLocaleValue(value, locales)
		}

		entry.Fields.SetTo(fields)
	}

	return projectEntryResponse(entry)
}

func (ts *Handler) projectContentTypeResponse(contentType cm.ContentType) cm.ContentType {
	contentType.Fields = slices.Clone(contentType.Fields)

	locales := ts.locales.List(contentType.Sys.Space.Sys.ID, contentType.Sys.Environment.Sys.ID)
	for index := range contentType.Fields {
		field := &contentType.Fields[index]
		field.DefaultValue = projectLocaleValue(field.DefaultValue, locales)
	}

	return contentType
}

func localeValueIsWritable(value jx.Raw, locales []*cm.Locale) bool {
	// Existing tests without an explicit Locale inventory do not specify this
	// validation boundary. Locale conformance fixtures always register inventory.
	if len(locales) == 0 {
		return true
	}

	var localized map[string]json.RawMessage
	if json.Unmarshal(value, &localized) != nil {
		return true
	}

	for code := range localized {
		if !slices.ContainsFunc(locales, func(locale *cm.Locale) bool {
			return locale.Code == code && locale.ContentManagementApi
		}) {
			return false
		}
	}

	return true
}

func (ts *Handler) validateEntryLocales(spaceID, environmentID string, req *cm.EntryRequest) *cm.ErrorStatusCode {
	locales := ts.locales.List(spaceID, environmentID)
	for _, value := range req.Fields.Value {
		if !localeValueIsWritable(value, locales) {
			return NewContentfulManagementErrorStatusCodeValidationFailed(new("Invalid field locale code"), nil)
		}
	}

	return nil
}

func (ts *Handler) validateContentTypeLocales(spaceID, environmentID string, req *cm.ContentTypeRequestData) *cm.ErrorStatusCode {
	locales := ts.locales.List(spaceID, environmentID)
	for _, field := range req.Fields {
		if !localeValueIsWritable(field.DefaultValue, locales) {
			return NewContentfulManagementErrorStatusCodeValidationFailed(new("Invalid default value locale code"), nil)
		}
	}

	return nil
}

func (ts *Handler) validateEntryRequiredLocales(entry *cm.Entry) *cm.ErrorStatusCode {
	contentType := ts.contentTypes.Get(entry.Sys.Space.Sys.ID, entry.Sys.Environment.Sys.ID, entry.Sys.ContentType.Sys.ID)

	locales := ts.locales.List(entry.Sys.Space.Sys.ID, entry.Sys.Environment.Sys.ID)
	if contentType == nil || len(locales) == 0 {
		return nil
	}

	for _, field := range contentType.Fields {
		if !field.Required.Or(false) {
			continue
		}

		var localized map[string]json.RawMessage

		_ = json.Unmarshal(entry.Fields.Value[field.ID], &localized)
		if len(localized) == 0 {
			return NewContentfulManagementErrorStatusCode(http.StatusUnprocessableEntity, "InvalidEntry", new("Validation error"), nil)
		}

		for _, locale := range locales {
			required := locale.Default
			if field.Localized.Or(false) {
				required = locale.ContentManagementApi && (locale.Default || !locale.Optional)
			}

			if required && (len(localized[locale.Code]) == 0 || entryFieldIsRawJSONNull(localized[locale.Code])) {
				return NewContentfulManagementErrorStatusCode(http.StatusUnprocessableEntity, "InvalidEntry", new("Validation error"), nil)
			}
		}
	}

	return nil
}
