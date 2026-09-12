# Content preview platform configuration and reconciliation

This document describes how `contentful_preview_environment` represents a
Contentful content preview platform and constructs its requests. The platform
belongs to a space; it is not a Contentful environment. The
[API research](../research/content-preview-environments.md) records the
independently observed HTTP behavior. Shared mutation consistency and recovery
rules remain in
[Terraform value semantics](terraform-value-semantics.md#lifecycle-ownership-and-plan-consistency).

## Request and response projection

The provider uses one canonical request model for Create and Update, with a
separate response model. Requests identify configurations with `entityType` and
`entityId`; they omit the response properties `contentType` and `example`. A
response object is never serialized directly into an Update request.

Response projection accepts `entityId` or the legacy `contentType` identity and
uses the content type ID as the Terraform map key. Missing or contradictory
identities, unsupported entity types, and duplicate active configurations produce
warnings because the map cannot preserve them. Mutation reconciliation also
reports a consistency error when this loss prevents comparison with the plan.
See the [request conversion](../../internal/provider/preview_environment_model_request.go)
and [response projection](../../internal/provider/preview_environment_model_response.go).

The generated client and test server model the canonical request shape. The
observed legacy Create-only `contentType` alias remains API evidence; it is not
exposed as a second request type.

## Active configuration ownership

The `content_type_configurations` map exposes only active Content Type
configurations, keyed by content type ID. The provider supplies `entityType` and
`enabled` internally:

- Adding a key or changing its URL sends `enabled: true`.
- Removing a key sends `enabled: false`.
- Unchanged configurations are omitted from the Update payload.
- Read and import omit disabled configurations from Terraform state.

Re-adding a disabled key re-enables the retained identity, or recreates it if
Contentful no longer retains that disabled record. Creation order has no
established practitioner-visible meaning. The provider sorts changed keys for
deterministic request construction.

## Concurrency and recovery

The provider builds configuration updates from the difference between Terraform
state and the effective plan, without an additional read. The payload therefore
contains only the configuration changes Terraform planned. It does not include
unrelated configuration changes made by another writer.

If the full mutation response contains an unexpected active configuration, the
provider reports the contradiction after saving the returned recovery state and
version. A later refresh can expose a concurrent change that was absent from the
mutation response. A concurrent change to the same configuration Terraform is
updating can be overwritten: the observed service behavior does not advance
`sys.version` for configuration-only changes.

The provider sends the last observed version and surfaces conflicts without
refreshing and replaying automatically. Removing a preview URL disables its
configuration. To replace a platform with a selected ID and nonempty
configuration, choose a new ID: successful reuse of a deleted configured
platform's ID was not established by the retained observations. Selected-ID
replacement must not use `create_before_destroy`, because the shared PUT endpoint
can update the existing address.

## Scope

The [resource schema](../../internal/provider/resource_preview_environment_schema.go)
limits selected IDs to 1–64 ASCII letters, digits, hyphens, or underscores. This
is the provider's supported subset, not a complete statement of Contentful's ID
grammar. It omits response-only timestamps, aliases, and `example` from public
state and stores `sys.version` privately.

Platform ordering, preview mode, and custom preview tokens are separate
concerns. The provider does not validate URL schemes, placeholder grammar,
remote Content Type existence, or account entitlements. Permissive client types
and mock behavior do not establish those API contracts. Retained live
observations support the undocumented endpoint behavior; deterministic mock
tests exercise broader provider lifecycles.
