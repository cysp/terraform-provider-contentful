package cmtesting

import (
	"cmp"
	"context"
	"net/http"
	"slices"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

// contentTypePublicationSnapshot records the published field set and whether
// each field was omitted in that publication.
type contentTypePublicationSnapshot map[string]bool

//nolint:ireturn
func (ts *Handler) GetContentTypes(_ context.Context, params cm.GetContentTypesParams) (cm.GetContentTypesRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	contentTypes := ts.contentTypes.List(params.SpaceID, params.EnvironmentID)

	skip := params.Skip.Or(0)
	if skip < 0 {
		return NewContentfulManagementErrorStatusCodeInvalidQuery(new(`The value provided for "skip" is invalid. Please provide a value larger than or equal to 0`), nil), nil
	}

	limit := params.Limit.Or(100) //nolint:mnd
	if limit < 0 || limit > 1000 {
		return NewContentfulManagementErrorStatusCodeInvalidQuery(new(`The value provided for "limit" is invalid. Please provide a value between 0 and 1000`), nil), nil
	}

	items := make([]cm.ContentType, 0, len(contentTypes))
	for _, ct := range contentTypes {
		items = append(items, *ct)
	}

	slices.SortFunc(items, func(a, b cm.ContentType) int {
		return cmp.Compare(a.Sys.ID, b.Sys.ID)
	})

	start := min(skip, int64(len(items)))
	end := min(start+limit, int64(len(items)))

	return &cm.ContentTypeCollection{
		Sys: cm.ContentTypeCollectionSys{
			Type: cm.ContentTypeCollectionSysTypeArray,
		},
		Total: cm.NewOptInt(len(contentTypes)),
		Skip:  cm.NewOptInt(int(skip)),
		Limit: cm.NewOptInt(int(limit)),
		Items: items[start:end],
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetContentType(_ context.Context, params cm.GetContentTypeParams) (cm.GetContentTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	contentType := ts.contentTypes.Get(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	if contentType == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("ContentType not found"), nil), nil
	}

	return contentType, nil
}

//nolint:ireturn
func (ts *Handler) PutContentType(_ context.Context, req *cm.ContentTypeRequestData, params cm.PutContentTypeParams) (cm.PutContentTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environment := ts.environments.Get(params.SpaceID, params.EnvironmentID)
	if environment == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	if metadata, metadataSet := req.Metadata.Get(); metadataSet && contentTypeMetadataIsRejectedAsEmpty(metadata) {
		return NewContentfulManagementErrorStatusCodeValidationFailed(new("Validation error"), nil), nil
	}

	contentType := ts.contentTypes.Get(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	if contentType == nil {
		newContentType := NewContentTypeFromRequestFields(params.SpaceID, params.EnvironmentID, params.ContentTypeID, *req)
		ts.contentTypes.Set(params.SpaceID, params.EnvironmentID, params.ContentTypeID, &newContentType)

		return &cm.ContentTypeStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newContentType,
		}, nil
	}

	version, versionSet := params.XContentfulVersion.Get()
	if !versionSet || version != contentType.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	if contentType.Sys.PublishedVersion.IsSet() {
		publicationSnapshot := ts.contentTypePublications.Get(
			params.SpaceID,
			params.EnvironmentID,
			params.ContentTypeID,
		)
		if removesNonOmittedPublishedField(publicationSnapshot, req.Fields) {
			return NewContentfulManagementErrorStatusCodeBadRequest(
				new("You need to omit a field before deleting it"),
				nil,
			), nil
		}
	}

	UpdateContentTypeFromRequestFields(contentType, *req)

	return &cm.ContentTypeStatusCode{
		StatusCode: http.StatusOK,
		Response:   *contentType,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteContentType(_ context.Context, params cm.DeleteContentTypeParams) (cm.DeleteContentTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	contentType := ts.contentTypes.Get(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	if contentType == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("ContentType not found"), nil), nil
	}

	if contentType.Sys.PublishedVersion.IsSet() {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("Cannot delete published"), nil), nil
	}

	ts.contentTypes.Delete(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	ts.contentTypePublications.Delete(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	ts.editorInterfaces.Delete(params.SpaceID, params.EnvironmentID, params.ContentTypeID)

	return &cm.NoContent{}, nil
}

//nolint:ireturn
func (ts *Handler) ActivateContentType(_ context.Context, params cm.ActivateContentTypeParams) (cm.ActivateContentTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	contentType := ts.contentTypes.Get(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	if contentType == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("ContentType not found"), nil), nil
	}

	if params.XContentfulVersion != contentType.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	publishContentType(contentType, time.Now().UTC())

	editorInterface := ts.editorInterfaces.Get(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	if editorInterface == nil {
		newEditorInterface := NewDefaultEditorInterface(params.SpaceID, params.EnvironmentID, params.ContentTypeID, contentType.Fields)
		ts.editorInterfaces.Set(params.SpaceID, params.EnvironmentID, params.ContentTypeID, &newEditorInterface)
	} else {
		previousPublicationSnapshot := ts.contentTypePublications.Get(
			params.SpaceID,
			params.EnvironmentID,
			params.ContentTypeID,
		)

		SyncEditorInterfaceWithContentType(editorInterface, previousPublicationSnapshot, contentType.Fields)
	}

	ts.contentTypePublications.Set(
		params.SpaceID,
		params.EnvironmentID,
		params.ContentTypeID,
		newContentTypePublicationSnapshot(contentType.Fields),
	)

	return &cm.ContentTypeStatusCode{
		StatusCode: http.StatusOK,
		Response:   *contentType,
	}, nil
}

func newContentTypePublicationSnapshot(fields []cm.ContentTypeFieldsItem) *contentTypePublicationSnapshot {
	snapshot := make(contentTypePublicationSnapshot, len(fields))

	for _, field := range fields {
		snapshot[field.ID] = field.Omitted.Or(false)
	}

	return &snapshot
}

func removesNonOmittedPublishedField(
	publicationSnapshot *contentTypePublicationSnapshot,
	requestedFields []cm.ContentTypeRequestDataFieldsItem,
) bool {
	if publicationSnapshot == nil {
		return false
	}

	requestedFieldIDs := make(map[string]struct{}, len(requestedFields))
	for _, field := range requestedFields {
		requestedFieldIDs[field.ID] = struct{}{}
	}

	for fieldID, wasOmitted := range *publicationSnapshot {
		if _, retained := requestedFieldIDs[fieldID]; retained {
			continue
		}

		if !wasOmitted {
			return true
		}
	}

	return false
}

//nolint:ireturn
func (ts *Handler) DeactivateContentType(_ context.Context, params cm.DeactivateContentTypeParams) (cm.DeactivateContentTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	contentType := ts.contentTypes.Get(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	if contentType == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("ContentType not found"), nil), nil
	}

	if !contentType.Sys.PublishedVersion.IsSet() {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("Not published"), nil), nil
	}

	contentType.Sys.PublishedVersion.Reset()
	contentType.Sys.PublishedAt.Reset()
	contentType.Sys.Version++

	ts.editorInterfaces.Delete(params.SpaceID, params.EnvironmentID, params.ContentTypeID)

	return contentType, nil
}
