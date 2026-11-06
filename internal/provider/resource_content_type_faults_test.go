package provider_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

func removeContentTypePublishedVersion(response map[string]any) {
	delete(mustContentTypeResponseObject(response["sys"]), "publishedVersion")
}

func putExternalContentTypeDraft(request *http.Request, server *cmt.Server, contentTypeID, name, description string) error {
	currentResponse, err := server.Handler().GetContentType(request.Context(), cm.GetContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID,
	})
	if err != nil {
		return fmt.Errorf("get Content Type before racing request: %w", err)
	}

	current, ok := currentResponse.(*cm.ContentType)
	if !ok {
		return fmt.Errorf("%w: got %T while reading draft before race", errUnexpectedContentTypeResponseType, currentResponse)
	}

	wireCurrent, err := json.Marshal(current)
	if err != nil {
		return fmt.Errorf("encode draft before racing request: %w", err)
	}

	var racingDraft cm.ContentTypeRequestData

	err = json.Unmarshal(wireCurrent, &racingDraft)
	if err != nil {
		return fmt.Errorf("decode draft before racing request: %w", err)
	}

	racingDraft.Name = name
	racingDraft.Description = cm.NewOptNilString(description)

	response, err := server.Handler().PutContentType(request.Context(), &racingDraft, cm.PutContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID, XContentfulVersion: cm.NewOptInt(current.Sys.Version),
	})
	if err != nil {
		return fmt.Errorf("create racing external draft: %w", err)
	}

	if _, ok := response.(*cm.ContentTypeStatusCode); !ok {
		return fmt.Errorf("%w: got %T while creating racing external draft", errUnexpectedContentTypeResponseType, response)
	}

	return nil
}

func putIdenticalExternalContentTypeDraft(t *testing.T, server *cmt.Server, contentTypeID string, version int) {
	t.Helper()

	currentResponse, err := server.Handler().GetContentType(t.Context(), cm.GetContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID,
	})
	require.NoError(t, err)

	current, ok := currentResponse.(*cm.ContentType)
	require.True(t, ok)
	require.Equal(t, version, current.Sys.Version)

	wireCurrent, err := json.Marshal(current)
	require.NoError(t, err)

	var identicalDraft cm.ContentTypeRequestData
	require.NoError(t, json.Unmarshal(wireCurrent, &identicalDraft))

	response, err := server.Handler().PutContentType(t.Context(), &identicalDraft, cm.PutContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID, XContentfulVersion: cm.NewOptInt(version),
	})
	require.NoError(t, err)
	require.IsType(t, &cm.ContentTypeStatusCode{}, response)
}

func mustContentTypeResponseObject(value any) map[string]any {
	converted, ok := value.(map[string]any)
	if !ok {
		panic(fmt.Sprintf("unexpected injected content type response object %T", value))
	}

	return converted
}

func contentTypeResponseMutator(t *testing.T, mutate func(map[string]any)) func(string) string {
	t.Helper()

	return func(body string) string {
		var response map[string]any

		decodeErr := json.Unmarshal([]byte(body), &response)
		if decodeErr != nil {
			t.Errorf("decode injected content type response: %v", decodeErr)

			return body
		}

		mutate(response)

		mutated, err := json.Marshal(response)
		if err != nil {
			t.Errorf("encode injected content type response: %v", err)

			return body
		}

		return string(mutated)
	}
}

type contentTypeActivationTestHandler struct {
	delegate                            http.Handler
	failActivation                      atomic.Bool
	activationVersionMismatch           atomic.Bool
	rateLimitPut                        atomic.Bool
	rateLimitActivation                 atomic.Bool
	activationFailureAfterSuccessStatus atomic.Int64
	activationFailureReturned           atomic.Bool

	mu               sync.Mutex
	events           []contentTypeEvent
	rawRequests      []contentTypeRawRequest
	putBodies        [][]byte
	beforeActivation func(*http.Request)
	beforePut        func(*http.Request)
	// mutatePutResponse is deliberately an HTTP-test-only fault injector. It
	// models a CMA success response that contradicts the accepted request.
	mutatePutResponse        func(string) string
	mutateActivationResponse func(string) string
	mutateReadResponse       func(string) string
}

type contentTypeVersionOffsetAdapter struct {
	delegate          http.Handler
	offset            int
	jumpOnDraftUpdate int

	mu     sync.Mutex
	jumped bool
}

type contentTypeHigherPostActivationVersionAdapter struct {
	delegate http.Handler
	server   *cmt.Server
	shot     entryOneShot
}

func (h *contentTypeVersionOffsetAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	isContentType := strings.Contains(request.URL.Path, "/content_types/") &&
		!strings.HasSuffix(request.URL.Path, "/editor_interface")
	if !isContentType || (request.Method != http.MethodGet && request.Method != http.MethodPut) {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	h.mu.Lock()
	offset := h.offset
	h.mu.Unlock()

	versionValues := request.Header.Values("X-Contentful-Version")
	isDraftUpdate := request.Method == http.MethodPut && !strings.HasSuffix(request.URL.Path, "/published") && len(versionValues) == 1

	if len(versionValues) == 1 {
		version, err := strconv.Atoi(versionValues[0])
		if err != nil {
			panic(fmt.Sprintf("parse offset Content Type version: %v", err))
		}

		request.Header.Set("X-Contentful-Version", strconv.Itoa(version-offset))
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	if recorder.Code < http.StatusOK || recorder.Code >= http.StatusMultipleChoices {
		for name, values := range recorder.Header() {
			responseWriter.Header()[name] = append([]string(nil), values...)
		}

		responseWriter.WriteHeader(recorder.Code)
		_, _ = responseWriter.Write(recorder.Body.Bytes())

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
		panic(fmt.Sprintf("decode offset Content Type response: %v", err))
	}

	sys := mustContentTypeResponseObject(payload["sys"])
	for _, name := range []string{"version", "publishedVersion"} {
		if value, present := sys[name].(float64); present {
			sys[name] = int(value) + offset
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		panic(fmt.Sprintf("encode offset Content Type response: %v", err))
	}

	for name, values := range recorder.Header() {
		responseWriter.Header()[name] = append([]string(nil), values...)
	}

	responseWriter.WriteHeader(recorder.Code)
	_, _ = responseWriter.Write(body)
}

func (h *contentTypeHigherPostActivationVersionAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPut || !strings.HasSuffix(request.URL.Path, "/published") || !h.shot.take() {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	if recorder.Code < http.StatusOK || recorder.Code >= http.StatusMultipleChoices {
		for name, values := range recorder.Header() {
			responseWriter.Header()[name] = append([]string(nil), values...)
		}

		responseWriter.WriteHeader(recorder.Code)
		_, _ = responseWriter.Write(recorder.Body.Bytes())

		return
	}

	contentTypeID := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/spaces/space/environments/environment/content_types/"), "/published")

	currentResponse, err := h.server.Handler().GetContentType(request.Context(), cm.GetContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID,
	})
	if err != nil {
		panic(fmt.Sprintf("read Content Type after activation: %v", err))
	}

	current, ok := currentResponse.(*cm.ContentType)
	if !ok {
		panic(fmt.Sprintf("unexpected Content Type response after activation: %T", currentResponse))
	}

	for range 2 {
		wireCurrent, marshalErr := json.Marshal(current)
		if marshalErr != nil {
			panic(fmt.Sprintf("encode Content Type draft: %v", marshalErr))
		}

		var draft cm.ContentTypeRequestData

		unmarshalErr := json.Unmarshal(wireCurrent, &draft)
		if unmarshalErr != nil {
			panic(fmt.Sprintf("decode Content Type draft: %v", unmarshalErr))
		}

		putResponse, putErr := h.server.Handler().PutContentType(request.Context(), &draft, cm.PutContentTypeParams{
			SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID,
			XContentfulVersion: cm.NewOptInt(current.Sys.Version),
		})
		if putErr != nil {
			panic(fmt.Sprintf("write Content Type draft: %v", putErr))
		}

		status, statusOK := putResponse.(*cm.ContentTypeStatusCode)
		if !statusOK {
			panic(fmt.Sprintf("unexpected Content Type draft response: %T", putResponse))
		}

		current = &status.Response
	}

	body, err := json.Marshal(current)
	if err != nil {
		panic(fmt.Sprintf("encode high Content Type response: %v", err))
	}

	for name, values := range recorder.Header() {
		responseWriter.Header()[name] = append([]string(nil), values...)
	}

	responseWriter.WriteHeader(recorder.Code)
	_, _ = responseWriter.Write(body)
}

func (h *contentTypeActivationTestHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet && strings.Contains(request.URL.Path, "/content_types/") && h.activationFailureReturned.Load() {
		h.recordEvent(contentTypeOperationGet, request)
	}

	if request.Method == http.MethodGet && strings.Contains(request.URL.Path, "/content_types/") &&
		!strings.HasSuffix(request.URL.Path, "/editor_interface") {
		if mutateReadResponse := h.takeReadResponseMutator(); mutateReadResponse != nil {
			recorder := httptest.NewRecorder()
			h.delegate.ServeHTTP(recorder, request)

			for name, values := range recorder.Header() {
				responseWriter.Header()[name] = append([]string(nil), values...)
			}

			responseWriter.WriteHeader(recorder.Code)
			_, _ = responseWriter.Write([]byte(mutateReadResponse(recorder.Body.String())))

			return
		}
	}

	if request.Method == http.MethodPut && strings.HasSuffix(request.URL.Path, "/published") {
		h.serveActivation(responseWriter, request)

		return
	}

	if request.Method == http.MethodPut &&
		strings.Contains(request.URL.Path, "/content_types/") &&
		!strings.HasSuffix(request.URL.Path, "/editor_interface") {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			panic(fmt.Sprintf("read content type PUT body: %v", err))
		}

		request.Body = io.NopCloser(bytes.NewReader(body))
		h.recordPutBody(body)
		h.recordEvent(contentTypeOperationPut, request)

		if h.rateLimitPut.Swap(false) {
			responseWriter.Header().Set("Retry-After", "0")

			message := "Injected content type draft rate limit"
			_ = cmt.WriteContentfulManagementErrorResponse(responseWriter, http.StatusTooManyRequests, "RateLimitExceeded", &message, nil)

			return
		}

		if beforePut := h.takeBeforePut(); beforePut != nil {
			beforePut(request)
		}

		if mutatePutResponse := h.takePutResponseMutator(); mutatePutResponse != nil {
			recorder := httptest.NewRecorder()
			h.delegate.ServeHTTP(recorder, request)

			for name, values := range recorder.Header() {
				responseWriter.Header()[name] = append([]string(nil), values...)
			}

			responseWriter.WriteHeader(recorder.Code)
			_, _ = responseWriter.Write([]byte(mutatePutResponse(recorder.Body.String())))

			return
		}
	}

	h.delegate.ServeHTTP(responseWriter, request)
}

func (h *contentTypeActivationTestHandler) serveActivation(responseWriter http.ResponseWriter, request *http.Request) {
	h.recordEvent(contentTypeOperationActivate, request)

	if h.rateLimitActivation.Swap(false) {
		responseWriter.Header().Set("Retry-After", "0")

		message := "Injected content type activation rate limit"
		_ = cmt.WriteContentfulManagementErrorResponse(responseWriter, http.StatusTooManyRequests, "RateLimitExceeded", &message, nil)

		return
	}

	if beforeActivation := h.takeBeforeActivation(); beforeActivation != nil {
		beforeActivation(request)
	}

	if h.activationVersionMismatch.Swap(false) {
		message := "Injected initial content type activation VersionMismatch"
		_ = cmt.WriteContentfulManagementErrorResponse(responseWriter, http.StatusConflict, "VersionMismatch", &message, nil)

		return
	}

	if h.failActivation.Load() {
		h.activationFailureReturned.Store(true)

		message := "Injected content type activation failure"
		_ = cmt.WriteContentfulManagementErrorResponse(
			responseWriter,
			http.StatusUnprocessableEntity,
			"ValidationFailed",
			&message,
			nil,
		)

		return
	}

	mutateActivationResponse := h.takeActivationResponseMutator()
	if h.activationFailureAfterSuccessStatus.Load() == 0 && mutateActivationResponse == nil {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	if mutateActivationResponse != nil {
		for name, values := range recorder.Header() {
			responseWriter.Header()[name] = append([]string(nil), values...)
		}

		responseWriter.WriteHeader(recorder.Code)
		_, _ = responseWriter.Write([]byte(mutateActivationResponse(recorder.Body.String())))

		return
	}

	if recorder.Code >= http.StatusOK && recorder.Code < http.StatusMultipleChoices {
		failureStatus := int(h.activationFailureAfterSuccessStatus.Swap(0))
		if failureStatus != 0 {
			h.activationFailureReturned.Store(true)

			message := "Injected content type activation response failure"
			_ = cmt.WriteContentfulManagementErrorResponse(
				responseWriter,
				failureStatus,
				"ServerError",
				&message,
				nil,
			)

			return
		}
	}

	for name, values := range recorder.Header() {
		responseWriter.Header()[name] = append([]string(nil), values...)
	}

	responseWriter.WriteHeader(recorder.Code)
	_, _ = responseWriter.Write(recorder.Body.Bytes())
}

func (h *contentTypeActivationTestHandler) takeBeforePut() func(*http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	hook := h.beforePut
	h.beforePut = nil

	return hook
}

func (h *contentTypeActivationTestHandler) takeBeforeActivation() func(*http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	hook := h.beforeActivation
	h.beforeActivation = nil

	return hook
}

func (h *contentTypeActivationTestHandler) takePutResponseMutator() func(string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	mutator := h.mutatePutResponse
	h.mutatePutResponse = nil

	return mutator
}

func (h *contentTypeActivationTestHandler) takeActivationResponseMutator() func(string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	mutator := h.mutateActivationResponse
	h.mutateActivationResponse = nil

	return mutator
}

func (h *contentTypeActivationTestHandler) takeReadResponseMutator() func(string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	mutator := h.mutateReadResponse
	h.mutateReadResponse = nil

	return mutator
}
