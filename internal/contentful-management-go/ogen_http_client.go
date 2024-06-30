package client

import (
	"net/http"

	ht "github.com/ogen-go/ogen/http"
)

type clientWithUserAgent struct {
	baseClient ht.Client

	UserAgent string
}

var _ ht.Client = (*clientWithUserAgent)(nil)

func wrapClientWithUserAgent(client ht.Client, userAgent string) *clientWithUserAgent {
	return &clientWithUserAgent{
		baseClient: client,
		UserAgent:  userAgent,
	}
}

func (c *clientWithUserAgent) Do(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" && c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	//nolint:wrapcheck
	return c.baseClient.Do(req)
}
