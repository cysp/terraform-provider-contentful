# Entry publication lifecycle evidence

The provider may publish only an exact draft version returned by a validated
Create or Update. This note separates the external evidence for that boundary
from the provider's own consistency and anomaly checks. The authoritative
[Entry publication contract](../design/terraform-value-semantics.md#entry-publication-ownership-and-partial-field-ownership)
records the implementation policy.

Lifecycle evidence reviewed 2026-08-30; metadata evidence reviewed 2026-09-03.
Sources are first-party documentation, upstream source, and the repository's
sanitized direct experiments. The lifecycle review used
`terraform-plugin-framework` v1.19.0 (commit
[`c7ac25e`](https://github.com/hashicorp/terraform-plugin-framework/tree/c7ac25e86333d194946fb5e3fd1114e7d101fc23)).
The reviewed Contentful JavaScript client is v12.15.0 (commit
[`cc096a3`](https://github.com/contentful/contentful-management.js/tree/cc096a337f0e1db6114e8da645d69bb6eb90f11c)).

## Findings

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
  an operation that writes and publishes a new draft.
- Resource private state can record a pending exact version across operations:
  Terraform persists response state even when the provider returns an error,
  and private data can be saved by Update. The provider uses that mechanism only
  after checkpointing and validating a complete provider-authored draft
  response. A later unchanged operation may publish only that marked version;
  Read, import, prior state, field equality, and fresh GET responses never create
  or replace the marker.
- Primary sources do not declare Entry metadata tags or concepts ordered and do
  not define duplicate-link semantics. Separate sanitized probes of both
  properties found that immediate mutation responses echoed submitted order and
  duplicate occurrences, while subsequent GET and Publish responses restored an
  earlier assignment order and reduced each repeated link to one occurrence.
  This is observed canonicalization, not a documented set contract.
- A sanitized live unpublish probe returned an Entry whose `version` advanced
  beyond the pending draft and whose `publishedVersion` was absent; a subsequent
  GET returned the same tuple. This establishes the fake's normal unpublish
  transition. It revokes a stale marker because the observed current version no
  longer equals the marked draft, and never grants authority to the new version.
- Create-with-ID has a distinct create contract: content-type header, create
  payload, and no prior-version lock. The first-party client implements it
  separately from update.

## Evidence classes

The classification in the version table distinguishes:

- **Documentation:** normative Terraform or Contentful requirements.
- **Client assumption:** behavior encoded by the first-party management client;
  compatibility evidence that does not establish an API guarantee.
- **Observation:** a sanitized live result or documentation example; evidence of
  a normal response, not proof that every valid response has that shape.
- **Consistency** and **anomaly detection:** reasons for provider checks. These
  express Terraform constraints or provider defenses, not Contentful guarantees.

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
together, those contracts support cross-operation pending-version recovery when
a validated draft write succeeds but Publish is not confirmed. An ambiguous
Publish response does not prove that publication remains pending, so recovery
does not infer status or adopt a GET result: normal Read clears the marker if the
exact draft is already published and revokes it on any changed lifecycle tuple;
with refresh disabled it submits only the stale exact marker, and
`VersionMismatch` revokes it immediately.

## Contentful Entry contract

The CMA [updating contract](https://www.contentful.com/developers/docs/references/content-management-api/overview/#updating-content)
requires sending the entire Entry body; omitted properties are lost. Its
[version-locking contract](https://www.contentful.com/developers/docs/references/content-management-api/overview/#updating-and-version-locking)
requires the current version in `X-Contentful-Version` and says Contentful rejects
the update if the version changed in between.

The [Entry reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/)
documents these relevant operations and behaviors:

- Create requires `X-Contentful-Content-Type`, creates a draft, and applies
  Content Type field defaults only for omitted fields at creation; defaults do
  not affect Update.
- Create-with-ID and Update share a path but have different headers: content type
  for Create, last Entry version for Update.
- Empty Entry fields, and `fields` itself when empty, are omitted from GET
  responses.
- Publish is a separate `/published` operation; locale-based publication is a
  separate optional feature.
- Whole-Entry unpublish uses `DELETE` of `/published`. The endpoint reference
  lists an optional version header; the first-party client omits it and decodes
  the response as an Entry without validating the version tuple. The separate
  [destroy evidence](entry-destroy-lifecycle.md) records direct precondition probes.

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
| Create returns draft version `1` | **Observation** | The current Create example and the repository's sanitized probes return `1`; no narrative contract promises the initial number, and the first-party client does not check it. The provider therefore accepts any positive, plan-consistent Create draft version. |
| Successful full-body Update returns exactly prior version `+ 1` | **Observation** | The fake and direct observations exhibit this arithmetic, but optimistic locking documents only the starting fence and Contentful does not document the response increment. The provider does not require it: the direct no-retry response, identity, complete projection, and plan consistency establish provenance for its exact positive returned version. |
| A draft mutation response has missing or older `publishedVersion` | **Anomaly detection** | CMA defines the fields but does not specify their arithmetic. The provider accepts an absent value or any non-negative value strictly older than the returned draft version; equal, future, negative, or unknown values are contradictory. |
| Publish returns `publishedVersion ==` submitted draft version | **Consistency / anomaly detection** | Sending the exact draft version is the safety fence. Once that request succeeds, response equality is post-publication state validation. CMA does not promise the equality and the current generated Publish example sends version `6` but shows `publishedVersion: 9`, so that page cannot support an equality guarantee. |
| Normal Publish returns current `version == submitted + 1` | **Client assumption** | The first-party [`isUpdated` helper](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/plain/checks.ts#L3-L11) explicitly says publishing increments version by one. This is a client assumption and matches the sanitized live observations, but is not in the CMA contract. |
| A higher coherent current version receives special treatment | **Anomaly detection** | The first-party helper treats `version > publishedVersion + 1` as later unpublished changes, but no source says a Publish response may or may not already contain such a later version. The provider therefore applies no separate warning or field policy to this arithmetic. |
| On Read, observe positive `publishedVersion >= version` | **Anomaly detection**, informed by **Client assumption** | This contradicts the lifecycle ordering required by exact-version recovery. The provider fails closed and cannot preserve or create publication authority from the tuple. |
| Whole-Entry unpublish advances `version` and removes `publishedVersion` | **Observation** | The first-party client establishes the unversioned DELETE and Entry response shape but no arithmetic. The sanitized [unpublish probe](entry-unpublish-version.md) observed `version` advance by one from a pending draft and `publishedVersion` disappear in both the response and subsequent GET. The fake models that observed normal transition. The provider does not infer publication authority from the resulting state. |

The repository's existing sanitized direct experiments in
[Entry null and omission behavior](entry-null-and-omission.md) observed Create
`1/absent`, first Publish `2/1`, Update `3/1`, and later Publish `4/3` (with
additional repetitions). The
[CMA test-server conformance note](cma-test-server-conformance.md#sanitized-direct-observations)
records another `1 -> 2/1` publication. These observations support the fake's
normal mode, but they do not establish an exact `+1` guarantee.

### Policy implication

The evidence supports an exact ownership boundary without requiring a fixed
version increment. The provider accepts a positive exact version returned by a
validated draft mutation and confirms publication only when `publishedVersion`
equals the submitted version and current `version` is greater. A GET can reveal
what happened after an ambiguous response, but matching fields or a newly
observed version cannot establish ownership of that draft.

The [Entry publication contract](../design/terraform-value-semantics.md#entry-publication-ownership-and-partial-field-ownership)
defines checkpointing, marker preservation and revocation, response consistency,
and recovery. The [HTTP retry policy](../design/contentful-http-retry-policy.md)
defines the no-replay boundary that preserves the returned draft's provenance.

## Metadata ordering and duplicate behavior

Reviewed 2026-09-03 against Contentful's current published references and
`contentful-management.js` v12.15.0 commit
[`cc096a3`](https://github.com/contentful/contentful-management.js/tree/cc096a337f0e1db6114e8da645d69bb6eb90f11c).

The [Entry](https://www.contentful.com/developers/docs/references/content-management-api/entries/),
[Tags](https://www.contentful.com/developers/docs/references/content-management-api/tags/#tags-on-entries-and-assets),
and [Taxonomy](https://www.contentful.com/developers/docs/references/content-management-api/taxonomy/#concepts-on-entries)
references describe both properties as lists of links, but do not define order
or repeated-link semantics. The first-party client models both as arrays in
[`MetadataProps`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/common-types.ts#L486-L489),
[sends supplied metadata unchanged](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L155),
and [wraps returned data](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/entities/entry.ts#L64-L68)
without normalizing it. That is evidence of client behavior, not a CMA
preservation guarantee.

### Sanitized direct observations

The probe used separate disposable Entries and activated Content Types for
three private tags and three concepts in a disposable concept scheme. Symbols
identify resources by creation order. For each property, the first unique
submission created the Entry; the remaining three submissions used full-body PUT
at the exact current version. Each mutation was followed by GET, whole-Entry
Publish, and another GET.

| Property | Unique submissions | Duplicate submissions | Immediate responses | Later GET and Publish responses |
| --- | --- | --- | --- | --- |
| Tags | `T1,T2,T3`; `T3,T1,T2` | `T1,T2,T1`; `T2,T1,T2` | Echoed each submission | `T1,T2,T3` after unique; `T1,T2` after duplicate |
| Concepts | `C1,C2,C3`; `C3,C1,C2` | `C1,C2,C1`; `C2,C1,C2` | Echoed each submission | `C1,C2,C3` after unique; `C1,C2` after duplicate |

All eight mutations succeeded. For both properties, mutation responses echoed
submitted order and duplicates;
GET and Publish restored the first assignment order and reduced repeated links
to one occurrence. Neither characteristic was therefore durable. The matching
tag and concept results were measured independently. They do not establish a
documented set contract or a general canonical-order rule.

The probe covered one account on the standard global CMA endpoint, three private
tags, and three concepts in one scheme. It did not cover other endpoints, public
tags, Assets, PATCH, locale-based publishing, concurrent mutations, or other
collection sizes. No credential or real identifier is retained here. All
disposable objects were removed; individual GETs returned HTTP 404 and a
collection scan found no matching identifier prefix. No non-disposable resource
or access-control setting was modified.

## Recovery limitation

An incomplete or contradictory successful draft response, or process loss before
Terraform persists both resource state and resource private state, can leave a
matching unpublished draft. Terraform must not claim that draft later merely
because its fields match the plan. This limitation extends beyond transport and
server errors: publication authority depends on the validated response and
persisted checkpoint, not on subsequently observed similarity.
