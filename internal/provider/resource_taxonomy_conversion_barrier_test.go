//nolint:testpackage // Resource lifecycle methods are intentionally tested through their package-local implementations.
package provider

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaxonomyConceptUpdateRequestConversionErrorStopsBeforeAPIRequest(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	model := taxonomyConceptUpdatePlan()
	model.AltLabels = types.MapValueMust(types.ListType{ElemType: types.StringType}, map[string]attr.Value{
		"en-US": types.ListUnknown(types.StringType),
	})

	assertTaxonomyUpdateRequestConversionErrorStopsBeforeAPIRequest(
		t,
		model,
		TaxonomyConceptResourceSchema(ctx),
		func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
			implementation := taxonomyConceptResource{providerData: ContentfulProviderData{client: client}}
			implementation.Update(ctx, request, response)
		},
	)
}

func TestTaxonomyConceptSchemeUpdateRequestConversionErrorStopsBeforeAPIRequest(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	model := taxonomyConceptSchemeUpdatePlan()
	model.TopConceptIDs = types.ListValueMust(types.StringType, []attr.Value{types.StringUnknown()})

	assertTaxonomyUpdateRequestConversionErrorStopsBeforeAPIRequest(
		t,
		model,
		TaxonomyConceptSchemeResourceSchema(ctx),
		func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
			implementation := taxonomyConceptSchemeResource{providerData: ContentfulProviderData{client: client}}
			implementation.Update(ctx, request, response)
		},
	)
}

func assertTaxonomyUpdateRequestConversionErrorStopsBeforeAPIRequest(
	t *testing.T,
	model any,
	resourceSchema schema.Schema,
	update func(*cm.Client, resource.UpdateRequest, *resource.UpdateResponse),
) {
	t.Helper()

	ctx := t.Context()
	client, requestCount := taxonomyRequestCountingClient(t)

	plan := tfsdk.Plan{Schema: resourceSchema}
	require.False(t, plan.Set(ctx, model).HasError())

	response := resource.UpdateResponse{State: tfsdk.State{Schema: resourceSchema}}
	update(client, resource.UpdateRequest{Plan: plan}, &response)

	require.True(t, response.Diagnostics.HasError())
	assert.Zero(t, requestCount.Load())
}

func taxonomyRequestCountingClient(t *testing.T) (*cm.Client, *atomic.Int64) {
	t.Helper()

	requestCount := &atomic.Int64{}

	testServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}))
	t.Cleanup(testServer.Close)

	client, err := cm.NewClient(
		testServer.URL,
		cm.NewAccessTokenSecuritySource("access-token"),
		cm.WithClient(testServer.Client()),
	)
	require.NoError(t, err)

	return client, requestCount
}

func taxonomyConceptUpdatePlan() TaxonomyConceptModel {
	listType := types.ListType{ElemType: types.StringType}

	return TaxonomyConceptModel{
		IDIdentityModel:              NewIDIdentityModelFromMultipartID("organization", "concept"),
		TaxonomyConceptIdentityModel: TaxonomyConceptIdentityModel{OrganizationID: types.StringValue("organization"), ConceptID: types.StringValue("concept")},
		URI:                          types.StringNull(),
		PrefLabel:                    types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Concept")}),
		AltLabels:                    types.MapNull(listType),
		HiddenLabels:                 types.MapNull(listType),
		Notations:                    types.ListNull(types.StringType),
		Note:                         types.MapNull(types.StringType),
		ChangeNote:                   types.MapNull(types.StringType),
		Definition:                   types.MapNull(types.StringType),
		EditorialNote:                types.MapNull(types.StringType),
		Example:                      types.MapNull(types.StringType),
		HistoryNote:                  types.MapNull(types.StringType),
		ScopeNote:                    types.MapNull(types.StringType),
		BroaderConceptIDs:            types.ListNull(types.StringType),
		RelatedConceptIDs:            types.ListNull(types.StringType),
		ConceptSchemeIDs:             types.SetUnknown(types.StringType),
		Timeouts:                     TimeoutsNull(),
	}
}

func taxonomyConceptSchemeUpdatePlan() TaxonomyConceptSchemeModel {
	return TaxonomyConceptSchemeModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID("organization", "scheme"),
		TaxonomyConceptSchemeIdentityModel: TaxonomyConceptSchemeIdentityModel{
			OrganizationID:  types.StringValue("organization"),
			ConceptSchemeID: types.StringValue("scheme"),
		},
		URI:           types.StringNull(),
		PrefLabel:     types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Scheme")}),
		Definition:    types.MapNull(types.StringType),
		TopConceptIDs: types.ListNull(types.StringType),
		ConceptIDs:    types.ListNull(types.StringType),
		TotalConcepts: types.Int64Unknown(),
		Timeouts:      TimeoutsNull(),
	}
}

func TestTaxonomyConceptSchemeUpdateConversionErrorDoesNotPatch(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	_, err = server.Handler().PutTaxonomyConceptScheme(ctx, &cm.TaxonomyConceptSchemeRequest{
		PrefLabel: cm.LocalizedString{"en-US": "Scheme"},
	}, cm.PutTaxonomyConceptSchemeParams{OrganizationID: "organization", TaxonomyConceptSchemeID: "scheme"})
	require.NoError(t, err)

	var patchCount atomic.Int64

	testServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPatch {
			patchCount.Add(1)
		}

		server.ServeHTTP(responseWriter, request)
	}))
	t.Cleanup(testServer.Close)

	client, err := cm.NewClient(testServer.URL, cm.NewAccessTokenSecuritySource(cmt.ValidAccessToken))
	require.NoError(t, err)

	model := TaxonomyConceptSchemeModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID("organization", "scheme"),
		TaxonomyConceptSchemeIdentityModel: TaxonomyConceptSchemeIdentityModel{
			OrganizationID:  types.StringValue("organization"),
			ConceptSchemeID: types.StringValue("scheme"),
		},
		URI:           types.StringUnknown(),
		PrefLabel:     types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Scheme")}),
		Definition:    types.MapNull(types.StringType),
		TopConceptIDs: types.ListNull(types.StringType),
		ConceptIDs:    types.ListNull(types.StringType),
		TotalConcepts: types.Int64Unknown(),
		Timeouts:      TimeoutsNull(),
	}

	resourceSchema := TaxonomyConceptSchemeResourceSchema(ctx)
	plan := tfsdk.Plan{Schema: resourceSchema}
	require.False(t, plan.Set(ctx, &model).HasError())

	implementation := taxonomyConceptSchemeResource{providerData: ContentfulProviderData{client: client}}
	response := resource.UpdateResponse{State: tfsdk.State{Schema: resourceSchema}}
	implementation.Update(ctx, resource.UpdateRequest{Plan: plan}, &response)

	require.True(t, response.Diagnostics.HasError())
	assert.Zero(t, patchCount.Load())
}
