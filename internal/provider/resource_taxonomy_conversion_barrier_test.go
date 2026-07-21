//nolint:testpackage
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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaxonomyConceptUpdateConversionErrorDoesNotPatch(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	_, err = server.Handler().PutTaxonomyConcept(ctx, &cm.TaxonomyConceptRequest{
		PrefLabel: cm.LocalizedString{"en-US": "Concept"},
	}, cm.PutTaxonomyConceptParams{OrganizationID: "organization", TaxonomyConceptID: "concept"})
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

	model := TaxonomyConceptModel{
		IDIdentityModel:              NewIDIdentityModelFromMultipartID("organization", "concept"),
		TaxonomyConceptIdentityModel: TaxonomyConceptIdentityModel{OrganizationID: types.StringValue("organization"), ConceptID: types.StringValue("concept")},
		URI:                          types.StringUnknown(),
		PrefLabel:                    types.MapValueMust(types.StringType, map[string]attr.Value{"en-US": types.StringValue("Concept")}),
		AltLabels:                    types.MapNull(types.ListType{ElemType: types.StringType}),
		HiddenLabels:                 types.MapNull(types.ListType{ElemType: types.StringType}),
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

	resourceSchema := TaxonomyConceptResourceSchema(ctx)
	plan := tfsdk.Plan{Schema: resourceSchema}
	require.False(t, plan.Set(ctx, &model).HasError())

	implementation := taxonomyConceptResource{providerData: ContentfulProviderData{client: client}}
	response := resource.UpdateResponse{State: tfsdk.State{Schema: resourceSchema}}
	implementation.Update(ctx, resource.UpdateRequest{Plan: plan}, &response)

	require.True(t, response.Diagnostics.HasError())
	assert.Zero(t, patchCount.Load())
}
