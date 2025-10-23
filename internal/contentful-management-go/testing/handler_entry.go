package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreateEntry(_ context.Context, req *cm.EntryRequest, params cm.CreateEntryParams) (cm.CreateEntryRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	entryID := generateResourceID()

	newEntry := NewEntryFromRequest(params.SpaceID, params.EnvironmentID, params.XContentfulContentType, entryID, req)
	ts.entries.Set(params.SpaceID, params.EnvironmentID, entryID, &newEntry)

	return &cm.EntryStatusCode{
		StatusCode: http.StatusCreated,
		Response:   newEntry,
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetEntry(_ context.Context, params cm.GetEntryParams) (cm.GetEntryRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID || params.EntryID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	entry := ts.entries.Get(params.SpaceID, params.EnvironmentID, params.EntryID)
	if entry == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Entry not found"), nil), nil
	}

	return entry, nil
}

//nolint:ireturn
func (ts *Handler) PutEntry(_ context.Context, req *cm.EntryRequest, params cm.PutEntryParams) (cm.PutEntryRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID || params.EntryID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	entry := ts.entries.Get(params.SpaceID, params.EnvironmentID, params.EntryID)
	if entry == nil {
		newEntry := NewEntryFromRequest(params.SpaceID, params.EnvironmentID, params.XContentfulContentType.Value, params.EntryID, req)
		ts.entries.Set(params.SpaceID, params.EnvironmentID, params.EntryID, &newEntry)

		return &cm.EntryStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newEntry,
		}, nil
	}

	UpdateEntryFromRequest(entry, req)

	return &cm.EntryStatusCode{
		StatusCode: http.StatusOK,
		Response:   *entry,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteEntry(_ context.Context, params cm.DeleteEntryParams) (cm.DeleteEntryRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID || params.EntryID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	entry := ts.entries.Get(params.SpaceID, params.EnvironmentID, params.EntryID)
	if entry == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Entry not found"), nil), nil
	}

	ts.entries.Delete(params.SpaceID, params.EnvironmentID, params.EntryID)

	return &cm.NoContent{}, nil
}

//nolint:ireturn
func (ts *Handler) PublishEntry(_ context.Context, params cm.PublishEntryParams) (cm.PublishEntryRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID || params.EntryID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	entry := ts.entries.Get(params.SpaceID, params.EnvironmentID, params.EntryID)
	if entry == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Entry not found"), nil), nil
	}

	publishEntry(entry)

	return &cm.EntryStatusCode{
		StatusCode: http.StatusOK,
		Response:   *entry,
	}, nil
}

//nolint:ireturn
func (ts *Handler) UnpublishEntry(_ context.Context, params cm.UnpublishEntryParams) (cm.UnpublishEntryRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID || params.EntryID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	entry := ts.entries.Get(params.SpaceID, params.EnvironmentID, params.EntryID)
	if entry == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Entry not found"), nil), nil
	}

	entry.Sys.PublishedVersion.Reset()

	return &cm.NoContent{}, nil
}
