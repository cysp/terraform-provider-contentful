package cmtesting_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentfulManagementServerRejectsMissingOrEmptyWebhookTopics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		want   string
	}{
		{
			name:   "create missing",
			method: http.MethodPost,
			path:   "/spaces/space/webhook_definitions",
			body:   `{"name":"Webhook","url":"https://example.com/webhook"}`,
			want: `{
				"sys":{"type":"Error","id":"ValidationFailed"},
				"message":"Validation error",
				"details":{"errors":[{
					"name":"invalid_type",
					"path":["topics"],
					"details":"Invalid input: expected array, received undefined at \"topics\""
				}]}
			}`,
		},
		{
			name:   "create empty",
			method: http.MethodPost,
			path:   "/spaces/space/webhook_definitions",
			body:   `{"name":"Webhook","url":"https://example.com/webhook","topics":[]}`,
			want: `{
				"sys":{"type":"Error","id":"ValidationFailed"},
				"message":"Validation error",
				"details":{"errors":[{
					"name":"topics",
					"path":[],
					"details":"Topics cannot be empty"
				}]}
			}`,
		},
		{
			name:   "update missing",
			method: http.MethodPut,
			path:   "/spaces/space/webhook_definitions/webhook",
			body:   `{"name":"Webhook","url":"https://example.com/webhook"}`,
			want: `{
				"sys":{"type":"Error","id":"ValidationFailed"},
				"message":"Validation error",
				"details":{"errors":[{
					"name":"invalid_type",
					"path":["topics"],
					"details":"Invalid input: expected array, received undefined at \"topics\""
				}]}
			}`,
		},
		{
			name:   "update empty",
			method: http.MethodPut,
			path:   "/spaces/space/webhook_definitions/webhook",
			body:   `{"name":"Webhook","url":"https://example.com/webhook","topics":[]}`,
			want: `{
				"sys":{"type":"Error","id":"ValidationFailed"},
				"message":"Validation error",
				"details":{"errors":[{
					"name":"topics",
					"path":[],
					"details":"Topics cannot be empty"
				}]}
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(100))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")

			if test.method == http.MethodPut {
				server.SetWebhookDefinition("space", "webhook", cm.WebhookDefinitionData{
					Name:   "Existing webhook",
					URL:    "https://example.com/webhook",
					Topics: []string{"Entry.publish"},
				})
			}

			testServer := httptest.NewServer(server)
			t.Cleanup(testServer.Close)

			request, err := http.NewRequestWithContext(t.Context(), test.method, testServer.URL+test.path, strings.NewReader(test.body))
			require.NoError(t, err)
			request.Header.Set("Authorization", "Bearer CFPAT-12345")
			request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")

			if test.method == http.MethodPut {
				request.Header.Set("X-Contentful-Version", "1")
			}

			response, err := testServer.Client().Do(request)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)

			assert.Equal(t, http.StatusUnprocessableEntity, response.StatusCode)
			assert.JSONEq(t, test.want, string(body))

			if test.method == http.MethodPut {
				stored, ok := server.StoredWebhookDefinition("space", "webhook")
				require.True(t, ok)
				assert.Equal(t, "Existing webhook", stored.Name)
				assert.Equal(t, []string{"Entry.publish"}, stored.Topics)
			}
		})
	}
}
