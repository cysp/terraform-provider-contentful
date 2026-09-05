package provider_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

var errUnexpectedImportedState = errors.New("unexpected imported state")

// testAccImportAttributes checks a seeded CLI import, where there is no prior
// applied state for ImportStateVerify. Presence is checked separately from value;
// a missing flattened attribute does not establish a known empty string or null.
func testAccImportAttributes(expected map[string]string) resource.ImportStateCheckFunc {
	return func(states []*terraform.InstanceState) error {
		if len(states) != 1 || states[0] == nil {
			return fmt.Errorf("%w: expected one non-nil resource, got %d", errUnexpectedImportedState, len(states))
		}

		for name, want := range expected {
			got, exists := states[0].Attributes[name]
			if !exists || got != want {
				return fmt.Errorf("%w: attribute %q: present=%t, got %q, want %q", errUnexpectedImportedState, name, exists, got, want)
			}
		}

		return nil
	}
}

func TestImportAttributesCheck(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		states    []*terraform.InstanceState
		wantError bool
	}{
		"empty value present": {states: []*terraform.InstanceState{{Attributes: map[string]string{"description": ""}}}},
		"missing empty value": {states: []*terraform.InstanceState{{Attributes: map[string]string{}}}, wantError: true},
		"different value":     {states: []*terraform.InstanceState{{Attributes: map[string]string{"description": "other"}}}, wantError: true},
		"no resource":         {wantError: true},
		"nil resource":        {states: []*terraform.InstanceState{nil}, wantError: true},
		"multiple resources":  {states: []*terraform.InstanceState{{}, {}}, wantError: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := testAccImportAttributes(map[string]string{"description": ""})(test.states)
			if test.wantError {
				require.ErrorIs(t, err, errUnexpectedImportedState)

				return
			}

			require.NoError(t, err)
		})
	}
}

// testAccPriorState runs typed checks on the state Core read before the next
// apply. CLI import and RefreshState steps do not execute ConfigStateChecks.
type testAccPriorState struct {
	check statecheck.StateCheck
}

func (c testAccPriorState) CheckPlan(ctx context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	if req.Plan == nil {
		resp.Error = fmt.Errorf("%w: plan is nil", errUnexpectedImportedState)

		return
	}

	var result statecheck.CheckStateResponse
	c.check.CheckState(ctx, statecheck.CheckStateRequest{State: req.Plan.PriorState}, &result)
	resp.Error = result.Error
}

func TestPriorStateCheckDistinguishesNullAndEmpty(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		attributes map[string]any
		wantError  bool
	}{
		"null":   {attributes: map[string]any{"value": nil}},
		"empty":  {attributes: map[string]any{"value": ""}, wantError: true},
		"absent": {attributes: map[string]any{}, wantError: true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			check := testAccPriorState{check: statecheck.ExpectKnownValue("contentful_app_signing_secret.test", tfjsonpath.New("value"), knownvalue.Null())}

			var response plancheck.CheckPlanResponse
			check.CheckPlan(t.Context(), plancheck.CheckPlanRequest{Plan: &tfjson.Plan{PriorState: &tfjson.State{Values: &tfjson.StateValues{RootModule: &tfjson.StateModule{Resources: []*tfjson.StateResource{{Address: "contentful_app_signing_secret.test", AttributeValues: test.attributes}}}}}}}, &response)

			if test.wantError {
				require.Error(t, response.Error)

				return
			}

			require.NoError(t, response.Error)
		})
	}
}
