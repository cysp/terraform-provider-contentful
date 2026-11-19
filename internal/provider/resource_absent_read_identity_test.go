package provider_test

import (
	"fmt"
	"maps"
	"net/http"
	"slices"
	"sync"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

func TestResourceReadAbsentIdentity(t *testing.T) {
	t.Parallel()

	// The schema RPCs need no Contentful requests.
	protocol, err := makeTestAccProtoV6ProviderFactories()["contentful"]()
	require.NoError(t, err)
	schema, err := protocol.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	require.Empty(t, schema.Diagnostics)
	identities, err := protocol.GetResourceIdentitySchemas(t.Context(), &tfprotov6.GetResourceIdentitySchemasRequest{})
	require.NoError(t, err)
	require.Empty(t, identities.Diagnostics)

	cases := readEndpointCases(t)
	names := make([]string, 0, len(cases))

	for _, test := range cases {
		resourceName := "contentful_" + test.name
		names = append(names, resourceName)
		require.Contains(t, identities.IdentitySchemas, resourceName)
		identityType := identities.IdentitySchemas[resourceName].ValueType()

		for _, supplied := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/prior_identity=%t", test.name, supplied), func(t *testing.T) {
				t.Parallel()

				var (
					requestMu sync.Mutex
					requests  []string
				)

				server := configuredProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requestMu.Lock()

					requests = append(requests, r.Method+" "+r.URL.RequestURI())
					requestMu.Unlock()

					_ = cmt.WriteContentfulManagementErrorResponse(w, http.StatusNotFound, "NotFound", new("Absent"), nil)
				}))
				state := mutationEndpointDynamicValue(t, test.schema, test.priorState())
				request := &tfprotov6.ReadResourceRequest{TypeName: resourceName, CurrentState: &state}
				expectedIdentity := maps.Clone(test.identity)

				if supplied {
					// Distinct values expose accidental replacement of an established identity.
					for name, value := range expectedIdentity {
						expectedIdentity[name] = "prior-" + value
					}

					request.CurrentIdentity = &tfprotov6.ResourceIdentityData{
						IdentityData: &tfprotov6.DynamicValue{JSON: []byte(testJSON(expectedIdentity))},
					}
				}

				response, err := server.ReadResource(t.Context(), request)
				require.NoError(t, err)
				requestMu.Lock()
				defer requestMu.Unlock()

				require.Equal(t, []string{"GET " + test.endpoint}, requests)

				if test.absentWarning == "" {
					require.Empty(t, response.Diagnostics)
				} else {
					detail := test.absentDetail
					if detail == "" {
						detail = "Error: NotFound: Absent"
					}

					require.Equal(t, []*tfprotov6.Diagnostic{{
						Severity: tfprotov6.DiagnosticSeverityWarning, Summary: test.absentWarning, Detail: detail,
					}}, response.Diagnostics)
				}

				require.NotNil(t, response.NewState)
				refreshed, err := response.NewState.Unmarshal(test.schema.Type().TerraformType(t.Context()))
				require.NoError(t, err)
				require.True(t, refreshed.IsNull())
				require.NotNil(t, response.NewIdentity)
				require.NotNil(t, response.NewIdentity.IdentityData)
				actual, err := response.NewIdentity.IdentityData.Unmarshal(identityType)
				require.NoError(t, err)
				expected, err := (&tfprotov6.DynamicValue{JSON: []byte(testJSON(expectedIdentity))}).Unmarshal(identityType)
				require.NoError(t, err)
				require.True(t, expected.Equal(actual), "expected %s, got %s", expected, actual)
			})
		}
	}

	require.ElementsMatch(t, slices.Collect(maps.Keys(schema.ResourceSchemas)), names)
}
