package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

const operationTimeoutDescription = "Timeout for the complete operation, including retries, such as `30s` or `5m`. Defaults to `2m`."

const storedOperationTimeoutDescription = operationTimeoutDescription + " Uses the value saved in state, with a minimum of `10s`."

//nolint:ireturn // Preserve the timeouts library's schema.Attribute interface.
func resourceTimeoutsAttribute(ctx context.Context, opts timeouts.Opts) schema.Attribute {
	opts.CreateDescription = operationTimeoutDescription
	opts.ReadDescription = storedOperationTimeoutDescription
	opts.UpdateDescription = operationTimeoutDescription
	opts.DeleteDescription = storedOperationTimeoutDescription

	return timeouts.Attributes(ctx, opts)
}
