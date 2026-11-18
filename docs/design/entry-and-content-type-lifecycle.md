# Entry publication and Content Type activation

Each resource writes a draft, then publishes the Entry or activates the Content Type. Only a confirmed draft write by that resource authorizes the second request. Read, import, an external draft, and a match between configuration and remote fields confer no such authority.

See the [Entry lifecycle observations](../research/entry-lifecycle.md) for the CMA requests and version transitions behind this contract.

## Draft versions and recovery

1. When the draft write returns a resource, clear any previous pending marker.
   Compare the response identity to the
   requested endpoint and checkpoint the representable draft state for that
   endpoint. After state and identity are saved, record the response's private
   `version` if the identity matches, or clear the private `version` if it differs.
2. Require confirmed endpoint identity, plan consistency, and a valid draft
   version tuple. `version` must be positive. If present, `publishedVersion`
   must be non-negative and less than `version`.
3. Record that exact draft version in private state and request activation or
   publication with it. Only positive marker versions authorize these operations.
4. Confirm that `publishedVersion` equals the submitted version and `version` is
   greater. The observed `V+1` increment is not a provider requirement. Checkpoint
   the complete confirmation response before diagnosing contradictions; a
   contradictory response revokes authority.

Confirmed activation or publication clears the marker. A terminal
`VersionMismatch` on a draft update, activation, or publication revokes any
pending authority, whether it occurs on the first attempt or after a 429 retry.
An unchanged plan then schedules no recovery mutation.

Other request failures during activation or publication retain the validated
draft and marker. A later unchanged plan schedules Update to retry the marked
version without another draft write. During Create, the provider reports a
failed activation or publication request as a warning so Terraform can retain
the resource, even when a version conflict has revoked the marker. Failures
before the validated draft checkpoint remain errors and grant no authority.

Read keeps a marker only while both the version and publication tuple match the
checkpointed draft. If that draft is already published, it clears the marker
without another mutation. Any other tuple revokes authority. With refresh
disabled, recovery still sends only the recorded version. No path fetches and
publishes a newer draft.

An ambiguous draft-write failure grants no authority. A later GET with matching
fields cannot prove that Terraform authored that version. Requests follow the
[HTTP retry policy](contentful-http-retry-policy.md), including explicit 429
retries with the same body and preconditions.

## Content Type activation ownership

Updates use the exact Contentful version and the effective plan, including refreshed values retained by `ignore_changes`. When planning preserves omitted metadata taxonomy, that known value constrains both the draft and activation responses. Unknown taxonomy remains response-owned.

`published_version` did not exist in state written by older provider versions.
A normal post-upgrade refresh projects `sys.publishedVersion`. With
`-refresh=false`, Terraform decodes the missing legacy Computed value as null,
but no pending activation marker exists, so publication remains observational
and the unchanged transition is a no-op.

## Entry specified-ID request selection

The Entry specified-ID endpoint serves both Create and Update, so the provider
selects the operation through its request headers. Create sends
`X-Contentful-Content-Type` and no `X-Contentful-Version`. A collision must fail
without mutating, publishing, or adopting the existing Entry. Update sends the
exact prior `sys.version` in `X-Contentful-Version` and omits
`X-Contentful-Content-Type`.

If an Entry disappears after refresh, or refresh is skipped and Terraform state
is stale, the Update request therefore fails instead of recreating the absent
Entry and proceeding toward publication. A preflight existence read cannot
provide this boundary because another writer can change target existence
between the read and the `PUT`. The Contentful contract and direct CMA evidence
are recorded in
[Entry and Content Type PUT header semantics](../research/entry-and-content-type-put-headers.md).

## Entry publication ownership and partial field ownership

Entry fields are a required map of field IDs to maps of locale codes and JSON strings. Terraform-null field and locale elements are omitted from the CMA request. A known per-locale `jsonencode(null)` is sent as JSON null. Change detection compares the effective plan and prior state after projecting those omissions. Adding or removing an omitted element can update Terraform state without writing or publishing an Entry. The stored `published_version` remains response-owned. See the [CMA null and omission observations](../research/entry-fields.md).

Read exposes returned field and locale values, including additions and removals. It retains a prior Terraform-null element when the corresponding key is absent. A returned value takes precedence. A formerly nonempty localized field is not reconstructed solely from omitted child locales. The provider also retains a known empty fields map when CMA omits the entire member and a field containing only empty localized arrays when CMA omits that field. A localized JSON-null value receives no omission fallback.

