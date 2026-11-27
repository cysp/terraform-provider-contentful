package provider_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

type webhookPrivateVerifiers struct {
	Basic *struct {
		Username string `json:"username"`
		Hash     string `json:"hash"`
	} `json:"basic"`
	Headers map[string]string `json:"headers"`
}

// Private state and failed Apply replies are unavailable to TestStep state checks.
// Observe the real provider replies produced during the framework's lifecycle.
type webhookProtocolResult struct {
	private []byte
}

func (r webhookProtocolResult) privateFields(t *testing.T) map[string][]byte {
	t.Helper()

	var fields map[string][]byte
	if len(r.private) > 0 {
		require.NoError(t, json.Unmarshal(r.private, &fields))
	}

	return fields
}

func (r webhookProtocolResult) verifiers(t *testing.T) webhookPrivateVerifiers {
	t.Helper()

	var result webhookPrivateVerifiers
	if data := r.privateFields(t)["write_only_secret_hashes"]; len(data) > 0 {
		require.NoError(t, json.Unmarshal(data, &result))
	}

	return result
}

type webhookProtocolObservations struct {
	mu      sync.Mutex
	applies []webhookProtocolResult
}

func (o *webhookProtocolObservations) lastApply() webhookProtocolResult {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.applies[len(o.applies)-1]
}

type observedWebhookProvider struct {
	tfprotov6.ProviderServer

	observations *webhookProtocolObservations
}

func (p observedWebhookProvider) ApplyResourceChange(ctx context.Context, req *tfprotov6.ApplyResourceChangeRequest) (*tfprotov6.ApplyResourceChangeResponse, error) {
	resp, err := p.ProviderServer.ApplyResourceChange(ctx, req)
	if req.TypeName == "contentful_webhook" && resp != nil {
		p.observations.mu.Lock()
		defer p.observations.mu.Unlock()

		p.observations.applies = append(p.observations.applies, webhookProtocolResult{private: resp.Private})
	}

	return resp, err //nolint:wrapcheck // Preserve the provider RPC error.
}

func testObservedWebhookResource(t *testing.T, handler http.Handler, observations *webhookProtocolObservations, steps []resource.TestStep) {
	t.Helper()

	resource.Test(t, validateAcceptanceTestCase(t, resource.TestCase{ProtoV6ProviderFactories: observations.providerFactories(t, handler), Steps: steps}))
}

func newSeededWebhookFixture(t *testing.T) *webhookSecretRecoveryFixture {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	server.SetWebhookDefinition("space", "imported", cm.WebhookDefinitionData{
		Name: "Imported", URL: "https://example.invalid/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
		HttpBasicUsername: cm.NewOptNilString("user"), HttpBasicPassword: cm.NewOptNilString(webhookPriorPassword),
		Headers: cm.WebhookDefinitionRequestHeaders{
			{Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString(webhookPriorHeader)},
			{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString(webhookSiblingHeader)},
		},
	})

	return &webhookSecretRecoveryFixture{server: server}
}

func webhookRecoveryStep(configText, password, header string) resource.TestStep {
	return resource.TestStep{Config: configText, ConfigVariables: config.Variables{"password": config.StringVariable(password), "header": config.StringVariable(header)}}
}

func importedWebhookRecoveryStep(t *testing.T, fixture *webhookSecretRecoveryFixture) resource.TestStep {
	t.Helper()

	step := webhookRecoveryStep(webhookRecoveryConfig(true, "")+webhookRecoveryImportBlock, webhookPriorPassword, webhookPriorHeader)
	step.PostApplyFunc = func() {
		fixture.requirePuts(t, webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true))
		fixture.requireStored(t, webhookPriorPassword, webhookPriorHeader, true)
	}

	return step
}

const webhookRecoveryImportBlock = `
import {
    to=contentful_webhook.test
    id="space/imported"
}
`

func (o *webhookProtocolObservations) providerFactories(t *testing.T, handler http.Handler) map[string]func() (tfprotov6.ProviderServer, error) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	factories := makeTestAccProtoV6ProviderFactories(testProviderOptionsWithHTTPServer(server)...)
	factory := factories["contentful"]
	factories["contentful"] = func() (tfprotov6.ProviderServer, error) {
		provider, err := factory()

		return observedWebhookProvider{ProviderServer: provider, observations: o}, err
	}

	return factories
}
