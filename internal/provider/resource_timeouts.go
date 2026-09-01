package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const defaultResourceOperationTimeout = 2 * time.Minute

const minimumStoredResourceOperationTimeout = 10 * time.Second

func resourceCreateContext(ctx context.Context, value timeouts.Value) (context.Context, context.CancelFunc, diag.Diagnostics) {
	timeout, diags := value.Create(ctx, defaultResourceOperationTimeout)
	if diags.HasError() {
		return ctx, func() {}, diags
	}

	operationCtx, cancel := context.WithTimeout(ctx, timeout)

	return operationCtx, cancel, diags
}

func resourceReadContext(ctx context.Context, value timeouts.Value) (context.Context, context.CancelFunc, diag.Diagnostics) {
	timeout, diags := value.Read(ctx, defaultResourceOperationTimeout)
	if diags.HasError() {
		return ctx, func() {}, diags
	}

	operationCtx, cancel := context.WithTimeout(ctx, max(timeout, minimumStoredResourceOperationTimeout))

	return operationCtx, cancel, diags
}

func resourceUpdateContext(ctx context.Context, value timeouts.Value) (context.Context, context.CancelFunc, diag.Diagnostics) {
	timeout, diags := value.Update(ctx, defaultResourceOperationTimeout)
	if diags.HasError() {
		return ctx, func() {}, diags
	}

	operationCtx, cancel := context.WithTimeout(ctx, timeout)

	return operationCtx, cancel, diags
}

func resourceDeleteContext(ctx context.Context, value timeouts.Value) (context.Context, context.CancelFunc, diag.Diagnostics) {
	timeout, diags := value.Delete(ctx, defaultResourceOperationTimeout)
	if diags.HasError() {
		return ctx, func() {}, diags
	}

	operationCtx, cancel := context.WithTimeout(ctx, max(timeout, minimumStoredResourceOperationTimeout))

	return operationCtx, cancel, diags
}

func TimeoutsNull() timeouts.Value {
	return timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"update": types.StringType,
			"delete": types.StringType,
		}),
	}
}
