package contentfulmanagement

//go:generate go run github.com/ogen-go/ogen/cmd/ogen -target . -package contentfulmanagement -clean ./openapi/openapi.yml

const (
	// DefaultServerURL is the default URL of the server.
	DefaultServerURL = "https://api.contentful.com"

	// DefaultUserAgent is the default user agent.
	DefaultUserAgent = "contentful-management-go/0.1"
)

const (
	ErrorSysIDNotFound = "NotFound"
)
