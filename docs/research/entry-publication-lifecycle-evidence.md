# Entry publication lifecycle evidence

Reviewed 2026-08-26. This note uses only first-party documentation, upstream
source, and the repository's sanitized direct experiments. The repository pins
`terraform-plugin-framework` v1.19.0 (commit
[`c7ac25e`](https://github.com/hashicorp/terraform-plugin-framework/tree/c7ac25e86333d194946fb5e3fd1114e7d101fc23)).
The reviewed Contentful JavaScript client is v12.15.0 (commit
[`cc096a3`](https://github.com/contentful/contentful-management.js/tree/cc096a337f0e1db6114e8da645d69bb6eb90f11c)).

## Conclusions

- Contentful explicitly requires the current Entry version on full-body update
  and describes optimistic locking as rejecting a stale update. The publish
  endpoint accepts the version being published, and the first-party client sends
  the fetched `sys.version` to both update and publish. These are strong grounds
  for an exact pre-publication version fence.
- Contentful does **not** document that Create must return version 1, Update must
  return exactly the submitted version plus one, Publish must return
  `publishedVersion` equal to the submitted version, or Publish must return
  `version` equal to the submitted version plus one. Current examples and direct
  observations show the normal arithmetic, but they are not guarantees.
- The first-party client contains one explicit arithmetic assumption: publishing
  increments `version` by one, so an entity has unpublished changes only when
  `version > publishedVersion + 1`. It does not validate mutation-response
  arithmetic before returning the response to its caller.
- Terraform Framework ordering supports a standard plan design for
  `published_version`: `UseStateForUnknown` can first preserve the prior value,
  and resource-level `ModifyPlan`, which runs later, can make it unknown only for
  a real publish or publication recovery.
- Resource private state is suitable for recording pending publication
  authority: Terraform persists response state even when the provider returns an
  error, and private data can be saved by Update. There is no mid-RPC persistence
  primitive, so a process loss before Terraform persists the Update response
  cannot durably grant recovery authority.
- Primary sources do not declare Entry metadata tags or concepts ordered, and a
  sanitized live multi-tag probe demonstrated that CMA can return tag order
  different from the order in the latest Update request. That justifies
  order-insensitive tag comparison, but not a public List-to-Set schema change.
- A sanitized live unpublish probe returned an Entry whose `version` advanced
  beyond the pending draft and whose `publishedVersion` was absent; a subsequent
  GET returned the same tuple. Exact pending-version recovery is therefore
  revoked naturally after the observed external-unpublish transition.
- Create-with-ID has a distinct create contract: content-type header, create
  payload, and no prior-version lock. The first-party client implements it
  separately from update. Correcting an inherited update-shaped Create path is
  justified, but it is not necessary to publication recovery and is best kept in
  a separate PR.

## Evidence classes

This note keeps the requested evidence classes separate:

- **C — documented guarantee:** normative Terraform or Contentful documentation.
- **D — first-party client assumption:** behavior encoded by
  `contentful-management.js`, which is strong compatibility evidence but not an
  API contract.
- **E — direct observation:** a sanitized live result or a documentation example;
  useful for modelling normal behavior, not proof that all valid responses have
  that shape.
- **A/B/F — provider rationale:** respectively a safety/ownership invariant, a
  Terraform consistency invariant, or defensive anomaly detection. These describe
  why the provider checks something; they do not turn its Contentful premise into
  a documented guarantee.

## Terraform lifecycle and state

Terraform's upstream resource lifecycle defines configuration, prior state,
planned state, and new state, and requires a known planned value to be returned
unchanged after apply. Values whose result is not predictable must remain unknown
in the plan. It also distinguishes response normalization (preserve the prior
representation) from actual drift (publish the remote value into state). See the
upstream [Resource Instance Change Lifecycle](https://github.com/hashicorp/terraform/blob/8b54d8708cc34f875aa59c3163893f72bdbf498a/docs/resource-instance-change-lifecycle.md#planresourcechange)
and the Framework [plan consistency rules](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#terraform-data-consistency-rules).

The Framework's documented plan sequence is:

1. apply defaults;
2. if the resource plan differs from prior state, mark unconfigured computed
   attributes unknown;
3. run attribute plan modifiers;
4. run resource plan modifiers.

The sequence is documented under
[Plan Modification Process](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#plan-modification-process).
In pinned v1.19.0,
[`int64planmodifier.UseStateForUnknown`](https://github.com/hashicorp/terraform-plugin-framework/blob/c7ac25e86333d194946fb5e3fd1114e7d101fc23/resource/schema/int64planmodifier/use_state_for_unknown.go#L12-L59)
copies a known prior value, including known null, only when the plan is unknown,
the resource already exists, and configuration is not unknown. Because
resource-level modification runs later, it can explicitly restore unknown for
operations whose publication result is not predictable. If it leaves a known
prior `published_version`, Update must return that exact value or Core will report
an inconsistent result.

Terraform documents `ignore_changes` as considering configured values on Create
but ignoring them on Update so that another process can share management of a
remote object. The Core developer documentation is more precise: for an existing
object, Terraform uses the corresponding prior-state value instead of the
configured value. See the user-facing
[`ignore_changes` contract](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#ignore_changes)
and upstream [planning behavior](https://github.com/hashicorp/terraform/blob/8b54d8708cc34f875aa59c3163893f72bdbf498a/docs/planning-behaviors.md#configuration-driven-behaviors).
Consequently, a provider that later performs a full-body CMA update from the
effective plan will include ignored fields using their refreshed/prior planned
values; publication applies to that complete Entry version.

Framework [private-state documentation](https://developer.hashicorp.com/terraform/plugin/framework/resources/private-state)
says private data is hidden from plans, is readable during plan/Read/Update, and
is writable during plan/Read/Update as well as Create and import. Framework
[diagnostic semantics](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics#how-errors-affect-state)
say Terraform persists returned state even with an error specifically so a
provider can checkpoint successful earlier calls in a multi-call mutation. Taken
together, those contracts support persisting a pending exact version when draft
write succeeded but publish failed. The provider gets one Update response; these
sources define no durable checkpoint part-way through the RPC.

## Contentful Entry contract

The CMA [updating contract](https://www.contentful.com/developers/docs/references/content-management-api/overview/#updating-content)
requires sending the entire Entry body; omitted properties are lost. Its
[version-locking contract](https://www.contentful.com/developers/docs/references/content-management-api/overview/#updating-and-version-locking)
requires the current version in `X-Contentful-Version` and says Contentful rejects
the update if the version changed in between.

The [Entry reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/)
adds these relevant guarantees:

- Create requires `X-Contentful-Content-Type`, creates a draft, and applies
  Content Type field defaults only for omitted fields at creation; defaults do
  not affect Update.
- Create-with-ID and Update share a path but have different headers: content type
  for Create, last Entry version for Update.
- Empty Entry fields, and `fields` itself when empty, are omitted from GET
  responses.
- Publish is a separate `/published` operation; locale-based publication is a
  separate optional feature.
- Whole-Entry unpublish is an unversioned `DELETE` of `/published`. The
  first-party client decodes its response as an Entry but does not validate the
  resulting version tuple.

The common-system-property table defines `sys.version` as the current version and
`sys.publishedVersion` as the published version, but specifies no arithmetic
relationship between them. See [Common resource attributes](https://www.contentful.com/developers/docs/references/content-management-api/overview/#common-resource-attributes).

The current [Create-with-ID endpoint](https://www.contentful.com/developers/docs/references/content-management-api/entries/create-an-entry-with-a-specified-id/)
says to pass `X-Contentful-Content-Type` when creating and
`X-Contentful-Version` when updating. Its generated example contains both headers,
so the endpoint narrative and first-party source are better evidence of the
create/update distinction than that one example. The first-party REST adapter's
[`createWithId`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L255-L280)
sends a create payload with `X-Contentful-Content-Type` and no version header.

The first-party adapter's
[`update`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L155)
deep-copies the Entry, removes `sys`, sends the remaining complete body, and
sets `X-Contentful-Version` from the fetched Entry. Its
[`publish`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L168-L185)
sets the same header from `sys.version`. Neither adapter checks the returned
version tuple.

## Version-arithmetic assessment

The categories below identify the provider purpose of each current check and the
actual Contentful evidence beneath it.

| Check | Primary classification | Evidence and implication |
| --- | --- | --- |
| Create returns draft version `1` | **E** | The current Create example and the repository's sanitized probes return `1`; no narrative contract promises the initial number, and the first-party client does not check it. The provider therefore accepts any positive, plan-consistent Create draft version. |
| Successful full-body Update returns exactly prior version `+ 1` | **A**, with an **E** arithmetic premise | Exactness is currently used to prove the draft version about to be published is the provider's write. Optimistic locking documents the starting fence, but Contentful does not document the response increment. Strictness is defensible at this pre-publication ownership boundary, but the exact arithmetic is not C. |
| A draft mutation response has missing or older `publishedVersion` | **F** | CMA defines the fields but does not specify their arithmetic. The provider accepts an absent value or any positive value strictly older than the returned draft version; equal, future, or nonpositive values are contradictory. |
| Publish returns `publishedVersion ==` submitted draft version | **B/F** | Sending the exact draft version is the safety fence. Once that request succeeds, response equality is post-publication state validation. CMA does not promise the equality and the current generated Publish example sends version `6` but shows `publishedVersion: 9`, so that page cannot support an equality guarantee. |
| Normal Publish returns current `version == submitted + 1` | **D** | The first-party [`isUpdated` helper](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/plain/checks.ts#L3-L11) explicitly says publishing increments version by one. This is a client assumption and matches the sanitized live observations, but is not in the CMA contract. |
| A higher coherent current version receives special treatment | **F** | The first-party helper treats `version > publishedVersion + 1` as later unpublished changes, but no source says a Publish response may or may not already contain such a later version. Warning/adoption is defensive interpretation, not documented mutation-response behavior. |
| On Read, observe positive `publishedVersion >= version` | **F**, informed by **D** | This differs from the first-party status arithmetic for normal published/updated entries, but CMA does not make that arithmetic a representability rule. The provider preserves the positive tuple and warns; it does not turn tolerant observation into publication authority. |
| Whole-Entry unpublish advances `version` and removes `publishedVersion` | **E** | The first-party client establishes the unversioned DELETE and Entry response shape but no arithmetic. The sanitized [unpublish probe](entry-unpublish-version.md) observed `version` advance by one from a pending draft and `publishedVersion` disappear in both the response and subsequent GET. The fake models that observed normal transition; the provider relies only on the resulting version mismatch to revoke stale recovery authority. |

The repository's existing sanitized direct experiments in
[Entry null and omission behavior](entry-null-and-omission.md) observed Create
`1/absent`, first Publish `2/1`, Update `3/1`, and later Publish `4/3` (with
additional repetitions). The
[CMA test-server conformance note](cma-test-server-conformance.md#sanitized-direct-observations)
records another `1 -> 2/1` publication. These are valuable **E** evidence for the
fake's normal mode, but they do not convert exact `+1` arithmetic into a guarantee.

### Policy implication

The primary evidence supports **strict ownership boundary, tolerant
observation**:

- Before Publish, require exact publication authority for the provider-written
  draft and send that exact version. Never grant ownership from content equality
  or from a GET that may include another actor's version.
- Keep recovery authority stricter than response observation: the pending
  publication marker must equal the current version, and publication must be
  absent or strictly older. An equal or newer observed `publishedVersion`
  revokes the marker even though a positive unusual tuple remains representable
  state.
- After a successful exact-version Publish request, treat representable CMA state
  as authoritative where doing so cannot authorize a later mutation. Unexpected
  arithmetic can be warned about, but it requires exact response fields and
  cannot classify response-only fields as creation defaults. Malformed identity,
  immutable Content Type, or unrepresentable lifecycle data remains an error.
- A bounded GET is useful only to learn remote status after an ambiguous response;
  it must not turn matching fields or a newly observed version into authority to
  publish.

Confidence is **high** that exact arithmetic is undocumented. The implemented
policy keeps exactness at the pre-publication ownership boundary and treats
representable post-publication arithmetic as authoritative observation, with a
warning when it differs from the repeatedly observed normal shape.

## Metadata ordering

The Entry reference calls `metadata.tags` and `metadata.concepts` lists. The
[Tags reference](https://www.contentful.com/developers/docs/references/content-management-api/tags/#tags-on-entries-and-assets)
describes adding/removing tag links and supports filters such as
`[all]` and `[in]`; it does not promise response order or say order is meaningful.
The taxonomy/Entry references likewise provide no ordering guarantee for concept
assignments.

The first-party client types both properties as arrays in
[`MetadataProps`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/common-types.ts#L486-L489).
The Entry adapter deep-copies and sends the caller's metadata without sorting,
and wraps returned Entry data without semantic normalization. This shows that the
client round-trips array order; it does **not** show that CMA guarantees that
order or that users assign semantic meaning to it.

The sanitized live probe below used three tags. CMA returned the request order in
the immediate Create and Update responses. After an Update submitted the same
three tag assignments in reverse or mixed order, however, the next GET and the
subsequent Publish response returned the original assignment order. Repeated GETs
were stable in that order. This is direct **E** evidence that tag request order is
not preserved across lifecycle boundaries. It does not establish a documented
set contract, nor does it establish concept behavior.

The public representation choices remain materially different:

- keep Lists and compare metadata as unordered while preserving a stable
  practitioner representation (no schema migration, moderate custom logic);
- migrate Lists to Sets (clean unordered model, but public type/state addressing
  and migration compatibility impact);
- retain order-sensitive Lists and document stable ordering as a provider
  assumption (no implementation churn, but possible false diffs if CMA reorders).

Because actual CMA reordering is observed, the resource keeps Lists for schema
compatibility but compares metadata without regard to order while preserving
duplicate multiplicity. It preserves the configured or prior order when the
assignments are otherwise equivalent. Concept ordering remains an explicit
provider assumption because the live probe could not exercise it.

## Sanitized live CMA probe

This direct **E** probe used credentials from the ignored example repository.
No access token, space/environment identifier, Entry identifier, tag identifier,
Content Type identifier, locale, or taxonomy identifier was logged or retained.
All objects created in the space used a unique per-run suffix.

The operations and response status classes were:

1. Read the target environment and default locale (`200`), and read up to three
   existing organization concepts (`200`, read-only).
2. Create three private, environment-scoped tags (`201` each).
3. Try to create a disposable Content Type referring directly to the three
   existing concepts. CMA rejected that shape with `422`; no organization-level
   taxonomy object was created or modified. Retry without taxonomy succeeded
   (`201`), and activation succeeded (`200`). Concept ordering was therefore not
   probed.
4. Create an Entry at a caller-selected ID with a PUT containing
   `X-Contentful-Content-Type` and deliberately no `X-Contentful-Version`
   (`201`). This directly confirms that the documented and first-party
   create-with-ID request shape is accepted by the associated live CMA.
5. Publish the returned draft version (`200`), perform a full-body Update with
   the three tag links reversed (`200`), publish (`200`), then perform another
   full-body Update in a mixed tag order (`200`) and publish (`200`). Repeated
   GETs were made after each publication and around the reverse-order draft.

The sanitized version tuples were:

| Boundary | `version` | `publishedVersion` |
| --- | ---: | ---: |
| Create response | 1 | absent |
| First Publish response and five GETs | 2 | 1 |
| Reverse-order Update response and three draft GETs | 3 | 1 |
| Second Publish response and five GETs | 4 | 3 |
| Mixed-order Update response | 5 | 3 |
| Third Publish response and five GETs | 6 | 5 |

All three rounds exhibited the normal `+1` arithmetic. This strengthens the
repeated-live-observation class **E** only; it does not resolve the absence of a
documented arithmetic guarantee and therefore does not settle the strict versus
tolerant post-publication policy.

The tags were labelled symbolically by creation order as `T1`, `T2`, and `T3`:

| Boundary | Request order | Response/GET order |
| --- | --- | --- |
| Create response, first Publish, five GETs | `T1,T2,T3` | `T1,T2,T3` |
| Reverse full-body Update response | `T3,T2,T1` | `T3,T2,T1` |
| Three GETs of that draft | no new order submitted | `T1,T2,T3` |
| Second Publish response and five GETs | no new order submitted | `T1,T2,T3` |
| Mixed full-body Update response | `T2,T1,T3` | `T2,T1,T3` |
| Third Publish response and five GETs | no new order submitted | `T1,T2,T3` |

Thus the immediate mutation response echoed request order, while later CMA reads
and publication canonicalized the unchanged assignments to an older assignment
order. Tests should not model metadata List order as stable across these
boundaries.

A separate sanitized [whole-Entry unpublish probe](entry-unpublish-version.md)
observed HTTP 200, an Entry response whose `version` advanced beyond the pending
draft, and absent `publishedVersion`; a subsequent GET returned the same tuple.
That observation is the evidence for the fake's normal unpublish transition.

## Implementation boundaries supported by the evidence

- Keep exact `X-Contentful-Version` publication fencing and pending publication
  authority recorded in resource private state.
- Authorize recovery only while the pending version is still current and has not
  been published or superseded; use the same boundary to revoke stale markers on
  Read.
- Model the live-observed whole-Entry unpublish response as an advanced Entry
  version without `publishedVersion`; that transition revokes any older pending
  publication authority by exact-version mismatch.
- Do not infer ownership from refreshed fields, import, or matching configuration.
- Use `UseStateForUnknown` plus later resource-level invalidation to keep
  `published_version` known for representation-only updates and unknown for real
  publication/recovery.
- Model `+1` version transitions as documented direct-observation behavior in the
  fake's normal mode, but name contradictory tuples as adversarial behavior.
- Permit response-only creation defaults after Create publication only for the
  complete normal publication tuple; anomalous positive arithmetic remains a
  warning with exact field ownership.
- Broaden the public failure limitation beyond transport/server errors: any
  incomplete or contradictory successful draft response, or process loss before
  Terraform persists the returned resource state and resource private state, can
  leave a matching unpublished draft that Terraform must not claim later.
- Keep Create-with-ID correction independent unless publication work directly
  depends on it.
