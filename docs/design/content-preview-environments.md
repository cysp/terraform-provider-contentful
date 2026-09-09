# Preview environment configuration and reconciliation

The Contentful [preview environment API reference](../research/content-preview-environments.md)
records the independently observed wire behavior. This document records the
provider's representation and request choices. Shared mutation consistency and
recovery rules remain in [Terraform value semantics](terraform-value-semantics.md#lifecycle-ownership-and-plan-consistency).

## Request and response projection

The provider:

- shares one canonical request model between create and update, with a distinct response model;
- uses only `entityType` and `entityId` as request identity fields;
- normalizes either response identity to the content type ID map key;
- omits `contentType` and `example` from requests; and
- never serializes a response object directly into an update request.

The generated client and test server model the canonical request shape used by the provider. The observed legacy create-only `contentType` alias remains historical API evidence and is not exposed as a second request type.

## Active configuration ownership

The provider exposes only active content-type configurations as a map keyed by content type ID. It keeps `entityType` and `enabled` behind the provider interface: adding a key or changing its URL sends `enabled: true`, removing a key sends `enabled: false`, and unchanged identities are omitted from the update payload. Reads and imports filter disabled configurations from Terraform state. Re-adding a disabled key either re-enables the retained identity or recreates it if the service no longer retains that disabled record.

Creation order has no established practitioner-visible meaning. The provider
uses a map and sorts changed keys only for deterministic request construction.

## Concurrency and recovery

The provider constructs configuration updates from the Terraform state-to-plan delta without an additional read. This prevents unrelated concurrent configuration changes from being included in the request. When the full mutation response exposes an unexpected active configuration, the provider immediately reports the contradiction after checkpointing the returned recovery state and version; a later refresh remains the fallback when the mutation response does not expose a concurrent change. A concurrent change to the same identity Terraform is updating remains last-writer-wins because the service does not advance `sys.version` for configuration-only changes.

The provider sends the last observed version and surfaces conflicts without
refreshing and replaying automatically. Removing a preview URL disables its
configuration; replacing a configured selected-ID platform with nonempty
configuration requires a new ID because successful reuse was not established.
Selected-ID replacement must not create before deleting the previous object:
the shared PUT endpoint can update an existing address.

## Scope

The public resource omits response-only `sys.version`, timestamps, aliases, and
`example`. Platform ordering, preview mode, and custom preview tokens remain
separate concerns. Do not infer provider validation for URL schemes, placeholder
grammar, remote content-type existence, entitlements, or unverified ID characters
from permissive client types or mock behavior. Live checks cover the undocumented
contract; deterministic mock tests cover broader provider lifecycles.
