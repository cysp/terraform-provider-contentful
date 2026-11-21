package provider_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest
func TestAccRolesDataSourceLivePagination(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for live Role pagination tests")
	}

	if os.Getenv("TF_ACC_MOCKED") != "" {
		t.Skip("live Role pagination requires the Contentful API")
	}

	require.NotEmpty(t, os.Getenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN"), "CONTENTFUL_MANAGEMENT_ACCESS_TOKEN must be set for live Role pagination tests")

	const spaceID = "0p38pssr0fi3"

	recorder := &rolePaginationRecorder{transport: http.DefaultTransport, path: "/spaces/" + spaceID + "/roles"}
	roles := statecheck.CompareValue(compare.ValuesSame())

	var baseline []string

	// Read the existing acceptance Space twice without creating or modifying Roles.
	testAccMockableResource(t, nil, resource.TestCase{
		ProtoV6ProviderFactories: makeTestAccProtoV6ProviderFactories(provider.WithHTTPClient(&http.Client{Transport: recorder})),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{"space_id": config.StringVariable(spaceID)},
				ConfigStateChecks: []statecheck.StateCheck{
					roles.AddStateValue("data.contentful_roles.test", tfjsonpath.New("roles")),
				},
				PostApplyFunc: func() {
					baseline = requireRolePaginationBaseline(t, recorder.pages())
				},
			},
			{
				PreConfig:       func() { recorder.setLimit(1) },
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{"space_id": config.StringVariable(spaceID)},
				ConfigStateChecks: []statecheck.StateCheck{
					roles.AddStateValue("data.contentful_roles.test", tfjsonpath.New("roles")),
				},
				PostApplyFunc: func() {
					requireRolePaginationReads(t, recorder.pages(), baseline)
				},
			},
		},
	})
}
