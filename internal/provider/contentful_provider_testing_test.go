package provider_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func ContentfulProviderMockedResourceTest(t *testing.T, server http.Handler, testcase resource.TestCase) {
	t.Helper()

	contentfulProviderMockableResourceTest(t, server, true, testcase)
}

func ContentfulProviderMockableResourceTest(t *testing.T, server http.Handler, testcase resource.TestCase) {
	t.Helper()

	contentfulProviderMockableResourceTest(t, server, false, testcase)
}

func contentfulProviderMockableResourceTest(t *testing.T, handler http.Handler, alwaysMock bool, testcase resource.TestCase) {
	t.Helper()

	switch {
	case alwaysMock || os.Getenv("TF_ACC_MOCKED") != "":
		if testcase.ProtoV6ProviderFactories != nil {
			t.Fatal("tc.ProtoV6ProviderFactories must be nil")
		}

		var testserver *httptest.Server
		if handler != nil {
			testserver = httptest.NewServer(handler)
			t.Cleanup(testserver.Close)
		}

		testcase.ProtoV6ProviderFactories = makeTestAccProtoV6ProviderFactories(ContentfulProviderOptionsWithHTTPTestServer(testserver)...)
		resource.Test(t, testcase)

	default:
		if testcase.ProtoV6ProviderFactories == nil {
			testcase.ProtoV6ProviderFactories = testAccProtoV6ProviderFactories
		}

		resource.Test(t, testcase)
	}
}

func ContentfulProviderOptionsWithHTTPTestServer(testserver *httptest.Server) []Option {
	if testserver == nil {
		return nil
	}

	return []Option{
		WithContentfulURL(testserver.URL),
		WithHTTPClient(testserver.Client()),
		WithAccessToken("12345"),
	}
}
