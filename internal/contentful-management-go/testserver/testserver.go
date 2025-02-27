package testserver

import (
	"net/http"
	"net/http/httptest"
	"sync"
)

type ContentfulManagementTestServer struct {
	mu       *sync.Mutex
	Server   *httptest.Server
	ServeMux *http.ServeMux
}

func NewContentfulManagementTestServer() *ContentfulManagementTestServer {
	testserver := &ContentfulManagementTestServer{mu: &sync.Mutex{}}

	testserver.ServeMux = http.NewServeMux()

	testserver.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testserver.mu.Lock()
		defer testserver.mu.Unlock()

		testserver.ServeMux.ServeHTTP(w, r)
	}))

	return testserver
}

func (ts *ContentfulManagementTestServer) ResetServeMux() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.ServeMux = http.NewServeMux()
}
