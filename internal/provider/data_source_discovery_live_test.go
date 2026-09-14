package provider_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest
func TestAccDiscoveryDataSourcesLiveRead(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for live discovery tests")
	}

	if os.Getenv("TF_ACC_MOCKED") != "" {
		t.Skip("live discovery requires the Contentful API")
	}

	require.NotEmpty(t, os.Getenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN"), "CONTENTFUL_MANAGEMENT_ACCESS_TOKEN must be set for live discovery tests")

	// Read the shared acceptance Space without creating or modifying its objects.
	testAccMockableResource(t, nil, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.contentful_space.test", tfjsonpath.New("space_id"), knownvalue.StringExact("0p38pssr0fi3")),
					statecheck.ExpectKnownValue("data.contentful_environment.test", tfjsonpath.New("environment_id"), knownvalue.StringExact("master")),
					statecheck.ExpectKnownValue("data.contentful_environment_aliases.test", tfjsonpath.New("environment_aliases"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.contentful_locale.test", tfjsonpath.New("default"), knownvalue.Bool(true)),
				},
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				PlanOnly:        true,
			},
		},
	})
}
