---
page_title: "Operation timeouts"
description: |-
  Explains Contentful provider operation timeout defaults, readiness waiting, and deadline precedence.
---

# Operation timeouts

The `timeouts` attribute sets provider operation budgets. It does not change Contentful service-side behavior, and a longer configured timeout cannot extend an earlier deadline supplied by Terraform or another parent operation. The earliest deadline always wins.

## Managed resource operations

Managed resources default each supported `create`, `read`, `update`, and `delete` operation to 2 minutes. A configured Create or Update timeout is used as written. Read and Delete load their timeout values from Terraform state; their effective provider deadline has a 10-second minimum, even when the stored value is shorter. The configured value remains unchanged in state. This minimum preserves enough time for refresh, recovery, and destroy work that must use previously stored settings, but it cannot extend an earlier parent deadline.

The timeout for an operation includes its Contentful requests and any eligible HTTP retries. Refer to each resource's schema for the operations it supports and to Terraform's [resource timeout syntax](https://developer.hashicorp.com/terraform/language/resources/configure#define-operation-timeouts).

## Data sources and list resources

Data-source timeouts are separate from managed-resource timeouts. Current data sources other than the readiness waiter default `timeouts.read` to 2 minutes. That Read timeout covers the complete data-source operation, including pagination when applicable.

The [`contentful_environment_status_ready` data source](../data-sources/environment_status_ready) is the deliberate exception: its `timeouts.read` controls the complete readiness wait and defaults to 10 minutes. It requests the environment immediately and uses a 15-second polling ticker until Contentful reports `ready`; it fails immediately for `failed`. The [`contentful_environment` resource](../resources/environment) does not wait for readiness, so increasing that resource's `timeouts.create` does not increase the readiness wait.

List resources do not expose a `timeouts` attribute. They retain a deadline supplied by Terraform. When an individual Contentful request has no deadline, the HTTP client adds a 2-minute request-and-retry budget; that fallback does not create a configurable timeout for the complete list operation.

Use Terraform's [`depends_on` meta-argument](https://developer.hashicorp.com/terraform/language/meta-arguments/depends_on) to express a behavioral dependency on a completed readiness data-source read when no attribute reference can express that dependency.
