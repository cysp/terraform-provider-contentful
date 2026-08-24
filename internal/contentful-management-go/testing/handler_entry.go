package cmtesting

import (
	"cmp"
	"context"
	"net/http"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetEntries(_ context.Context, params cm.GetEntriesParams) (cm.GetEntriesRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	skip := params.Skip.Or(0)
	if skip < 0 {
		return NewContentfulManagementErrorStatusCodeInvalidQuery(new(`The value provided for "skip" is invalid. Please provide a value larger than or equal to 0`), nil), nil
	}

	limit := params.Limit.Or(100) //nolint:mnd
	if limit < 0 || limit > 1000 {
		return NewContentfulManagementErrorStatusCodeInvalidQuery(new(`The value provided for "limit" is invalid. Please provide a value between 0 and 1000`), nil), nil
	}

	entries := make([]cm.Entry, 0)

	for _, entry := range ts.entries.List(params.SpaceID, params.EnvironmentID) {
		if params.ContentType.IsSet() && entry.Sys.ContentType.Sys.ID != params.ContentType.Value {
			continue
		}

		entries = append(entries, projectEntryResponse(*entry))
	}

	slices.SortFunc(entries, func(a, b cm.Entry) int {
		return cmp.Compare(a.Sys.ID, b.Sys.ID)
	})

	start := min(skip, int64(len(entries)))
	end := min(start+limit, int64(len(entries)))

	collection := cm.EntryCollection{
		Sys: cm.EntryCollectionSys{
			Type: cm.EntryCollectionSysTypeArray,
		},
		Total: cm.NewOptInt(len(entries)),
		Skip:  cm.NewOptInt(int(skip)),
		Limit: cm.NewOptInt(int(limit)),
		Items: entries[start:end],
	}

	return &collection, nil
}

//nolint:ireturn
func (ts *Handler) CreateEntry(_ context.Context, req *cm.EntryRequest, params cm.CreateEntryParams) (cm.CreateEntryRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	entryID := generateResourceID()

	newEntry := NewEntryFromRequest(params.SpaceID, params.EnvironmentID, params.XContentfulContentType, entryID, req)
	ts.entries.Set(params.SpaceID, params.EnvironmentID, entryID, &newEntry)

	return &cm.EntryStatusCode{
		StatusCode: http.StatusCreated,
		Response:   ts.entryMutationResponse(newEntry),
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetEntry(_ context.Context, params cm.GetEntryParams) (cm.GetEntryRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	entry := ts.entries.Get(params.SpaceID, params.EnvironmentID, params.EntryID)
	if entry == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Entry not found"), nil), nil
	}

	response := projectEntryResponse(*entry)

	return &response, nil
}

//nolint:ireturn
func (ts *Handler) PutEntry(_ context.Context, req *cm.EntryRequest, params cm.PutEntryParams) (cm.PutEntryRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	entry := ts.entries.Get(params.SpaceID, params.EnvironmentID, params.EntryID)
	if entry == nil {
		newEntry := NewEntryFromRequest(params.SpaceID, params.EnvironmentID, params.XContentfulContentType.Value, params.EntryID, req)
		ts.entries.Set(params.SpaceID, params.EnvironmentID, params.EntryID, &newEntry)

		return &cm.EntryStatusCode{
			StatusCode: http.StatusCreated,
			Response:   ts.entryMutationResponse(newEntry),
		}, nil
	}

	if params.XContentfulVersion != entry.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	UpdateEntryFromRequest(entry, req)

	return &cm.EntryStatusCode{
		StatusCode: http.StatusOK,
		Response:   ts.entryMutationResponse(*entry),
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteEntry(_ context.Context, params cm.DeleteEntryParams) (cm.DeleteEntryRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	entry := ts.entries.Get(params.SpaceID, params.EnvironmentID, params.EntryID)
	if entry == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Entry not found"), nil), nil
	}

	ts.entries.Delete(params.SpaceID, params.EnvironmentID, params.EntryID)

	return &cm.NoContent{}, nil
}

//nolint:ireturn
func (ts *Handler) PublishEntry(_ context.Context, params cm.PublishEntryParams) (cm.PublishEntryRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	entry := ts.entries.Get(params.SpaceID, params.EnvironmentID, params.EntryID)
	if entry == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Entry not found"), nil), nil
	}

	if params.XContentfulVersion != entry.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	publishEntry(entry)

	return &cm.EntryStatusCode{
		StatusCode: http.StatusOK,
		Response:   ts.entryMutationResponse(*entry),
	}, nil
}

func (ts *Handler) entryMutationResponse(entry cm.Entry) cm.Entry {
	entry = projectEntryResponse(entry)

	if ts.omitEntryMutationResponseFields {
		entry.Fields = cm.OptEntryFields{}
	}

	return entry
}

//nolint:ireturn
func (ts *Handler) UnpublishEntry(_ context.Context, params cm.UnpublishEntryParams) (cm.UnpublishEntryRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	entry := ts.entries.Get(params.SpaceID, params.EnvironmentID, params.EntryID)
	if entry == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Entry not found"), nil), nil
	}

	entry.Sys.Version++
	entry.Sys.PublishedVersion.Reset()

	response := ts.entryMutationResponse(*entry)

	return &response, nil
}
