package provider_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
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
	state        *tfprotov6.DynamicValue
	private      []byte
	priorPrivate []byte
	diagnostics  []*tfprotov6.Diagnostic
	logs         string
}

func (r webhookProtocolResult) model(t *testing.T) WebhookModel {
	t.Helper()
	require.NotNil(t, r.state)
	schema := WebhookResourceSchema(t.Context())
	raw, err := r.state.Unmarshal(schema.Type().TerraformType(t.Context()))
	require.NoError(t, err)

	state := tfsdk.State{Schema: schema, Raw: raw}

	var model WebhookModel
	require.Empty(t, state.Get(t.Context(), &model))

	return model
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
	mu                sync.Mutex
	reads, applies    []webhookProtocolResult
	nextReadVerifiers []byte
}

func (o *webhookProtocolObservations) firstApply() webhookProtocolResult {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.applies[0]
}

func (o *webhookProtocolObservations) injectNextReadVerifiers(verifiers []byte) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.nextReadVerifiers = verifiers
}

func (o *webhookProtocolObservations) lastApply() webhookProtocolResult {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.applies[len(o.applies)-1]
}

func (o *webhookProtocolObservations) lastRead() webhookProtocolResult {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.reads[len(o.reads)-1]
}

type observedWebhookProvider struct {
	tfprotov6.ProviderServer

	observations *webhookProtocolObservations
}

// The testing framework already attached a logging sink to the RPC context.
// Installing tflogtest directly there would reuse that sink and ignore the buffer.
// Override only the test logger's keys, retaining RPC cancellation and metadata.
type webhookRPCLogContext struct {
	context.Context //nolint:containedctx // Forward cancellation, deadlines and remaining values to the live RPC.

	logValue func(any) any
}

func (c webhookRPCLogContext) Value(key any) any {
	if value := c.logValue(key); value != nil {
		return value
	}

	return c.Context.Value(key)
}

func captureWebhookRPCLogs(ctx context.Context, logs *bytes.Buffer) context.Context {
	logger := tflogtest.RootLogger(context.Background(), logs)

	return webhookRPCLogContext{Context: ctx, logValue: logger.Value}
}

func (p observedWebhookProvider) ApplyResourceChange(ctx context.Context, req *tfprotov6.ApplyResourceChangeRequest) (*tfprotov6.ApplyResourceChangeResponse, error) {
	var logs bytes.Buffer

	resp, err := p.ProviderServer.ApplyResourceChange(captureWebhookRPCLogs(ctx, &logs), req)
	if req.TypeName == "contentful_webhook" && resp != nil {
		p.observations.mu.Lock()
		defer p.observations.mu.Unlock()

		p.observations.applies = append(p.observations.applies, webhookProtocolResult{state: resp.NewState, private: resp.Private, diagnostics: resp.Diagnostics, logs: logs.String()})
	}

	return resp, err //nolint:wrapcheck // Preserve the provider RPC error.
}

func (p observedWebhookProvider) ReadResource(ctx context.Context, req *tfprotov6.ReadResourceRequest) (*tfprotov6.ReadResourceResponse, error) {
	if req.TypeName != "contentful_webhook" {
		return p.ProviderServer.ReadResource(ctx, req) //nolint:wrapcheck // Preserve the provider RPC error.
	}

	p.observations.mu.Lock()
	stale := p.observations.nextReadVerifiers
	p.observations.nextReadVerifiers = nil
	p.observations.mu.Unlock()

	if stale != nil {
		var fields map[string][]byte

		err := json.Unmarshal(req.Private, &fields)
		if err != nil {
			return nil, fmt.Errorf("decode prior private fields: %w", err)
		}

		fields["write_only_secret_hashes"] = stale

		data, err := json.Marshal(fields)
		if err != nil {
			return nil, fmt.Errorf("encode injected private fields: %w", err)
		}

		req.Private = data
	}

	var logs bytes.Buffer

	resp, err := p.ProviderServer.ReadResource(captureWebhookRPCLogs(ctx, &logs), req)
	if resp != nil {
		p.observations.mu.Lock()
		defer p.observations.mu.Unlock()

		p.observations.reads = append(p.observations.reads, webhookProtocolResult{priorPrivate: req.Private, state: resp.NewState, private: resp.Private, diagnostics: resp.Diagnostics, logs: logs.String()})
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

func webhookReplyFields(fixture *webhookSecretRecoveryFixture, method string, edit func(map[string]any)) {
	fixture.onNextResponse(method, func(w http.ResponseWriter, rec *httptest.ResponseRecorder) {
		_, _ = w.Write([]byte(mutateTestJSON(rec.Body.String(), edit)))
	})
}

func webhookPUTWithFields(version, password, header string, fields map[string]any) webhookMutation {
	put := webhookExpectedRecoveryPut(version, password, header, true)
	put.body = mutateTestJSON(put.body, func(doc map[string]any) { maps.Copy(doc, fields) })

	return put
}

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
