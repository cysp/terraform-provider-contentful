package contentfulmanagement_test

import (
	"io"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type transportClientStub struct {
	response *http.Response
	err      error
	request  *http.Request
}

func (c *transportClientStub) Do(request *http.Request) (*http.Response, error) {
	c.request = request

	return c.response, c.err
}

func TestTransportClientNormalizesResponseContentType(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		contentType string
		want        string
	}{
		"vendor JSON":              {contentType: "application/vnd.contentful.management.v1+json", want: "application/json"},
		"vendor JSON with charset": {contentType: "application/vnd.contentful.management.v1+json; charset=utf-8", want: "application/json; charset=utf-8"},
		"application JSON":         {contentType: "application/json", want: "application/json"},
		"other media type":         {contentType: "text/plain", want: "text/plain"},
		"empty":                    {},
		"invalid":                  {contentType: "not a media type", want: "not a media type"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &http.Response{Header: http.Header{"Content-Type": []string{test.contentType}}}
			client := cm.NewTransportClient(&transportClientStub{response: response}, "test-user-agent")
			request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.contentful.com", nil)
			require.NoError(t, err)

			got, err := client.Do(request) //nolint:bodyclose // Stub responses have no body.
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, test.want, got.Header.Get("Content-Type"))
		})
	}
}

func TestTransportClientComposesUserAgentAndResponseNormalization(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		defaultUserAgent string
		requestUserAgent string
		want             string
	}{
		"default":       {defaultUserAgent: "test-user-agent", want: "test-user-agent"},
		"existing":      {defaultUserAgent: "test-user-agent", requestUserAgent: "existing-user-agent", want: "existing-user-agent"},
		"empty default": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &http.Response{Header: http.Header{"Content-Type": []string{"application/vnd.contentful.management.v1+json"}}}
			inner := &transportClientStub{response: response}
			client := cm.NewTransportClient(inner, test.defaultUserAgent)
			request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.contentful.com", nil)
			require.NoError(t, err)

			if test.requestUserAgent != "" {
				request.Header.Set("User-Agent", test.requestUserAgent)
			}

			got, err := client.Do(request) //nolint:bodyclose // Stub responses have no body.
			require.NoError(t, err)
			require.NotNil(t, inner.request)
			assert.Equal(t, test.want, inner.request.Header.Get("User-Agent"))
			require.NotNil(t, got)
			assert.Equal(t, "application/json", got.Header.Get("Content-Type"))
		})
	}
}

func TestTransportClientPreservesResponseAndError(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		response *http.Response
		err      error
	}{
		"nil response":           {},
		"error without response": {err: io.ErrUnexpectedEOF},
		"error with response": {
			response: &http.Response{Header: http.Header{"Content-Type": []string{"application/vnd.contentful.management.v1+json"}}},
			err:      io.ErrUnexpectedEOF,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := cm.NewTransportClient(&transportClientStub{response: test.response, err: test.err}, "test-user-agent")
			request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.contentful.com", nil)
			require.NoError(t, err)

			got, err := client.Do(request) //nolint:bodyclose // Stub responses have no body.
			if test.err == nil {
				require.NoError(t, err)
			} else {
				assert.Same(t, test.err, err)
			}

			if test.response == nil {
				assert.Nil(t, got)
			} else {
				require.Same(t, test.response, got)
				assert.Equal(t, "application/json", got.Header.Get("Content-Type"))
			}
		})
	}
}
