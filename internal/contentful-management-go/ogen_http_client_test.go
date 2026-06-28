package contentfulmanagement_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/require"
)

type staticClient struct {
	resp *http.Response
}

func (c staticClient) Do(*http.Request) (*http.Response, error) {
	return c.resp, nil
}

type recordingClient struct {
	resp *http.Response
	req  *http.Request
}

func (c *recordingClient) Do(req *http.Request) (*http.Response, error) {
	c.req = req

	return c.resp, nil
}

func TestResponseContentTypeNormalizingClientNormalizesResponseContentType(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		contentType string
		want        string
	}{
		"vendor json": {
			contentType: "application/vnd.contentful.management.v1+json",
			want:        "application/json",
		},
		"vendor json with charset": {
			contentType: "application/vnd.contentful.management.v1+json; charset=utf-8",
			want:        "application/json; charset=utf-8",
		},
		"application json": {
			contentType: "application/json",
			want:        "application/json",
		},
		"other media type": {
			contentType: "text/plain",
			want:        "text/plain",
		},
		"empty": {
			contentType: "",
			want:        "",
		},
		"invalid": {
			contentType: "not a media type",
			want:        "not a media type",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &http.Response{
				Header: http.Header{},
				Body:   io.NopCloser(strings.NewReader("")),
			}
			resp.Header.Set("Content-Type", test.contentType)

			client := cm.NewResponseContentTypeNormalizingClient(staticClient{resp: resp})
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.contentful.com", nil)
			require.NoError(t, err)

			got, err := client.Do(req)
			require.NoError(t, err)

			defer func() {
				require.NoError(t, got.Body.Close())
			}()

			require.Equal(t, test.want, got.Header.Get("Content-Type"))
		})
	}
}

func TestResponseContentTypeNormalizingClientAllowsNilResponse(t *testing.T) {
	t.Parallel()

	client := cm.NewResponseContentTypeNormalizingClient(staticClient{})
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.contentful.com", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	if resp != nil && resp.Body != nil {
		defer func() {
			require.NoError(t, resp.Body.Close())
		}()
	}

	require.Nil(t, resp)
}

func TestUserAgentClientSetsDefaultUserAgent(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		Header: http.Header{},
		Body:   io.NopCloser(strings.NewReader("")),
	}
	inner := &recordingClient{resp: resp}
	client := cm.NewUserAgentClient(inner, "test-user-agent")
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.contentful.com", nil)
	require.NoError(t, err)

	got, err := client.Do(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, got.Body.Close())
	}()

	require.Equal(t, "test-user-agent", inner.req.Header.Get("User-Agent"))
}

func TestUserAgentClientPreservesExistingUserAgent(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		Header: http.Header{},
		Body:   io.NopCloser(strings.NewReader("")),
	}
	inner := &recordingClient{resp: resp}
	client := cm.NewUserAgentClient(inner, "test-user-agent")
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.contentful.com", nil)
	require.NoError(t, err)
	req.Header.Set("User-Agent", "existing-user-agent")

	got, err := client.Do(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, got.Body.Close())
	}()

	require.Equal(t, "existing-user-agent", inner.req.Header.Get("User-Agent"))
}

func TestTransportClientComposesTransportBehavior(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		Header: http.Header{},
		Body:   io.NopCloser(strings.NewReader("")),
	}
	resp.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")

	inner := &recordingClient{resp: resp}
	client := cm.NewTransportClient(inner, "test-user-agent")
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.contentful.com", nil)
	require.NoError(t, err)

	got, err := client.Do(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, got.Body.Close())
	}()

	require.Equal(t, "test-user-agent", inner.req.Header.Get("User-Agent"))
	require.Equal(t, "application/json", got.Header.Get("Content-Type"))
}
