package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"testing/synctest"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentStatusReadyDataSourceCancellationBetweenPolls(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		requestCount := 0
		responseBody := environmentStatusReadyTestResponseBody(t, "queued")
		client := environmentStatusReadyTestClient(t, roundTripFunc(func(request *http.Request) (*http.Response, error) {
			requestCount++

			return environmentStatusReadyTestResponse(request, io.NopCloser(strings.NewReader(responseBody))), nil
		}))

		ctx, cancel := context.WithCancel(t.Context())
		t.Cleanup(cancel)

		implementation, request, response := environmentStatusReadyTestRead(t, client)
		result := make(chan struct{})

		go func() {
			implementation.Read(ctx, request, &response)
			close(result)
		}()

		// Let Read consume the queued response and block on its polling timer.
		synctest.Wait()
		require.Equal(t, 1, requestCount)

		select {
		case <-result:
			t.Fatal("environment readiness read completed before cancellation")
		default:
		}

		cancel()
		synctest.Wait()

		select {
		case <-result:
		default:
			t.Fatal("environment readiness wait did not stop after cancellation")
		}

		require.Len(t, response.Diagnostics.Errors(), 1)
		assert.Equal(t, "Cancelled waiting for environment to become ready", response.Diagnostics.Errors()[0].Summary())
		assert.Equal(t, 1, requestCount)

		var state EnvironmentStatusReadyModel

		stateDiagnostics := response.State.Get(t.Context(), &state)
		require.False(t, stateDiagnostics.HasError(), stateDiagnostics)
		assert.Equal(t, types.StringValue("queued"), state.Status)
	})
}

func TestEnvironmentStatusReadyDataSourceCancellationDuringHTTPIO(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		releaseResponse := make(chan struct{})
		release := sync.OnceFunc(func() { close(releaseResponse) })
		t.Cleanup(release)

		requestCount := 0
		responseBody := environmentStatusReadyTestResponseBody(t, environmentStatusReadyValue)
		client := environmentStatusReadyTestClient(t, roundTripFunc(func(request *http.Request) (*http.Response, error) {
			requestCount++

			<-releaseResponse

			return environmentStatusReadyTestResponse(request, io.NopCloser(strings.NewReader(responseBody))), nil
		}))

		ctx, cancel := context.WithCancel(t.Context())
		t.Cleanup(cancel)

		implementation, request, response := environmentStatusReadyTestRead(t, client)
		result := make(chan struct{})

		go func() {
			implementation.Read(ctx, request, &response)
			close(result)
		}()

		// Keep the HTTP response in flight while cancellation arrives.
		synctest.Wait()
		require.Equal(t, 1, requestCount)

		cancel()
		release()
		synctest.Wait()

		select {
		case <-result:
		default:
			t.Fatal("environment readiness read did not stop after cancellation")
		}

		require.Len(t, response.Diagnostics.Errors(), 1)
		assert.Equal(t, "Cancelled waiting for environment to become ready", response.Diagnostics.Errors()[0].Summary())
		assert.Equal(t, 1, requestCount)
		assert.True(t, response.State.Raw.IsNull())
	})
}

func environmentStatusReadyTestClient(t *testing.T, transport http.RoundTripper) *cm.Client {
	t.Helper()

	client, err := cm.NewClient(
		"https://api.contentful.invalid",
		cm.NewAccessTokenSecuritySource(cmt.ValidAccessToken),
		cm.WithClient(&http.Client{Transport: transport}),
	)
	require.NoError(t, err)

	return client
}

func environmentStatusReadyTestResponse(
	request *http.Request,
	body io.ReadCloser,
) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body:    body,
		Request: request,
	}
}

func environmentStatusReadyTestResponseBody(t *testing.T, status string) string {
	t.Helper()

	environment := cm.Environment{
		Sys:  cm.NewEnvironmentSys("space", "environment", status),
		Name: "environment",
	}

	encoded, err := json.Marshal(environment)
	require.NoError(t, err)

	return string(encoded)
}

func environmentStatusReadyTestRead(
	t *testing.T,
	client *cm.Client,
) (environmentStatusReadyDataSource, datasource.ReadRequest, datasource.ReadResponse) {
	t.Helper()

	ctx := t.Context()
	schema := EnvironmentStatusReadyDataSourceSchema(ctx)
	model := environmentStatusReadyTestModel(
		types.StringNull(),
		environmentStatusReadyTimeoutValue(types.StringNull()),
	)
	plan := tfsdk.Plan{Schema: schema}
	require.False(t, plan.Set(ctx, &model).HasError())

	config := tfsdk.Config{Raw: plan.Raw, Schema: schema}
	response := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	implementation := environmentStatusReadyDataSource{providerData: ContentfulProviderData{client: client}}

	return implementation, datasource.ReadRequest{Config: config}, response
}
