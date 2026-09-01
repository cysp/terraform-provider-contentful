package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type resourceOperationContextFunc func(context.Context, timeouts.Value) (context.Context, context.CancelFunc, diag.Diagnostics)

func TestResourceOperationContextDurations(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		operation        string
		operationContext resourceOperationContextFunc
		explicitDuration time.Duration
	}{
		"create": {
			operation:        "create",
			operationContext: resourceCreateContext,
			explicitDuration: time.Second,
		},
		"read": {
			operation:        "read",
			operationContext: resourceReadContext,
			explicitDuration: 10 * time.Second,
		},
		"update": {
			operation:        "update",
			operationContext: resourceUpdateContext,
			explicitDuration: time.Second,
		},
		"delete": {
			operation:        "delete",
			operationContext: resourceDeleteContext,
			explicitDuration: 10 * time.Second,
		},
	}

	for name, test := range tests {
		t.Run(name+" default", func(t *testing.T) {
			t.Parallel()

			started := time.Now()

			operationCtx, cancel, diags := test.operationContext(t.Context(), TimeoutsNull())
			defer cancel()

			require.Empty(t, diags)
			assertContextDeadline(operationCtx, t, started.Add(2*time.Minute))
		})

		t.Run(name+" explicit", func(t *testing.T) {
			t.Parallel()

			started := time.Now()

			operationCtx, cancel, diags := test.operationContext(
				t.Context(), resourceTimeoutValue(test.operation, types.StringValue("1s")),
			)
			defer cancel()

			require.Empty(t, diags)
			assertContextDeadline(operationCtx, t, started.Add(test.explicitDuration))
		})
	}
}

func TestResourceOperationContextInvalidTimeouts(t *testing.T) {
	t.Parallel()

	tests := map[string]resourceOperationContextFunc{
		"create": resourceCreateContext,
		"read":   resourceReadContext,
		"update": resourceUpdateContext,
		"delete": resourceDeleteContext,
	}

	for operation, operationContext := range tests {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()

			parentCtx := t.Context()
			operationCtx, cancel, diags := operationContext(
				parentCtx, resourceTimeoutValue(operation, types.StringValue("invalid")),
			)
			cancel()

			assert.Equal(t, parentCtx, operationCtx)
			_, hasDeadline := operationCtx.Deadline()
			assert.False(t, hasDeadline)

			require.Len(t, diags, 1)
			assert.Equal(t, "Timeout Cannot Be Parsed", diags[0].Summary())
			assert.Equal(
				t,
				fmt.Sprintf("timeout for %q cannot be parsed, time: invalid duration %q", operation, "invalid"),
				diags[0].Detail(),
			)
		})
	}
}

func TestResourceOperationContextsHonorParentDeadline(t *testing.T) {
	t.Parallel()

	tests := map[string]resourceOperationContextFunc{
		"create": resourceCreateContext,
		"read":   resourceReadContext,
		"update": resourceUpdateContext,
		"delete": resourceDeleteContext,
	}

	for operation, operationContext := range tests {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()

			parentDeadline := time.Now().Add(30 * time.Second)

			parentCtx, parentCancel := context.WithDeadline(t.Context(), parentDeadline)
			defer parentCancel()

			operationCtx, cancel, diags := operationContext(
				parentCtx, resourceTimeoutValue(operation, types.StringValue("1h")),
			)
			defer cancel()

			require.Empty(t, diags)
			assertContextDeadline(operationCtx, t, parentDeadline)
		})
	}
}

func TestResourceOperationContextCancellation(t *testing.T) {
	t.Parallel()

	tests := map[string]resourceOperationContextFunc{
		"create": resourceCreateContext,
		"read":   resourceReadContext,
		"update": resourceUpdateContext,
		"delete": resourceDeleteContext,
	}

	for operation, operationContext := range tests {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()

			operationCtx, cancel, diags := operationContext(t.Context(), TimeoutsNull())
			require.Empty(t, diags)

			cancel()
			<-operationCtx.Done()

			assert.ErrorIs(t, operationCtx.Err(), context.Canceled)
		})
	}
}

func resourceTimeoutValue(operation string, value types.String) timeouts.Value {
	return timeouts.Value{Object: types.ObjectValueMust(
		map[string]attr.Type{operation: types.StringType},
		map[string]attr.Value{operation: value},
	)}
}

func assertContextDeadline(ctx context.Context, t *testing.T, expected time.Time) {
	t.Helper()

	deadline, ok := ctx.Deadline()
	require.True(t, ok)
	assert.WithinDuration(t, expected, deadline, 100*time.Millisecond)
}
