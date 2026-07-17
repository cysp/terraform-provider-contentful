package contentfulmanagement

import "net/http"

// putPreviewEnvironmentRes keeps the generic default error response available
// alongside the explicitly modeled 400, 404, and 409 responses. Ogen does not
// currently generate this marker when multiple concrete error statuses share
// the default error schema.
func (*ErrorStatusCode) putPreviewEnvironmentRes() {}

func (*PutPreviewEnvironmentBadRequest) GetStatusCode() int {
	return http.StatusBadRequest
}

func (r *PutPreviewEnvironmentBadRequest) GetError() (Error, bool) {
	return (*ErrorStatusCode)(r).GetError()
}

func (*PutPreviewEnvironmentNotFound) GetStatusCode() int {
	return http.StatusNotFound
}

func (r *PutPreviewEnvironmentNotFound) GetError() (Error, bool) {
	return (*ErrorStatusCode)(r).GetError()
}

func (*PutPreviewEnvironmentConflict) GetStatusCode() int {
	return http.StatusConflict
}

func (r *PutPreviewEnvironmentConflict) GetError() (Error, bool) {
	return (*ErrorStatusCode)(r).GetError()
}
