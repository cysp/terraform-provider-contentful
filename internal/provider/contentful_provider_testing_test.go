package provider_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

type resourceTestHandlerResult interface {
	handlerError() error
}

var liveAcceptanceTestMutex sync.Mutex

func parallelWhenMocked(t *testing.T) {
	t.Helper()

	if os.Getenv("TF_ACC_MOCKED") != "" {
		t.Parallel()
	}
}

func ContentfulProviderMockedResourceTest(t *testing.T, server http.Handler, testcase resource.TestCase) {
	t.Helper()

	contentfulProviderMockableResourceTest(t, server, true, testcase)
}

func ContentfulProviderMockedResourceTestWithFactoryCounter(
	t *testing.T,
	handler http.Handler,
	testcase resource.TestCase,
	factoryCalls *atomic.Int64,
) {
	t.Helper()

	if result, ok := handler.(resourceTestHandlerResult); ok {
		t.Cleanup(func() {
			require.NoError(t, result.handlerError())
		})
	}

	testserver := httptest.NewServer(handler)
	t.Cleanup(testserver.Close)

	baseFactory := makeTestAccProtoV6ProviderFactories(ContentfulProviderOptionsWithHTTPTestServer(testserver)...)["contentful"]
	testcase.ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"contentful": func() (tfprotov6.ProviderServer, error) {
			factoryCalls.Add(1)

			return baseFactory()
		},
	}
	resource.Test(t, testcase)
}

func ContentfulProviderMockableResourceTest(t *testing.T, server http.Handler, testcase resource.TestCase) {
	t.Helper()

	contentfulProviderMockableResourceTest(t, server, false, testcase)
}

func contentfulProviderMockableResourceTest(t *testing.T, handler http.Handler, alwaysMock bool, testcase resource.TestCase) {
	t.Helper()

	if result, ok := handler.(resourceTestHandlerResult); ok {
		t.Cleanup(func() {
			require.NoError(t, result.handlerError())
		})
	}

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
		// Live acceptance tests share one Contentful account and Space quota.
		// Serialize them because selected lifecycle mutations deliberately return
		// the first 429 instead of transparently replaying it. Mocked tests remain
		// parallel and independently exercise exact request counts.
		liveAcceptanceTestMutex.Lock()
		defer liveAcceptanceTestMutex.Unlock()

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
		WithAccessToken("CFPAT-12345"),
	}
}
