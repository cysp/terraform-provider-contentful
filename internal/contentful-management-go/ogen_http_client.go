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

type transportClient struct {
	client    ht.Client
	userAgent string
}

var _ ht.Client = (*transportClient)(nil)

// NewTransportClient wraps client with Contentful transport behavior.
//
// It sets the default User-Agent on outgoing requests and normalizes Contentful
// vendor JSON response Content-Type values to application/json before decoding.
//
//nolint:revive
func NewTransportClient(client ht.Client, userAgent string) *transportClient {
	return &transportClient{
		client:    client,
		userAgent: userAgent,
	}
}

func (c *transportClient) Do(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" && c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

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