The effective plan after `ignore_changes` controls the complete mutation payload. Paths can identify a whole field or a single locale. An ignored-only external draft remains untouched. A subsequent managed write includes the ignored values observed during refresh. Mutation response checks apply the same omission and empty-array rules as Read.

The v0 decoder preserves supported state, including Terraform-null elements and metadata. The legacy whole-field JSON null cannot be represented in the nested map. Upgrading such state emits a diagnostic at the field path and stops before publishing any upgraded state.

Entry metadata tags and taxonomy concepts retain their public List schema.
Configured values must be unique. Comparison ignores order but remains
multiplicity-sensitive for response/state data so irregular remote duplicates
are not hidden. A configuration-only reorder changes only Terraform's stored
list representation; it does not write or publish an Entry.

Contentful may apply Content Type default values to fields omitted during Entry
creation. A Create response may therefore add response-only field keys,
including a value for a Terraform-null key omitted from the request, but every
sent field must still be present and semantically equal after the recognized
omission and empty-array fallbacks above. Update uses full-body
replacement and Contentful does not apply defaults to Entry updates, so an
Update response must have exactly the sent field keys and values after those
same fallbacks. Create publication responses retain the same creation-default
projection used for the Create draft response; Update and recovery publication
responses require the exact effective-plan field keys. Field ownership does not
depend on publication-version arithmetic.

After those checks succeed, mutation state restores the exact effective-plan
`fields`, `metadata`, and `timeouts`; identity and lifecycle values remain
response-owned where the plan does not already establish an endpoint identity.
A later Read remains authoritative and can expose Contentful-added defaults,
which `ignore_changes` can preserve through a subsequent managed update. When
Contentful contradicts a planned field, metadata, content type, or endpoint
identity, the provider reports an error and checkpoints the representable
response `fields`, `metadata`, and `content_type_id`, while retaining the requested
endpoint identity. The content type is a related-resource reference, not part of
the Entry endpoint. A refresh-disabled retry can plan replacement for a
contradictory content type; cleanup still addresses the intended Entry. A
successful refresh can instead restore the actual content type without granting
publication authority. Response-only field additions cannot become
configuration-owned values or new publication authority.

The computed `published_version` exposes Contentful's `sys.publishedVersion`. Save a private `version` only when the response confirms the endpoint under the [identity policy](terraform-value-semantics.md#mutation-endpoint-identity).

## Entry destroy lifecycle

Entry destroy consumes no private `version`, performs no recovery GET, and
sends neither `X-Contentful-Version` nor `If-Match`. It always requests
whole-Entry unpublish before delete. A 404 response is benign, as is an exact
HTTP 400 response with `sys.id` `BadRequest` and message `Not published`;
destroy continues to delete after either result. Any other unpublish error
stops destroy before delete. Because Contentful does not enforce preconditions
on these operations, an Entry changed outside Terraform can still be deleted.
This policy is limited to Entry destroy; Entry update, Entry publish, and other
endpoints are out of scope.

See [Entry unpublish and delete evidence](../research/entry-lifecycle.md#unpublish-and-delete-preconditions).

## Terraform guarantees used by recovery

The Framework runs resource plan modification after attribute plan modifiers. Where `UseStateForUnknown` preserves a known `published_version`, planning must restore unknown before an operation whose result can change it. Otherwise Core requires the prior value after apply. [Framework planning process](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#plan-modification-process) and [v1.19.0 modifier implementation](https://github.com/hashicorp/terraform-plugin-framework/blob/c7ac25e86333d194946fb5e3fd1114e7d101fc23/resource/schema/int64planmodifier/use_state_for_unknown.go#L12-L59)

Core substitutes prior state for ignored paths on update. A complete Entry write therefore includes the refreshed values of ignored fields when another field changes. [Core planning behavior](https://github.com/hashicorp/terraform/blob/8b54d8708cc34f875aa59c3163893f72bdbf498a/docs/planning-behaviors.md#configuration-driven-behaviors)

Private state carries the pending draft version between operations. Terraform can persist returned state alongside an error, allowing a provider to record completed calls in a multi-request operation. Recheck these guarantees on [dependency upgrades](terraform-value-semantics.md#upgrade-triggers). [Framework private state](https://developer.hashicorp.com/terraform/plugin/framework/resources/private-state) and [error diagnostics](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics#how-errors-affect-state)
