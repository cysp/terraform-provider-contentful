package provider_test

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func ContentfulProviderMockedResourceTest(t *testing.T, testserver *httptest.Server, testcase resource.TestCase) {
	t.Helper()

	contentfulProviderMockableResourceTest(t, testserver, true, testcase)
}

func ContentfulProviderMockableResourceTest(t *testing.T, testserver *httptest.Server, testcase resource.TestCase) {
	t.Helper()

	contentfulProviderMockableResourceTest(t, testserver, false, testcase)
}

func contentfulProviderMockableResourceTest(t *testing.T, testserver *httptest.Server, alwaysMock bool, testcase resource.TestCase) {
	t.Helper()

	switch {
	case alwaysMock || os.Getenv("TF_ACC_MOCKED") != "":
		if testcase.ProtoV6ProviderFactories != nil {
			t.Fatal("tc.ProtoV6ProviderFactories must be nil")
		}

		testcase.ProtoV6ProviderFactories = makeTestAccProtoV6ProviderFactories(ContentfulProviderOptionsWithHTTPTestServer(testserver)...)
		resource.UnitTest(t, testcase)

	default:
		if testcase.ProtoV6ProviderFactories == nil {
			testcase.ProtoV6ProviderFactories = testAccProtoV6ProviderFactories
		}

		resource.Test(t, testcase)
	}
}

func ContentfulProviderOptionsWithHTTPTestServer(testserver *httptest.Server) []provider.Option {
	if testserver == nil {
		return nil
	}

	return []provider.Option{
		provider.WithContentfulURL(testserver.URL),
		provider.WithHTTPClient(testserver.Client()),
		provider.WithAccessToken("12345"),
	}
}
