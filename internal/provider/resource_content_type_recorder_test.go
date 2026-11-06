package provider_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
)

type contentTypeOperation uint8

const (
	contentTypeOperationGet contentTypeOperation = iota
	contentTypeOperationPut
	contentTypeOperationActivate
)

type contentTypeEvent struct {
	Operation      contentTypeOperation
	Method         string
	Path           string
	Version        int64
	VersionPresent bool
}

type contentTypeRawRequest struct {
	contentTypeEvent

	VersionValues []string
	Body          []byte
	ContentLength int64
}

func (h *contentTypeActivationTestHandler) recordEvent(operation contentTypeOperation, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		panic(fmt.Sprintf("read content type event body: %v", err))
	}

	request.Body = io.NopCloser(bytes.NewReader(body))

	version := int64(0)

	versionValues := append([]string(nil), request.Header.Values("X-Contentful-Version")...)
	versionPresent := len(versionValues) > 0

	if operation != contentTypeOperationGet {
		version = -1

		if versionPresent {
			parsedVersion, err := strconv.ParseInt(request.Header.Get("X-Contentful-Version"), 10, 64)
			if err == nil {
				version = parsedVersion
			}
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	event := contentTypeEvent{
		Operation: operation, Method: request.Method, Path: request.URL.Path, Version: version, VersionPresent: versionPresent,
	}
	h.events = append(h.events, event)
	h.rawRequests = append(h.rawRequests, contentTypeRawRequest{
		contentTypeEvent: event,
		VersionValues:    versionValues,
		Body:             append([]byte(nil), body...),
		ContentLength:    request.ContentLength,
	})
}

func (h *contentTypeActivationTestHandler) activationVersionHistory() []int64 {
	events := h.eventHistory()
	versions := make([]int64, 0, len(events))

	for _, event := range events {
		if event.Operation == contentTypeOperationActivate {
			versions = append(versions, event.Version)
		}
	}

	return versions
}

func (h *contentTypeActivationTestHandler) resetRequestHistory() {
	h.mu.Lock()
	h.events = nil
	h.rawRequests = nil
	h.putBodies = nil
	h.mu.Unlock()
	h.activationFailureReturned.Store(false)
}

func (h *contentTypeActivationTestHandler) recordPutBody(body []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.putBodies = append(h.putBodies, append([]byte(nil), body...))
}

func (h *contentTypeActivationTestHandler) lastPutBody() []byte {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.putBodies) == 0 {
		return nil
	}

	return append([]byte(nil), h.putBodies[len(h.putBodies)-1]...)
}

func (h *contentTypeActivationTestHandler) eventHistory() []contentTypeEvent {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]contentTypeEvent(nil), h.events...)
}

func (h *contentTypeActivationTestHandler) rawRequestHistory() []contentTypeRawRequest {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]contentTypeRawRequest(nil), h.rawRequests...)
}

func (h *contentTypeActivationTestHandler) eventCount(operation contentTypeOperation) int {
	count := 0

	for _, event := range h.eventHistory() {
		if event.Operation == operation {
			count++
		}
	}

	return count
}

func (h *contentTypeActivationTestHandler) lastVersion(operation contentTypeOperation) int64 {
	events := h.eventHistory()

	for _, event := range slices.Backward(events) {
		if event.Operation == operation {
			return event.Version
		}
	}

	return -1
}
