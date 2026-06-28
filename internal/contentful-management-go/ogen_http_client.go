package contentfulmanagement

import (
	"mime"
	"net/http"

	ht "github.com/ogen-go/ogen/http"
)

const (
	contentTypeApplicationJSON = "application/json"
	contentTypeContentfulJSON  = "application/vnd.contentful.management.v1+json"
)

type userAgentClient struct {
	client    ht.Client
	userAgent string
}

type responseContentTypeNormalizingClient struct {
	client ht.Client
}

var (
	_ ht.Client = (*userAgentClient)(nil)
	_ ht.Client = (*responseContentTypeNormalizingClient)(nil)
)

// NewTransportClient wraps client with Contentful transport behavior.
//
// It sets the default User-Agent on outgoing requests and normalizes Contentful
// vendor JSON response Content-Type values to application/json before decoding.
//
//nolint:revive
func NewTransportClient(client ht.Client, userAgent string) *responseContentTypeNormalizingClient {
	return NewResponseContentTypeNormalizingClient(NewUserAgentClient(client, userAgent))
}

// NewUserAgentClient wraps client to set a default User-Agent on outgoing requests.
//
//nolint:revive
func NewUserAgentClient(client ht.Client, userAgent string) *userAgentClient {
	return &userAgentClient{
		client:    client,
		userAgent: userAgent,
	}
}

// NewResponseContentTypeNormalizingClient wraps client to normalize Contentful
// vendor JSON response Content-Type values to application/json.
//
//nolint:revive
func NewResponseContentTypeNormalizingClient(client ht.Client) *responseContentTypeNormalizingClient {
	return &responseContentTypeNormalizingClient{
		client: client,
	}
}

func (c *userAgentClient) Do(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" && c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	//nolint:wrapcheck
	return c.client.Do(req)
}

func (c *responseContentTypeNormalizingClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	normalizeResponseContentType(resp)

	//nolint:wrapcheck
	return resp, err
}

func normalizeResponseContentType(resp *http.Response) {
	if resp == nil {
		return
	}

	contentType := resp.Header.Get("Content-Type")

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != contentTypeContentfulJSON {
		return
	}

	resp.Header.Set("Content-Type", mime.FormatMediaType(contentTypeApplicationJSON, params))
}
