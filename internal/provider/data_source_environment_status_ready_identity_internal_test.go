package provider

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/synctest"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentStatusReadyRejectsContradictoryIdentityBeforePublication(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, space, environment, status string
		queuedFirst                      bool
	}{
		{name: "wrong space ready", space: "other-space", environment: "environment", status: "ready"},
		{name: "wrong environment ready", space: "space", environment: "other-environment", status: "ready"},
		{name: "wrong pending stops polling", space: "space", environment: "other-environment", status: "queued"},
		{name: "valid queued then wrong ready", space: "space", environment: "other-environment", status: "ready", queuedFirst: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				invalid := cm.Environment{Sys: cm.NewEnvironmentSys(test.space, test.environment, test.status), Name: "unrelated"}
				invalidBody, err := json.Marshal(&invalid)
				require.NoError(t, err)
				queued := environmentStatusReadyTestResponseBody(t, "queued")
				ready := environmentStatusReadyTestResponseBody(t, "ready")

				expectedRequests := 1
				if test.queuedFirst {
					expectedRequests = 2
				}

				requests := 0
				client := environmentStatusReadyTestClient(t, roundTripFunc(func(request *http.Request) (*http.Response, error) {
					requests++

					assert.Equal(t, "/spaces/space/environments/environment", request.URL.Path)

					body := string(invalidBody)
					if test.queuedFirst && requests == 1 {
						body = queued
					}

					if requests > expectedRequests {
						body = ready
					}

					return environmentStatusReadyTestResponse(request, io.NopCloser(strings.NewReader(body))), nil
				}))
				implementation, request, response := environmentStatusReadyTestRead(t, client)
				response.State.Raw = request.Config.Raw.Copy()
				before := response.State.Raw.Copy()
				implementation.Read(t.Context(), request, &response)

				if assert.Len(t, response.Diagnostics.Errors(), 1) {
					assert.Equal(t, "Unexpected environment readiness identity", response.Diagnostics.Errors()[0].Summary())
				}

				assert.Equal(t, expectedRequests, requests)

				if test.queuedFirst {
					var state EnvironmentStatusReadyModel
					require.Empty(t, response.State.Get(t.Context(), &state))
					assert.Equal(t, types.StringValue("space/environment"), state.ID)
					assert.Equal(t, types.StringValue("space"), state.SpaceID)
					assert.Equal(t, types.StringValue("environment"), state.EnvironmentID)
					assert.Equal(t, types.StringValue("queued"), state.Status)
					assert.True(t, environmentStatusReadyTimeoutValue(types.StringNull()).Equal(state.Timeouts))
				} else {
					assert.True(t, before.Equal(response.State.Raw))
				}
			})
		})
	}
}

func TestEnvironmentStatusReadyPreservesAddressedAlias(t *testing.T) {
	t.Parallel()

	environment := cm.Environment{Sys: cm.NewEnvironmentSys("space", "environment", "ready"), Name: "target"}
	environment.Sys.AliasedEnvironment = cm.NewOptEnvironmentLink(cm.NewEnvironmentLink("different-target"))
	body, err := json.Marshal(&environment)
	require.NoError(t, err)

	requests := 0
	client := environmentStatusReadyTestClient(t, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++

		assert.Equal(t, "/spaces/space/environments/environment", request.URL.Path)

		return environmentStatusReadyTestResponse(request, io.NopCloser(strings.NewReader(string(body)))), nil
	}))
	implementation, request, response := environmentStatusReadyTestRead(t, client)
	implementation.Read(t.Context(), request, &response)
	require.Empty(t, response.Diagnostics)
	assert.Equal(t, 1, requests)

	var state EnvironmentStatusReadyModel
	require.Empty(t, response.State.Get(t.Context(), &state))
	assert.Equal(t, types.StringValue("space/environment"), state.ID)
	assert.Equal(t, types.StringValue("environment"), state.EnvironmentID)
	assert.Equal(t, types.StringValue("ready"), state.Status)
}
