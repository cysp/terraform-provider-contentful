package provider

import (
	"io"
	"net/http"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Resource calls must set the policy themselves; the transport never tags their context.
func TestAppSigningSecretMutationsStopAfterUncertainFailureOrRedirect(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		for _, test := range []struct {
			name           string
			status         int
			transportError bool
			outerRedirect  bool
		}{
			{name: "server error", status: http.StatusInternalServerError},
			{name: "transport error", transportError: true},
			{name: "inner preserving redirect", status: http.StatusTemporaryRedirect},
			{name: "inner rewriting redirect", status: http.StatusSeeOther},
			{name: "outer preserving redirect", status: http.StatusTemporaryRedirect, outerRedirect: true},
			{name: "outer rewriting redirect", status: http.StatusSeeOther, outerRedirect: true},
		} {
			t.Run(method+"/"+test.name, func(t *testing.T) {
				t.Parallel()

				attempts, redirected := 0, 0

				base := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
					response := contentfulRetryTestResponse(request, http.StatusBadRequest)
					response.Header.Set("Content-Type", "application/json")

					if request.URL.Path == "/redirect" {
						redirected++

						return response, nil
					}

					attempts++

					assert.Equal(t, method, request.Method)
					assert.Equal(t, "/organizations/organization/app_definitions/app/signing_secret", request.URL.Path)

					if attempts == 1 {
						if test.transportError {
							return nil, errContentfulRetryTestConnectionLost
						}

						response.StatusCode = test.status
						response.Header.Set("Location", "/redirect")
					}
					// A replay gets a terminal 400, keeping the negative control bounded.
					return response, nil
				})}
				if test.outerRedirect {
					// Return the redirect from the inner client so the outer policy is exercised.
					base.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
				}

				httpClient, retryClient := contentfulRetryTestClient(t, base)
				removeContentfulRetryTestDelay(retryClient)

				client, err := cm.NewClient("https://contentful.invalid", cm.NewAccessTokenSecuritySource("test"), cm.WithClient(httpClient))
				require.NoError(t, err)

				implementation := &appSigningSecretResource{providerData: ContentfulProviderData{client: client}}

				state := appSigningSecretRetryTestState(t)
				if method == http.MethodPut {
					response := resource.CreateResponse{State: state}
					implementation.Create(t.Context(), resource.CreateRequest{Plan: tfsdk.Plan(state), Config: tfsdk.Config(state)}, &response)
					assert.True(t, response.Diagnostics.HasError(), response.Diagnostics)
				} else {
					response := resource.DeleteResponse{State: state}
					implementation.Delete(t.Context(), resource.DeleteRequest{State: state}, &response)
					assert.True(t, response.Diagnostics.HasError(), response.Diagnostics)
				}

				assert.Equal(t, 1, attempts)
				assert.Zero(t, redirected)
			})
		}
	}
}

func TestAppSigningSecretReadRetainsRetries(t *testing.T) {
	t.Parallel()

	attempts := 0
	base := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
		attempts++

		assert.Equal(t, http.MethodGet, request.Method)

		status := http.StatusBadRequest
		if attempts == 1 {
			status = http.StatusTooManyRequests
		}

		response := contentfulRetryTestResponse(request, status)
		response.Header.Set("Content-Type", "application/json")

		return response, nil
	})}
	httpClient, retryClient := contentfulRetryTestClient(t, base)
	removeContentfulRetryTestDelay(retryClient)

	client, err := cm.NewClient("https://contentful.invalid", cm.NewAccessTokenSecuritySource("test"), cm.WithClient(httpClient))
	require.NoError(t, err)

	implementation := &appSigningSecretResource{providerData: ContentfulProviderData{client: client}}
	state := appSigningSecretRetryTestState(t)
	response := resource.ReadResponse{State: state}
	implementation.Read(t.Context(), resource.ReadRequest{State: state}, &response)
	assert.True(t, response.Diagnostics.HasError(), response.Diagnostics)
	assert.Equal(t, 2, attempts)
}

func appSigningSecretRetryTestState(t *testing.T) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: AppSigningSecretResourceSchema(t.Context())}
	model := AppSigningSecretModel{
		IDIdentityModel:               NewIDIdentityModelFromMultipartID("organization", "app"),
		AppSigningSecretIdentityModel: AppSigningSecretIdentityModel{OrganizationID: types.StringValue("organization"), AppDefinitionID: types.StringValue("app")},
		Value:                         types.StringValue(strings.Repeat("a", 64)),
		Timeouts:                      TimeoutsNull(),
	}
	require.Empty(t, state.Set(t.Context(), &model))

	return state
}

func TestAppSigningSecretMutationRetriesRateLimit(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			var bodies []string

			base := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
				assert.Equal(t, method, request.Method)
				assert.Equal(t, "/organizations/organization/app_definitions/app/signing_secret", request.URL.Path)
				assert.Empty(t, request.Header.Get("X-Contentful-Version"))

				var body []byte

				if request.Body != nil {
					var err error

					body, err = io.ReadAll(request.Body)
					require.NoError(t, err)
				}

				bodies = append(bodies, string(body))

				response := contentfulRetryTestResponse(request, http.StatusNoContent)
				response.Header.Set("Content-Type", "application/json")

				if len(bodies) == 1 {
					response.StatusCode = http.StatusTooManyRequests
					response.Header.Set("X-Contentful-Ratelimit-Reset", "0")
					response.Body = io.NopCloser(strings.NewReader(testJSON(map[string]any{"sys": map[string]any{"type": "Error", "id": "RateLimitExceeded"}, "message": "rate limit"})))
				} else if method == http.MethodPut {
					response.StatusCode = http.StatusOK
					response.Body = io.NopCloser(strings.NewReader(testJSON(map[string]any{
						"sys": map[string]any{
							"type":          "AppSigningSecret",
							"organization":  map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "organization"}},
							"appDefinition": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "AppDefinition", "id": "app"}},
						},
						"redactedValue": "opaque",
					})))
				}

				return response, nil
			})}
			httpClient, retryClient := contentfulRetryTestClient(t, base)
			removeContentfulRetryTestDelay(retryClient)

			client, err := cm.NewClient("https://contentful.invalid", cm.NewAccessTokenSecuritySource("test"), cm.WithClient(httpClient))
			require.NoError(t, err)

			implementation := &appSigningSecretResource{providerData: ContentfulProviderData{client: client}}
			state := appSigningSecretRetryTestState(t)

			if method == http.MethodPut {
				response := resource.CreateResponse{State: state, Identity: appSigningSecretTestIdentity(t.Context())}
				implementation.Create(t.Context(), resource.CreateRequest{Plan: tfsdk.Plan(state), Config: tfsdk.Config(state)}, &response)
				require.False(t, response.Diagnostics.HasError(), response.Diagnostics)

				var actual AppSigningSecretModel
				require.False(t, response.State.Get(t.Context(), &actual).HasError())
				assert.Equal(t, strings.Repeat("a", 64), actual.Value.ValueString())
				assert.Equal(t, "organization/app", actual.ID.ValueString())

				wantBody := testJSON(map[string]any{"value": strings.Repeat("a", 64)})
				assert.Equal(t, []string{wantBody, wantBody}, bodies)
			} else {
				response := resource.DeleteResponse{State: state}
				implementation.Delete(t.Context(), resource.DeleteRequest{State: state}, &response)
				require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
				assert.Equal(t, []string{"", ""}, bodies)
			}
		})
	}
}
