package provider_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
)

var (
	errEntryPublishResponseSysObject = errors.New("publish response sys is not an object")
	errEntryPublishResponseFields    = errors.New("publish response omitted fields")
	errUnexpectedEntryResponseType   = errors.New("unexpected entry response type")
)

type entryCreateDefaultingAdapter struct {
	delegate  http.Handler
	errorSink *entryFixtureErrorSink

	mu      sync.Mutex
	applied bool
}

// entryCommittedUpdateFailureAdapter models a draft PUT that commits remotely
// while the caller receives only an ambiguous server failure.
type entryCommittedUpdateFailureAdapter struct {
	delegate http.Handler
	shot     entryOneShot
}

func (h *entryCommittedUpdateFailureAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPut || !strings.HasSuffix(request.URL.Path, "/entries/entry") {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	if !h.shot.take() {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	h.delegate.ServeHTTP(httptest.NewRecorder(), request)
	http.Error(responseWriter, "injected ambiguous update response", http.StatusInternalServerError)
}

func (h *entryCommittedUpdateFailureAdapter) failAfterCommitOnce() {
	h.shot.arm()
}

type entryOneShot struct {
	mu    sync.Mutex
	armed bool
}

func (s *entryOneShot) arm() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.armed = true
}

func (s *entryOneShot) take() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	armed := s.armed
	s.armed = false

	return armed
}

