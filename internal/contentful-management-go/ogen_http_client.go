package client

import (
	"net/http"

	ht "github.com/ogen-go/ogen/http"
)

type clientWithUserAgent struct {
	client    ht.Client
	userAgent string
}

var _ ht.Client = (*clientWithUserAgent)(nil)

//nolint:revive
func NewClientWithUserAgent(client ht.Client, userAgent string) *clientWithUserAgent {
	return &clientWithUserAgent{
		client:    client,
		userAgent: userAgent,
	}
}

func (c *clientWithUserAgent) Do(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" && c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	//nolint:wrapcheck
	return c.client.Do(req)
}
