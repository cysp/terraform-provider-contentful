package provider_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	entryTestUpdatePath  = "/spaces/space/environments/environment/entries/entry"
	entryTestPublishPath = entryTestUpdatePath + "/published"
)

type entryMutationRequest struct {
	method         string
	path           string
	version        string
	versionPresent bool
	versionValues  []string
	body           []byte
	contentLength  int64
	contentType    string
	fields         map[string]json.RawMessage
}

type entryMutationRecorder struct {
	delegate  http.Handler
	errorSink *entryFixtureErrorSink

	mu       sync.Mutex
	requests []entryMutationRequest
}

func newEntryMutationRecorder(delegate http.Handler, errorSink *entryFixtureErrorSink) *entryMutationRecorder {
	return &entryMutationRecorder{delegate: delegate, errorSink: errorSink}
}

func (h *entryMutationRecorder) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	h.recordUpdate(request)

	if request.Method == http.MethodPut && strings.HasSuffix(request.URL.Path, "/entries/entry/published") {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			h.errorSink.record(err)

			return
		}

		request.Body = io.NopCloser(bytes.NewReader(body))

		h.record(entryMutationRequest{
			method:         request.Method,
			path:           request.URL.Path,
			version:        request.Header.Get("X-Contentful-Version"),
			versionPresent: len(request.Header.Values("X-Contentful-Version")) > 0,
			versionValues:  append([]string(nil), request.Header.Values("X-Contentful-Version")...),
			body:           append([]byte(nil), body...),
			contentLength:  request.ContentLength,
		})
	}

	h.delegate.ServeHTTP(responseWriter, request)
}

func (h *entryMutationRecorder) recordUpdate(request *http.Request) {
	if request.Method != http.MethodPut || !strings.HasSuffix(request.URL.Path, "/entries/entry") {
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		h.errorSink.record(err)

		return
	}

	request.Body = io.NopCloser(bytes.NewReader(body))

	var payload struct {
		Fields map[string]json.RawMessage `json:"fields"`
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		h.errorSink.record(err)

		return
	}

	h.record(entryMutationRequest{
		method:         request.Method,
		path:           request.URL.Path,
		version:        request.Header.Get("X-Contentful-Version"),
		versionPresent: len(request.Header.Values("X-Contentful-Version")) > 0,
		versionValues:  append([]string(nil), request.Header.Values("X-Contentful-Version")...),
		body:           append([]byte(nil), body...),
		contentLength:  request.ContentLength,
		contentType:    request.Header.Get("X-Contentful-Content-Type"),
		fields:         payload.Fields,
	})
}

func (h *entryMutationRecorder) record(request entryMutationRequest) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.requests = append(h.requests, request)
}

func (h *entryMutationRecorder) reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.requests = nil
}

func (h *entryMutationRecorder) snapshot() []entryMutationRequest {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]entryMutationRequest(nil), h.requests...)
}

func (h *entryMutationRecorder) handlerError() error {
	return h.errorSink.error()
}

func requireEntryUpdate(t *testing.T, request entryMutationRequest) {
	t.Helper()

	require.Equal(t, entryTestUpdatePath, request.path, "expected an Entry draft update")
	require.Equal(t, http.MethodPut, request.method)
	require.NotEmpty(t, request.body, "an Entry draft update must include its JSON request body")
	require.Positive(t, request.contentLength)
}

func requireEntryPublish(t *testing.T, request entryMutationRequest) {
	t.Helper()

	require.Equal(t, entryTestPublishPath, request.path, "expected an Entry publication")
	require.Equal(t, http.MethodPut, request.method)
	require.Equal(t, []string{request.version}, request.versionValues)
	require.Empty(t, request.body, "Entry Publish must have a zero-byte request body")
	require.Zero(t, request.contentLength)
}

func requireEntryUpdateThenPublish(
	t *testing.T,
	requests []entryMutationRequest,
) (entryMutationRequest, entryMutationRequest) {
	t.Helper()

	require.Len(t, requests, 2, "expected one Entry draft update followed by one publication")
	requireEntryUpdate(t, requests[0])
	requireEntryPublish(t, requests[1])

	return requests[0], requests[1]
}

func requireNoEntryMutations(t *testing.T, recorder *entryMutationRecorder, msgAndArgs ...any) {
	t.Helper()

	require.Empty(t, recorder.snapshot(), msgAndArgs...)
}