func (h *entryCreateDefaultingAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPut || request.URL.Path != "/spaces/space/environments/environment/entries/entry" {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	h.mu.Lock()
	applyDefault := !h.applied

	if applyDefault {
		h.applied = true
	}
	h.mu.Unlock()

	if !applyDefault {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		h.errorSink.record(err)

		request.Body = io.NopCloser(bytes.NewReader(body))
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	originalBody := body

	var payload map[string]json.RawMessage

	err = json.Unmarshal(body, &payload)
	if err != nil {
		h.errorSink.record(err)

		request.Body = io.NopCloser(bytes.NewReader(originalBody))
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	var fields map[string]json.RawMessage

	err = json.Unmarshal(payload["fields"], &fields)
	if err != nil {
		h.errorSink.record(err)

		request.Body = io.NopCloser(bytes.NewReader(originalBody))
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	fields["defaulted"] = json.RawMessage(`{"en-US":"content-type default"}`)

	payload["fields"], err = json.Marshal(fields)
	if err != nil {
		h.errorSink.record(err)

		request.Body = io.NopCloser(bytes.NewReader(originalBody))
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	body, err = json.Marshal(payload)
	if err != nil {
		h.errorSink.record(err)

		request.Body = io.NopCloser(bytes.NewReader(originalBody))
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	request.Body = io.NopCloser(bytes.NewReader(body))
	request.ContentLength = int64(len(body))

	h.delegate.ServeHTTP(responseWriter, request)
}

type entryRejectedPublishAdapter struct {
	delegate http.Handler
	shot     entryOneShot
}

type entryVersionMismatchPublishAdapter struct {
	delegate http.Handler
	shot     entryOneShot
}

type entryUpdateVersionMismatchAdapter struct {
	delegate  http.Handler
	server    *cmt.Server
	shot      entryOneShot
	errorSink *entryFixtureErrorSink
}

type entryRateLimitAdapter struct {
	delegate http.Handler
	path     string
	shot     entryOneShot
}

func (h *entryRejectedPublishAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPut && strings.HasPrefix(request.URL.Path, entryTestCollectionPath+"/") &&
		strings.HasSuffix(request.URL.Path, "/published") && h.shot.take() {
		message := "injected publish failure"
		_ = cmt.WriteContentfulManagementErrorResponse(responseWriter, http.StatusBadRequest, "BadRequest", &message, nil)

		return
	}

	h.delegate.ServeHTTP(responseWriter, request)
}

func (h *entryVersionMismatchPublishAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPut && strings.HasSuffix(request.URL.Path, "/entries/entry/published") && h.shot.take() {
		message := "injected initial publication VersionMismatch"
		_ = cmt.WriteContentfulManagementErrorResponse(responseWriter, http.StatusConflict, "VersionMismatch", &message, nil)

		return
	}

	h.delegate.ServeHTTP(responseWriter, request)
}

func (h *entryUpdateVersionMismatchAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPut || request.URL.Path != entryTestUpdatePath || !h.shot.take() {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	entry, err := getEntryFromTestServer(request.Context(), h.server)
	if err == nil {
		_, err = h.server.Handler().PutEntry(request.Context(), &cm.EntryRequest{
			Fields: entry.Fields, Metadata: entry.Metadata,
		}, cm.PutEntryParams{
			SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
			XContentfulVersion: cm.NewOptInt(entry.Sys.Version),
		})
	}

	if err != nil {
		h.errorSink.record(err)
	}

	h.delegate.ServeHTTP(responseWriter, request)
}

func (h *entryRateLimitAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPut && request.URL.Path == h.path && h.shot.take() {
		message := "injected rate limit"
		_ = cmt.WriteContentfulManagementErrorResponse(responseWriter, http.StatusTooManyRequests, "RateLimitExceeded", &message, nil)

		return
	}

	h.delegate.ServeHTTP(responseWriter, request)
}

type entryCommittedPublishFailureAdapter struct {
	delegate http.Handler
	shot     entryOneShot
}

func (h *entryCommittedPublishFailureAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPut && strings.HasSuffix(request.URL.Path, "/entries/entry/published") && h.shot.take() {
		h.delegate.ServeHTTP(httptest.NewRecorder(), request)
		http.Error(responseWriter, "injected ambiguous publish response", http.StatusInternalServerError)

		return
	}

	h.delegate.ServeHTTP(responseWriter, request)
}

type entryAdditionalUpdateFieldAdapter struct {
	delegate  http.Handler
	shot      entryOneShot
	errorSink *entryFixtureErrorSink
}

func (h *entryAdditionalUpdateFieldAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPut || !strings.HasSuffix(request.URL.Path, "/entries/entry") || !h.shot.take() {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	var responsePayload map[string]json.RawMessage

	err := json.Unmarshal(recorder.Body.Bytes(), &responsePayload)
	if err != nil {
		h.errorSink.record(err)
		replayEntryAdapterResponse(responseWriter, recorder, h.errorSink)

		return
	}

	var responseFields map[string]json.RawMessage

	err = json.Unmarshal(responsePayload["fields"], &responseFields)
	if err != nil {
		h.errorSink.record(err)
		replayEntryAdapterResponse(responseWriter, recorder, h.errorSink)

		return
	}

	responseFields["response-only"] = json.RawMessage(`{"en-US":"unexpected"}`)

	fields, err := json.Marshal(responseFields)
	if err != nil {
		h.errorSink.record(err)
		replayEntryAdapterResponse(responseWriter, recorder, h.errorSink)

		return
	}

	responsePayload["fields"] = fields
	writeEntryAdapterJSONResponse(responseWriter, recorder, responsePayload, h.errorSink)
}

type entryPublishTupleAdapter struct {
	delegate         http.Handler
	shot             entryOneShot
	version          *int
	publishedVersion *int
	errorSink        *entryFixtureErrorSink
}

func (h *entryPublishTupleAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPut || !strings.HasSuffix(request.URL.Path, "/entries/entry/published") || !h.shot.take() {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	var payload map[string]any

	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	if err != nil {
		h.errorSink.record(err)
		replayEntryAdapterResponse(responseWriter, recorder, h.errorSink)

		return
	}

	sys, ok := payload["sys"].(map[string]any)
	if !ok {
		h.errorSink.record(errEntryPublishResponseSysObject)
		replayEntryAdapterResponse(responseWriter, recorder, h.errorSink)

		return
	}

	if h.version != nil {
		sys["version"] = *h.version
	}

	if h.publishedVersion != nil {
		sys["publishedVersion"] = *h.publishedVersion
	}

	writeEntryAdapterJSONResponse(responseWriter, recorder, payload, h.errorSink)
}

// entryHigherPostPublishVersionAdapter models a successful publication response
// whose Contentful sys.version is followed by two external draft writes before
// the response reaches Terraform.
type entryHigherPostPublishVersionAdapter struct {
	delegate  http.Handler
	server    *cmt.Server
	shot      entryOneShot
	errorSink *entryFixtureErrorSink
}

// entryVersionOffsetAdapter is a stateful HTTP-only fault model for a coherent
// CMA resource whose initial positive sys.version is greater than one. It maps
// optimistic-lock headers to the normal fake server while preserving the
// offset in every returned lifecycle tuple.
type entryVersionOffsetAdapter struct {
	delegate          http.Handler
	offset            int
	jumpOnDraftUpdate int
	errorSink         *entryFixtureErrorSink

	mu     sync.Mutex
	jumped bool
}

// entryAdditionalPublishFieldAdapter injects one response-only field into one
// otherwise successful publication response.
type entryAdditionalPublishFieldAdapter struct {
	delegate  http.Handler
	shot      entryOneShot
	errorSink *entryFixtureErrorSink
}

// entryReorderedMetadataReadAdapter is a named adversarial CMA mode based on a
// live observation: a later GET can return unchanged tags in a different order
// from the mutation response.
type entryReorderedMetadataReadAdapter struct {
	delegate  http.Handler
	errorSink *entryFixtureErrorSink
}

type entryPublicationTupleReadAdapter struct {
	delegate         http.Handler
	shot             entryOneShot
	version          *int
	publishedVersion *int
	errorSink        *entryFixtureErrorSink
}

func (h *entryReorderedMetadataReadAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet || !strings.HasSuffix(request.URL.Path, "/entries/entry") {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	var payload map[string]any

	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	if err != nil {
		h.errorSink.record(err)
		replayEntryAdapterResponse(responseWriter, recorder, h.errorSink)

		return
	}

	if metadata, ok := payload["metadata"].(map[string]any); ok {
		if tags, ok := metadata["tags"].([]any); ok {
			slices.Reverse(tags)
		}
	}

	writeEntryAdapterJSONResponse(responseWriter, recorder, payload, h.errorSink)
}

func (h *entryPublicationTupleReadAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet || !strings.HasSuffix(request.URL.Path, "/entries/entry") || !h.shot.take() {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	var payload map[string]any

	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	if err != nil {
		h.errorSink.record(err)
		replayEntryAdapterResponse(responseWriter, recorder, h.errorSink)

		return
	}

	sys, ok := payload["sys"].(map[string]any)
	if !ok {
		h.errorSink.record(errEntryPublishResponseSysObject)
		replayEntryAdapterResponse(responseWriter, recorder, h.errorSink)

		return
	}

	if h.version != nil {
		sys["version"] = *h.version
	}

	if h.publishedVersion == nil {
		delete(sys, "publishedVersion")
	} else {
		sys["publishedVersion"] = *h.publishedVersion
	}

	writeEntryAdapterJSONResponse(responseWriter, recorder, payload, h.errorSink)
}

func (h *entryHigherPostPublishVersionAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPut || !strings.HasSuffix(request.URL.Path, "/entries/entry/published") || !h.shot.take() {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	entry, err := getEntryFromTestServer(request.Context(), h.server)
	if err == nil {
		for range 2 {
			response, putErr := h.server.Handler().PutEntry(request.Context(), &cm.EntryRequest{
				Fields: entry.Fields, Metadata: entry.Metadata,
			}, cm.PutEntryParams{
				SpaceID: "space", EnvironmentID: "environment", EntryID: "entry", XContentfulVersion: cm.NewOptInt(entry.Sys.Version),
			})
			if putErr != nil {
				err = putErr

				break
			}

			statusResponse, ok := response.(*cm.EntryStatusCode)
			if !ok {
				err = fmt.Errorf("%w: %T", errUnexpectedEntryResponseType, response)

				break
			}

			entry = &statusResponse.Response
		}
	}

	if err != nil {
		h.errorSink.record(err)
		replayEntryAdapterResponse(responseWriter, recorder, h.errorSink)

		return
	}

	writeEntryAdapterResponse(responseWriter, recorder, entry, h.errorSink)
}

func (h *entryVersionOffsetAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	isEntry := strings.HasSuffix(request.URL.Path, "/entries/entry") ||
		strings.HasSuffix(request.URL.Path, "/entries/entry/published")
	if !isEntry || (request.Method != http.MethodGet && request.Method != http.MethodPut) {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	h.mu.Lock()
	offset := h.offset
	h.mu.Unlock()

	versionValues := request.Header.Values("X-Contentful-Version")
	isDraftUpdate := request.URL.Path == entryTestUpdatePath && request.Method == http.MethodPut && len(versionValues) == 1

	if len(versionValues) == 1 {
		version, err := strconv.Atoi(versionValues[0])
		if err != nil {
			h.errorSink.record(err)
		} else {
			request.Header.Set("X-Contentful-Version", strconv.Itoa(version-offset))
		}
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	if recorder.Code < http.StatusOK || recorder.Code >= http.StatusMultipleChoices {
		replayEntryAdapterResponse(responseWriter, recorder, h.errorSink)

		return
	}

	h.mu.Lock()
	if isDraftUpdate && h.jumpOnDraftUpdate > 0 && !h.jumped {
		h.offset += h.jumpOnDraftUpdate
		h.jumped = true
	}

	offset = h.offset
	h.mu.Unlock()

	var payload map[string]any

	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	if err != nil {
		h.errorSink.record(err)
		replayEntryAdapterResponse(responseWriter, recorder, h.errorSink)

		return
	}

	sys, ok := payload["sys"].(map[string]any)
	if !ok {
		h.errorSink.record(errEntryPublishResponseSysObject)
		replayEntryAdapterResponse(responseWriter, recorder, h.errorSink)

		return
	}

	for _, name := range []string{"version", "publishedVersion"} {
		if value, present := sys[name].(float64); present {
			sys[name] = int(value) + offset
		}
	}

	writeEntryAdapterJSONResponse(responseWriter, recorder, payload, h.errorSink)
}

func (h *entryAdditionalPublishFieldAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPut || !strings.HasSuffix(request.URL.Path, "/entries/entry/published") || !h.shot.take() {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	var entry cm.Entry

	err := json.Unmarshal(recorder.Body.Bytes(), &entry)
	if err != nil {
		h.errorSink.record(err)
		replayEntryAdapterResponse(responseWriter, recorder, h.errorSink)

		return
	}

	responseFields, ok := entry.Fields.Get()
	if !ok {
		h.errorSink.record(errEntryPublishResponseFields)
		replayEntryAdapterResponse(responseWriter, recorder, h.errorSink)

		return
	}

	fields := maps.Clone(responseFields)
	fields["response-only"] = jx.Raw(`{"en-US":"unexpected"}`)
	entry.Fields = cm.NewOptEntryFields(fields)

	writeEntryAdapterResponse(responseWriter, recorder, &entry, h.errorSink)
}

func writeEntryAdapterResponse(
	responseWriter http.ResponseWriter,
	recorder *httptest.ResponseRecorder,
	entry *cm.Entry,
	errorSink *entryFixtureErrorSink,
) {
	maps.Copy(responseWriter.Header(), recorder.Header())
	responseWriter.WriteHeader(recorder.Code)

	encoder := new(jx.Encoder)
	entry.Encode(encoder)

	_, err := encoder.WriteTo(responseWriter)
	if err != nil {
		errorSink.record(err)
	}
}

func writeEntryAdapterJSONResponse(
	responseWriter http.ResponseWriter,
	recorder *httptest.ResponseRecorder,
	payload any,
	errorSink *entryFixtureErrorSink,
) {
	maps.Copy(responseWriter.Header(), recorder.Header())
	responseWriter.WriteHeader(recorder.Code)

	err := json.NewEncoder(responseWriter).Encode(payload)
	if err != nil {
		errorSink.record(err)
	}
}

func replayEntryAdapterResponse(
	responseWriter http.ResponseWriter,
	recorder *httptest.ResponseRecorder,
	errorSink *entryFixtureErrorSink,
) {
	maps.Copy(responseWriter.Header(), recorder.Header())
	responseWriter.WriteHeader(recorder.Code)

	_, err := responseWriter.Write(recorder.Body.Bytes())
	errorSink.record(err)
}

func TestEntryFaultAdapterReportsFixtureErrorOutsideHandlerGoroutine(t *testing.T) {
	t.Parallel()

	invalidResponse := http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.WriteHeader(http.StatusOK)
		_, _ = responseWriter.Write([]byte(`{"sys":`))
	})
	tupleFault := &entryPublishTupleAdapter{delegate: invalidResponse}
	tupleFault.shot.arm()

	errorSink := new(entryFixtureErrorSink)
	tupleFault.errorSink = errorSink
	recorder := newEntryMutationRecorder(tupleFault, errorSink)
	testServer := httptest.NewServer(recorder)
	t.Cleanup(testServer.Close)

	request, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPut,
		testServer.URL+"/spaces/space/environments/environment/entries/entry/published",
		nil,
	)
	require.NoError(t, err)

	response, err := testServer.Client().Do(request)
	require.NoError(t, err, "a fixture error must not be converted into a transport failure")
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Error(t, recorder.handlerError())
}
