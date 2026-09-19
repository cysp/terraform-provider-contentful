package provider

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const actionDiscoveryFixture = `{"sys":{"id":"action","type":"AppAction","organization":{"sys":{"type":"Link","linkType":"Organization","id":"org"}},"appDefinition":{"sys":{"type":"Link","linkType":"AppDefinition","id":"app"}}},"name":"Action","category":"Custom","type":"endpoint","url":"https://example.invalid/action","parameters":[]}`

func TestAppActionDiscoveryRejectsPartialResults(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		status int
		page   string
	}{
		{"later error", http.StatusForbidden, `{"sys":{"type":"Error","id":"AccessDenied"}}`},
		{"invalid response JSON", http.StatusOK, discoveryPage(1, 1, 2, strings.Replace(actionDiscoveryFixture, `"parameters":[]`, `"parameters":[invalid]`, 1))},
		{"wrong scope", http.StatusOK, discoveryPage(1, 1, 2, strings.Replace(strings.Replace(actionDiscoveryFixture, `"id":"action"`, `"id":"second"`, 1), `"id":"org"`, `"id":"wrong"`, 1))},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			calls := 0
			response := discoveryReadTest(t.Context(), t, NewAppActionsDataSource, map[string]any{"organization_id": "org", "app_definition_id": "app"}, contentfulRetryTestRoundTripper(func(req *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return discoveryHTTPResponse(req, http.StatusOK, discoveryPage(0, 1, 2, actionDiscoveryFixture)), nil
				}

				return discoveryHTTPResponse(req, test.status, test.page), nil
			}))
			require.True(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			assert.True(t, response.State.Raw.IsNull())
			assert.Equal(t, 2, calls)
		})
	}
}

func TestAppActionDiscoveryEmptyAndNotFound(t *testing.T) {
	t.Parallel()
	response := discoveryReadTest(t.Context(), t, NewAppActionsDataSource, map[string]any{"organization_id": "org", "app_definition_id": "app"}, contentfulRetryTestRoundTripper(func(req *http.Request) (*http.Response, error) {
		return discoveryHTTPResponse(req, http.StatusOK, discoveryPage(0, 100, 0, "")), nil
	}))
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)

	var data AppActionsDataSourceModel
	require.False(t, response.State.Get(t.Context(), &data).HasError())
	assert.NotNil(t, data.AppActions)
	assert.Empty(t, data.AppActions)
	response = discoveryReadTest(t.Context(), t, NewAppActionDataSource, map[string]any{"organization_id": "org", "app_definition_id": "app", "app_action_id": "action"}, contentfulRetryTestRoundTripper(func(req *http.Request) (*http.Response, error) {
		return discoveryHTTPResponse(req, http.StatusNotFound, `{"sys":{"type":"Error","id":"NotFound"}}`), nil
	}))
	require.True(t, response.Diagnostics.HasError())
	assert.True(t, response.State.Raw.IsNull())
}

func TestAppActionDiscoveryCancellation(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		factory func() datasource.DataSource
		inputs  map[string]any
		body    string
	}{
		{"singular", NewAppActionDataSource, map[string]any{"organization_id": "org", "app_definition_id": "app", "app_action_id": "action"}, actionDiscoveryFixture},
		{"final page", NewAppActionsDataSource, map[string]any{"organization_id": "org", "app_definition_id": "app"}, discoveryPage(0, 1, 1, actionDiscoveryFixture)},
		{"between pages", NewAppActionsDataSource, map[string]any{"organization_id": "org", "app_definition_id": "app"}, discoveryPage(0, 1, 2, actionDiscoveryFixture)},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			calls := 0
			response := discoveryReadTest(ctx, t, test.factory, test.inputs, contentfulRetryTestRoundTripper(func(req *http.Request) (*http.Response, error) {
				calls++

				cancel()

				return discoveryHTTPResponse(req, http.StatusOK, test.body), nil
			}))
			require.Len(t, response.Diagnostics, 1)
			require.True(t, response.Diagnostics.HasError())
			assert.Equal(t, context.Canceled.Error(), response.Diagnostics[0].Detail())
			assert.Equal(t, 1, calls)
			assert.True(t, response.State.Raw.IsNull())
		})
	}
}

func TestAppActionDiscoveryPreservesIrregularIDs(t *testing.T) {
	t.Parallel()

	for _, actionID := range []string{"", ".", "..", "a/b"} {
		t.Run(actionID, func(t *testing.T) {
			t.Parallel()

			body := strings.Replace(actionDiscoveryFixture, `"id":"action"`, `"id":"`+actionID+`"`, 1)
			response := discoveryReadTest(t.Context(), t, NewAppActionsDataSource, map[string]any{"organization_id": "org", "app_definition_id": "app"}, contentfulRetryTestRoundTripper(func(req *http.Request) (*http.Response, error) {
				return discoveryHTTPResponse(req, http.StatusOK, discoveryPage(0, 100, 1, body)), nil
			}))
			require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)

			var data AppActionsDataSourceModel
			require.False(t, response.State.Get(t.Context(), &data).HasError())
			require.Len(t, data.AppActions, 1)
			assert.Equal(t, actionID, data.AppActions[0].AppActionID.ValueString())
			lookup := discoveryReadTest(t.Context(), t, NewAppActionDataSource, map[string]any{"organization_id": "org", "app_definition_id": "app", "app_action_id": actionID}, contentfulRetryTestRoundTripper(func(req *http.Request) (*http.Response, error) {
				t.Error("invalid lookup must not send a request")

				return discoveryHTTPResponse(req, http.StatusOK, body), nil
			}))
			require.True(t, lookup.Diagnostics.HasError())
			assert.True(t, lookup.State.Raw.IsNull())
		})
	}
}
