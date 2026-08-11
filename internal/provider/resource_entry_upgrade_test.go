package provider_test

import (
	"context"
	"net/http/httptest"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	frameworklist "github.com/hashicorp/terraform-plugin-framework/list"
	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	testingresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccEntryResourceLegacyStateDoesNotAuthorizePublication(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetEntry("space", "environment", "article", "entry", cm.EntryRequest{
		Fields:   cm.NewOptEntryFields(cm.EntryFields{"managed": []byte(`{"en-US":"one"}`)}),
		Metadata: cm.NewOptEntryMetadata(cm.EntryMetadata{Concepts: []cm.TaxonomyConceptLink{}, Tags: []cm.TagLink{}}),
	})

	draft := getTestEntry(t, server)
	publishResponse, err := server.Handler().PublishEntry(t.Context(), cm.PublishEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "entry", XContentfulVersion: draft.Sys.Version,
	})
	require.NoError(t, err)

	published, ok := publishResponse.(*cm.EntryStatusCode)
	require.True(t, ok)

	errorSink := new(entryFixtureErrorSink)
	recorder := newEntryMutationRecorder(server, errorSink)

	defer func() {
		require.NoError(t, errorSink.error())
	}()

	testserver := httptest.NewServer(recorder)
	t.Cleanup(testserver.Close)
	options := ContentfulProviderOptionsWithHTTPTestServer(testserver)
	additionalCLIOptions := &testingresource.AdditionalCLIOptions{
		Plan: testingresource.PlanOptions{NoRefresh: true},
	}
	config := `
resource "contentful_entry" "test" {
  space_id        = "space"
  environment_id  = "environment"
  entry_id        = "entry"
  content_type_id = "article"
  fields = { managed = jsonencode({ "en-US" = "one" }) }
}
`

	testingresource.Test(t, testingresource.TestCase{
		AdditionalCLIOptions: additionalCLIOptions,
		Steps: []testingresource.TestStep{
			{
				ProtoV6ProviderFactories: legacyEntryProviderFactories(published.Response, options...),
				Config:                   config,
			},
			{
				PreConfig:                recorder.reset,
				ProtoV6ProviderFactories: makeTestAccProtoV6ProviderFactories(options...),
				Config:                   config,
				ConfigPlanChecks: testingresource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectKnownValue("contentful_entry.test", tfjsonpath.New("published_version"), knownvalue.Null()),
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_entry.test", tfjsonpath.New("published_version"), knownvalue.Null()),
				},
				Check: func(*terraform.State) error {
					requireNoEntryMutations(t, recorder)

					return nil
				},
			},
			{
				PreConfig: func() {
					additionalCLIOptions.Plan.NoRefresh = false

					recorder.reset()
				},
				ProtoV6ProviderFactories: makeTestAccProtoV6ProviderFactories(options...),
				Config:                   config,
				ConfigPlanChecks: testingresource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectKnownValue("contentful_entry.test", tfjsonpath.New("published_version"), knownvalue.Int64Exact(1)),
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_entry.test", tfjsonpath.New("published_version"), knownvalue.Int64Exact(1)),
				},
				Check: func(*terraform.State) error {
					requireNoEntryMutations(t, recorder)

					return nil
				},
			},
		},
	})
}

type legacyEntryProvider struct {
	*ContentfulProvider

	entry cm.Entry
}

func (p *legacyEntryProvider) ListResources(context.Context) []func() frameworklist.ListResource {
	return nil
}

func (p *legacyEntryProvider) Resources(context.Context) []func() frameworkresource.Resource {
	return []func() frameworkresource.Resource{
		func() frameworkresource.Resource { return &legacyEntryResource{entry: p.entry} },
	}
}

func legacyEntryProviderFactories(entry cm.Entry, options ...Option) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"contentful": providerserver.NewProtocol6WithError(func() frameworkprovider.Provider {
			return &legacyEntryProvider{ContentfulProvider: New("test", options...), entry: entry}
		}()),
	}
}

type legacyEntryResource struct {
	entry cm.Entry
}

func (r *legacyEntryResource) Metadata(_ context.Context, req frameworkresource.MetadataRequest, resp *frameworkresource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entry"
}

func (r *legacyEntryResource) Schema(ctx context.Context, _ frameworkresource.SchemaRequest, resp *frameworkresource.SchemaResponse) {
	resp.Schema = EntryResourceSchemaV0(ctx)
	delete(resp.Schema.Attributes, "published_version")
}

func (r *legacyEntryResource) Create(ctx context.Context, req frameworkresource.CreateRequest, resp *frameworkresource.CreateResponse) {
	var plan legacyEntryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	current, projectionDiags := NewEntryResourceModelFromResponse(ctx, r.entry)
	resp.Diagnostics.Append(projectionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	legacyFields := NewTypedMapNull[jsontypes.Normalized]()
	if r.entry.Fields.IsSet() {
		fields := make(map[string]jsontypes.Normalized, len(r.entry.Fields.Value))
		for fieldID, raw := range r.entry.Fields.Value {
			fields[fieldID] = NewNormalizedJSONTypesNormalizedValue(raw)
		}
		legacyFields = NewTypedMap(fields)
	}

	state := legacyEntryModel{
		IDIdentityModel:    current.IDIdentityModel,
		EntryIdentityModel: current.EntryIdentityModel,
		ContentTypeID:      current.ContentTypeID,
		Fields:             legacyFields,
		Metadata:           plan.Metadata,
		Timeouts:           plan.Timeouts,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", r.entry.Sys.Version)...)
}

func (r *legacyEntryResource) Read(context.Context, frameworkresource.ReadRequest, *frameworkresource.ReadResponse) {
}

func (r *legacyEntryResource) Update(context.Context, frameworkresource.UpdateRequest, *frameworkresource.UpdateResponse) {
}

func (r *legacyEntryResource) Delete(context.Context, frameworkresource.DeleteRequest, *frameworkresource.DeleteResponse) {
}

type legacyEntryModel struct {
	IDIdentityModel
	EntryIdentityModel

	ContentTypeID types.String `tfsdk:"content_type_id"`

	Fields   TypedMap[jsontypes.Normalized]  `tfsdk:"fields"`
	Metadata TypedObject[EntryMetadataValue] `tfsdk:"metadata"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
