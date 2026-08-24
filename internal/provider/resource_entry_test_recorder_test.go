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
	path    string
	version string
	fields  map[string]json.RawMessage
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
		h.record(entryMutationRequest{path: request.URL.Path, version: request.Header.Get("X-Contentful-Version")})
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
		path: request.URL.Path, version: request.Header.Get("X-Contentful-Version"), fields: payload.Fields,
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
}

func requireEntryPublish(t *testing.T, request entryMutationRequest) {
	t.Helper()

	require.Equal(t, entryTestPublishPath, request.path, "expected an Entry publication")
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
