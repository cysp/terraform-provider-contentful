package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSigningSecretCancelledPlanIsNotAWarning(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	schema := AppSigningSecretResourceSchema(ctx)
	model := appSigningSecretTestModel("")
	model.Value = types.StringNull()
	plan := appSigningSecretTestPlan(ctx, t, schema, model)
	model.ValueWO = types.StringValue(strings.Repeat("a", 64))
	config := appSigningSecretTestPlan(ctx, t, schema, model)

	cancel()

	response := resource.ModifyPlanResponse{Plan: plan}
	modifySigningSecretPlan(ctx, resource.ModifyPlanRequest{
		Config: tfsdk.Config(config), State: tfsdk.State(plan), Plan: plan,
	}, &response)
	require.Len(t, response.Diagnostics, 1)
	assert.Equal(t, diag.SeverityError, response.Diagnostics[0].Severity())
	assert.Equal(t, "Write-only secret operation cancelled", response.Diagnostics[0].Summary())
	assert.Equal(t, "context canceled", response.Diagnostics[0].Detail())
	assert.True(t, response.Plan.Raw.Equal(plan.Raw))
}
