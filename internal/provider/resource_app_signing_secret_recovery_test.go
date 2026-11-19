package provider_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppSigningSecretLostAcknowledgement(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("organization", "app", cm.AppDefinitionData{Name: "App"})
	puts := &appSigningSecretPutRecorder{handler: server}

	var (
		requests        atomic.Int64
		loseResponse    atomic.Bool
		committedStatus atomic.Int64
	)

	protocol := configuredProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		assert.Equal(t, "/organizations/organization/app_definitions/app/signing_secret", r.URL.Path)
		assert.Empty(t, r.Header.Values("X-Contentful-Version"))

		if loseResponse.Swap(false) {
			// Run the mutation to completion, then disconnect before sending any response.
			response := httptest.NewRecorder()
			puts.ServeHTTP(response, r)
			committedStatus.Store(int64(response.Code))
			panic(http.ErrAbortHandler)
		}

		puts.ServeHTTP(w, r)
	}))

	ctx := t.Context()
	schema := AppSigningSecretResourceSchema(ctx)
	absent := nullResourceDynamicValue(t, schema)
	state := func(value string) tfprotov6.DynamicValue {
		return mutationEndpointDynamicValue(t, schema, map[string]any{
			"id": "organization/app", "organization_id": "organization", "app_definition_id": "app", "value": value,
		})
	}

	const resourceType = "contentful_app_signing_secret"

	apply := func(prior, plan *tfprotov6.DynamicValue) *tfprotov6.ApplyResourceChangeResponse {
		t.Helper()

		result, err := protocol.ApplyResourceChange(ctx, &tfprotov6.ApplyResourceChangeRequest{
			TypeName: resourceType, PriorState: prior, PlannedState: plan, Config: plan,
		})
		require.NoError(t, err)

		return result
	}
	requireState := func(expected, actual *tfprotov6.DynamicValue) {
		t.Helper()
		require.NotNil(t, actual)

		valueType := schema.Type().TerraformType(ctx)
		want, err := expected.Unmarshal(valueType)
		require.NoError(t, err)
		got, err := actual.Unmarshal(valueType)
		require.NoError(t, err)
		require.True(t, want.Equal(got), "unexpected Terraform state")
	}

	initial, rotated := strings.Repeat("a", 64), strings.Repeat("b", 64)
	initialPlan, rotatedPlan := state(initial), state(rotated)

	loseResponse.Store(true)

	create := apply(&absent, &initialPlan)
	require.Len(t, create.Diagnostics, 1)
	assert.Equal(t, "Failed to create app signing secret", create.Diagnostics[0].Summary)
	assert.Contains(t, create.Diagnostics[0].Detail, "final four characters")
	requireState(&absent, create.NewState)
	require.EqualValues(t, http.StatusCreated, committedStatus.Load())
	require.EqualValues(t, 1, requests.Load())
	puts.requireValues(t, initial)
	requireAppSigningSecretRedactedValue(t, server, "aaaa")

	// A deliberate second create receives an acknowledgement for the installed value.
	acknowledged := apply(&absent, &initialPlan)
	require.Empty(t, acknowledged.Diagnostics)
	requireState(&initialPlan, acknowledged.NewState)

	loseResponse.Store(true)

	update := apply(acknowledged.NewState, &rotatedPlan)
	require.Len(t, update.Diagnostics, 1)
	assert.Equal(t, "Failed to update app signing secret", update.Diagnostics[0].Summary)
	assert.Contains(t, update.Diagnostics[0].Detail, "Reverting")
	requireState(acknowledged.NewState, update.NewState)
	require.EqualValues(t, http.StatusOK, committedStatus.Load())
	require.EqualValues(t, 3, requests.Load())
	puts.requireValues(t, initial, initial, rotated)
	requireAppSigningSecretRedactedValue(t, server, "bbbb")

	read, err := protocol.ReadResource(ctx, &tfprotov6.ReadResourceRequest{TypeName: resourceType, CurrentState: update.NewState})
	require.NoError(t, err)
	require.Empty(t, read.Diagnostics)
	requireState(&initialPlan, read.NewState)
	require.EqualValues(t, 4, requests.Load())

	reverted := apply(read.NewState, &initialPlan)
	require.Empty(t, reverted.Diagnostics)
	requireState(&initialPlan, reverted.NewState)
	require.EqualValues(t, 4, requests.Load())
	requireAppSigningSecretRedactedValue(t, server, "bbbb")

	loseResponse.Store(true)

	deleted := apply(reverted.NewState, &absent)
	require.Len(t, deleted.Diagnostics, 1)
	assert.Equal(t, "Failed to delete app signing secret", deleted.Diagnostics[0].Summary)
	assert.Contains(t, deleted.Diagnostics[0].Detail, "remove a replacement secret")
	requireState(&initialPlan, deleted.NewState)
	require.EqualValues(t, http.StatusNoContent, committedStatus.Load())
	require.EqualValues(t, 5, requests.Load())

	remote, err := server.Handler().GetAppSigningSecret(ctx, cm.GetAppSigningSecretParams{OrganizationID: "organization", AppDefinitionID: "app"})
	require.NoError(t, err)

	status, ok := remote.(cm.StatusCodeResponse)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, status.GetStatusCode())
}
