package provider_test

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// expectRecoveryTaint inspects Core's persisted state before refresh can replace
// the failed mutation's checkpoint. The framework has no built-in taint check.
type expectRecoveryTaint struct {
	address string
	tainted bool
}

func (check expectRecoveryTaint) CheckPlan(_ context.Context, request plancheck.CheckPlanRequest, response *plancheck.CheckPlanResponse) {
	if request.Plan != nil && request.Plan.PriorState != nil && request.Plan.PriorState.Values != nil && request.Plan.PriorState.Values.RootModule != nil {
		for _, resource := range request.Plan.PriorState.Values.RootModule.Resources {
			if resource.Address != check.address {
				continue
			}

			if resource.Tainted != check.tainted {
				response.Error = fmt.Errorf("%w: resource %s tainted = %t, want %t", errUnexpectedTerraformPlan, check.address, resource.Tainted, check.tainted)
			}

			return
		}
	}

	response.Error = fmt.Errorf("%w: resource %s is absent from prior state", errUnexpectedTerraformPlan, check.address)
}
