package provider_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
)

func deleteTaxonomyConceptRemote(server *cmt.Server, organizationID, conceptID string) error {
	ctx := context.Background()

	response, err := server.Handler().GetTaxonomyConcept(ctx, cm.GetTaxonomyConceptParams{
		OrganizationID: organizationID, TaxonomyConceptID: conceptID,
	})
	if err != nil {
		return fmt.Errorf("get taxonomy concept: %w", err)
	}

	concept, ok := response.(*cm.TaxonomyConcept)
	if !ok {
		return fmt.Errorf("%w: taxonomy concept: %T", errUnexpectedTaxonomyResponse, response)
	}

	deleted, err := server.Handler().DeleteTaxonomyConcept(ctx, cm.DeleteTaxonomyConceptParams{
		OrganizationID: organizationID, TaxonomyConceptID: conceptID, XContentfulVersion: concept.Sys.Version,
	})
	if err != nil {
		return fmt.Errorf("delete taxonomy concept: %w", err)
	}

	if _, ok := deleted.(*cm.NoContent); !ok {
		return fmt.Errorf("%w: taxonomy concept deletion: %T", errUnexpectedTaxonomyResponse, deleted)
	}

	return nil
}

func deleteTaxonomyConceptSchemeRemote(server *cmt.Server, organizationID, schemeID string) error {
	ctx := context.Background()

	response, err := server.Handler().GetTaxonomyConceptScheme(ctx, cm.GetTaxonomyConceptSchemeParams{
		OrganizationID: organizationID, TaxonomyConceptSchemeID: schemeID,
	})
	if err != nil {
		return fmt.Errorf("get taxonomy concept scheme: %w", err)
	}

	scheme, ok := response.(*cm.TaxonomyConceptScheme)
	if !ok {
		return fmt.Errorf("%w: taxonomy concept scheme: %T", errUnexpectedTaxonomyResponse, response)
	}

	deleted, err := server.Handler().DeleteTaxonomyConceptScheme(ctx, cm.DeleteTaxonomyConceptSchemeParams{
		OrganizationID: organizationID, TaxonomyConceptSchemeID: schemeID, XContentfulVersion: scheme.Sys.Version,
	})
	if err != nil {
		return fmt.Errorf("delete taxonomy concept scheme: %w", err)
	}

	if _, ok := deleted.(*cm.NoContent); !ok {
		return fmt.Errorf("%w: taxonomy concept scheme deletion: %T", errUnexpectedTaxonomyResponse, deleted)
	}

	return nil
}

type taxonomyRequestHook struct {
	next http.Handler

	mu     sync.Mutex
	method string
	path   string
	hook   func() error
	called bool
}

type taxonomyResponseFailure struct {
	next http.Handler

	mu               sync.Mutex
	method           string
	path             string
	remainingMatches int
	called           bool
}

func (f *taxonomyResponseFailure) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if f.consume(request.Method, request.URL.Path) {
		http.Error(responseWriter, "injected taxonomy failure", http.StatusBadRequest)

		return
	}

	f.next.ServeHTTP(responseWriter, request)
}

func (f *taxonomyResponseFailure) failOnce(method, path string, occurrence int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.method, f.path, f.remainingMatches, f.called = method, path, occurrence, false
}

func (f *taxonomyResponseFailure) consume(method, path string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	if method != f.method || !strings.HasSuffix(path, f.path) || f.called {
		return false
	}

	f.remainingMatches--
	if f.remainingMatches > 0 {
		return false
	}

	f.called = true

	return true
}

func (f *taxonomyResponseFailure) wasCalled() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.called
}

func (h *taxonomyRequestHook) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	hook := h.consume(request.Method, request.URL.Path)
	if hook != nil {
		err := hook()
		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusInternalServerError)

			return
		}
	}

	h.next.ServeHTTP(responseWriter, request)
}

func (h *taxonomyRequestHook) runOnce(method, path string, hook func() error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.method, h.path, h.hook, h.called = method, path, hook, false
}

func (h *taxonomyRequestHook) consume(method, path string) func() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if method != h.method || !strings.HasSuffix(path, h.path) {
		return nil
	}

	hook := h.hook
	h.method, h.path, h.hook, h.called = "", "", nil, true

	return hook
}

func (h *taxonomyRequestHook) wasCalled() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.called
}

type taxonomyResponseMutator struct {
	next http.Handler

	mu     sync.Mutex
	method string
	path   string
	locale string
	add    bool
}

func (m *taxonomyResponseMutator) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	recorder := httptest.NewRecorder()
	m.next.ServeHTTP(recorder, request)

	body := recorder.Body.Bytes()

	if recorder.Code >= http.StatusOK && recorder.Code < http.StatusMultipleChoices {
		locale, add, mutate := m.consumeMutation(request.Method, request.URL.Path)
		if mutate {
			if add {
				body = addEmptyLabelLocale(body, locale)
			} else {
				body = removePreferredLabelLocale(body, locale)
			}
		}
	}

	for key, values := range recorder.Header() {
		responseWriter.Header()[key] = append([]string(nil), values...)
	}

	responseWriter.Header().Del("Content-Length")
	responseWriter.WriteHeader(recorder.Code)

	_, _ = io.Copy(responseWriter, bytes.NewReader(body))
}

func (m *taxonomyResponseMutator) dropPreferredLabelOnce(method, path, locale string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.method, m.path, m.locale, m.add = method, path, locale, false
}

func (m *taxonomyResponseMutator) addEmptyLabelLocaleOnce(method, path, locale string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.method, m.path, m.locale, m.add = method, path, locale, true
}

func (m *taxonomyResponseMutator) consumeMutation(method, path string) (string, bool, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if method != m.method || !strings.Contains(path, m.path) {
		return "", false, false
	}

	locale := m.locale
	add := m.add
	m.method, m.path, m.locale, m.add = "", "", "", false

	return locale, add, true
}

func addEmptyLabelLocale(body []byte, locale string) []byte {
	document := map[string]json.RawMessage{}

	err := json.Unmarshal(body, &document)
	if err != nil {
		return body
	}

	for _, field := range []string{"altLabels", "hiddenLabels"} {
		labels := map[string][]string{}

		err = json.Unmarshal(document[field], &labels)
		if err != nil {
			return body
		}

		labels[locale] = []string{}

		document[field], err = json.Marshal(labels)
		if err != nil {
			return body
		}
	}

	result, err := json.Marshal(document)
	if err != nil {
		return body
	}

	return result
}

func removePreferredLabelLocale(body []byte, locale string) []byte {
	document := map[string]json.RawMessage{}

	err := json.Unmarshal(body, &document)
	if err != nil {
		return body
	}

	labels := map[string]string{}

	err = json.Unmarshal(document["prefLabel"], &labels)
	if err != nil {
		return body
	}

	delete(labels, locale)

	encodedLabels, err := json.Marshal(labels)
	if err != nil {
		return body
	}

	document["prefLabel"] = encodedLabels

	result, err := json.Marshal(document)
	if err != nil {
		return body
	}

	return result
}

func deleteTaxonomyConceptOutOfBand(t *testing.T, server *cmt.Server, organizationID, conceptID string) {
	t.Helper()

	err := deleteTaxonomyConceptRemote(server, organizationID, conceptID)
	if err != nil {
		t.Fatal(err)
	}
}

func deleteTaxonomyConceptSchemeOutOfBand(t *testing.T, server *cmt.Server, organizationID, schemeID string) {
	t.Helper()

	err := deleteTaxonomyConceptSchemeRemote(server, organizationID, schemeID)
	if err != nil {
		t.Fatal(err)
	}
}
