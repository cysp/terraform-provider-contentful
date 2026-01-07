package integration_tests_test

import (
	"net/http/httptest"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

func testContentfulManagementHTTPTestServer(t *testing.T) (*cmt.Server, *httptest.Server) {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	testserver := httptest.NewServer(server)

	return server, testserver
}
