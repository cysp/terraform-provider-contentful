package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"

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

	firstResponseClosed := make(chan string, 1)

	requestCount := 0

	responseBody := environmentStatusReadyTestResponseBody(t, "queued")

	client := environmentStatusReadyTestClient(t, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requestCount++

		return environmentStatusReadyTestResponse(
			request,
			&environmentStatusReadySignalingBody{
				Reader: strings.NewReader(responseBody),
				closed: firstResponseClosed,
			},
		), nil
	}))

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	implementation, request, response := environmentStatusReadyTestRead(t, client)
	result := make(chan struct{})

	go func() {
		implementation.Read(ctx, request, &response)
		close(result)
	}()

	select {
	case readGoroutineID := <-firstResponseClosed:
		waitForEnvironmentStatusReadyPoll(t, readGoroutineID)
	case <-time.After(time.Second):
		t.Fatal("first environment response was not consumed")
	}

	cancel()

	select {
	case <-result:
	case <-time.After(time.Second):
		t.Fatal("environment readiness wait did not stop after cancellation")
	}

	require.Len(t, response.Diagnostics.Errors(), 1)
	assert.Equal(t, "Cancelled waiting for environment to become ready", response.Diagnostics.Errors()[0].Summary())
	assert.Equal(t, 1, requestCount)

	var state EnvironmentStatusReadyModel

	stateDiagnostics := response.State.Get(t.Context(), &state)
	require.False(t, stateDiagnostics.HasError(), stateDiagnostics)
	assert.Equal(t, types.StringValue("queued"), state.Status)
}

func TestEnvironmentStatusReadyDataSourceCancellationDuringHTTPIO(t *testing.T) {
	t.Parallel()

	requestStarted := make(chan struct{})
	releaseResponse := make(chan struct{})

	requestCount := 0

	responseBody := environmentStatusReadyTestResponseBody(t, environmentStatusReadyValue)

	client := environmentStatusReadyTestClient(t, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requestCount++

		close(requestStarted)
		<-releaseResponse

		return environmentStatusReadyTestResponse(
			request,
			io.NopCloser(strings.NewReader(responseBody)),
		), nil
	}))

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	implementation, request, response := environmentStatusReadyTestRead(t, client)
	result := make(chan struct{})

	go func() {
		implementation.Read(ctx, request, &response)
		close(result)
	}()

	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("environment request did not start")
	}

	cancel()
	close(releaseResponse)

	select {
	case <-result:
	case <-time.After(time.Second):
		t.Fatal("environment readiness read did not stop after cancellation")
	}

	require.Len(t, response.Diagnostics.Errors(), 1)
	assert.Equal(t, "Cancelled waiting for environment to become ready", response.Diagnostics.Errors()[0].Summary())
	assert.Equal(t, 1, requestCount)
	assert.True(t, response.State.Raw.IsNull())
}

type environmentStatusReadySignalingBody struct {
	io.Reader

	closed chan<- string
}

func waitForEnvironmentStatusReadyPoll(t *testing.T, readGoroutineID string) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	stackBuffer := make([]byte, 1<<20)

	for time.Now().Before(deadline) {
		stackLength := runtime.Stack(stackBuffer, true)

		for stack := range strings.SplitSeq(string(stackBuffer[:stackLength]), "\n\n") {
			firstLine, _, _ := strings.Cut(stack, "\n")
			if strings.HasPrefix(firstLine, "goroutine "+readGoroutineID+" [select]") &&
				strings.Contains(stack, "(*environmentStatusReadyDataSource).Read") {
				return
			}
		}

		runtime.Gosched()
	}

	t.Fatal("environment readiness read did not enter the poll wait")
}

func (b *environmentStatusReadySignalingBody) Close() error {
	stackBuffer := make([]byte, 64)

	stackLength := runtime.Stack(stackBuffer, false)
	b.closed <- strings.Fields(string(stackBuffer[:stackLength]))[1]

	return nil
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
