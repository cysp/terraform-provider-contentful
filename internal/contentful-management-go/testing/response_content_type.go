package cmtesting

// ResponseContentType identifies a response media type emitted by the test server.
type ResponseContentType string

const (
	// ResponseContentTypeApplicationJSON is the generic JSON media type.
	ResponseContentTypeApplicationJSON ResponseContentType = "application/json"
	// ResponseContentTypeContentfulJSON is Contentful's vendor JSON media type.
	ResponseContentTypeContentfulJSON ResponseContentType = "application/vnd.contentful.management.v1+json"
)
