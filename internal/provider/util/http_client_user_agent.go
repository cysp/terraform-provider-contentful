package util

import (
	"net/http"
)

type ClientWithUserAgent struct {
	client *http.Client

	UserAgent string
}

func NewClientWithUserAgent(client *http.Client, userAgent string) *ClientWithUserAgent {
	return &ClientWithUserAgent{
		client:    client,
		UserAgent: userAgent,
	}
}

func (c *ClientWithUserAgent) Do(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" && c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	//nolint:wrapcheck,gosec // request target is provided by trusted caller-owned request construction
	return c.client.Do(req)
}
