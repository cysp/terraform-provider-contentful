package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAcceptanceTestCaseRejectsIgnoredAssertions(t *testing.T) {
	t.Parallel()

	check := func(*terraform.State) error { return nil }
	importCheck := func([]*terraform.InstanceState) error { return nil }
	stateChecks := []statecheck.StateCheck{
		statecheck.ExpectKnownValue("contentful_tag.test", tfjsonpath.New("name"), knownvalue.StringExact("Test")),
	}

	tests := map[string]struct {
		step      resource.TestStep
		wantError string
	}{
		"configuration assertions": {step: resource.TestStep{Check: check, ConfigStateChecks: stateChecks}},
		"refresh assertion":        {step: resource.TestStep{RefreshState: true, Check: check}},
		"CLI import assertion":     {step: resource.TestStep{ImportState: true, ImportStateCheck: importCheck}},
		"nil import assertions":    {step: resource.TestStep{ImportState: true}},
		"empty import assertions":  {step: resource.TestStep{ImportState: true, ConfigStateChecks: []statecheck.StateCheck{}}},
		"ID import plan":           {step: resource.TestStep{ImportState: true, ImportStateKind: resource.ImportBlockWithID}},
		"identity import plan":     {step: resource.TestStep{ImportState: true, ImportStateKind: resource.ImportBlockWithResourceIdentity}},
		"CLI import legacy check": {
			step:      resource.TestStep{ImportState: true, Check: check},
			wantError: "acceptance assertion does not execute: step 2: Check on an import step",
		},
		"CLI import state checks": {
			step:      resource.TestStep{ImportState: true, ConfigStateChecks: stateChecks},
			wantError: "acceptance assertion does not execute: step 2: ConfigStateChecks on an import step",
		},
		"ID import state callback": {
			step:      resource.TestStep{ImportState: true, ImportStateKind: resource.ImportBlockWithID, ImportStateCheck: importCheck},
			wantError: "acceptance assertion does not execute: step 2: ImportStateCheck on an import-block planning step",
		},
		"identity import state callback": {
			step:      resource.TestStep{ImportState: true, ImportStateKind: resource.ImportBlockWithResourceIdentity, ImportStateCheck: importCheck},
			wantError: "acceptance assertion does not execute: step 2: ImportStateCheck on an import-block planning step",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testcase := resource.TestCase{Steps: []resource.TestStep{{}}}
			testcase.Steps = append(testcase.Steps, test.step)

			err := acceptanceTestCaseError(testcase)
			if test.wantError != "" {
				require.ErrorIs(t, err, errIgnoredAcceptanceCheck)
				require.EqualError(t, err, test.wantError)

				return
			}

			validateAcceptanceTestCase(t, testcase)
		})
	}
}
