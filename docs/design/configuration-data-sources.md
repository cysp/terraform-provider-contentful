# Configuration data sources

Space, Environment, Environment Alias, and Locale data sources provide read-only
reference and discovery. Singular forms require addressing IDs. Plural forms use
scoped offset collections and return lists sorted by the entity's addressing ID.
The [practitioner guide](../guides/existing-configuration.md) supplies composition
and exact-name/code/default selection patterns.

## Identity and projection

A singular `id` joins its endpoint IDs with `/`; collection IDs contain only
addressing scope. The all-accessible Spaces view uses `spaces`, and organization
scope uses `organizations/{organization_id}/spaces`. Credentials and mutable
names, targets, codes, statuses, and versions never enter lookup identity.
Input IDs must be known and non-null at Read, nonempty, contain no `/`, and differ
from `.` and `..`. These bounds protect slash-joined identity and HTTP routing;
they are not a declaration of Contentful's complete ID grammar. The generated
client escapes path segments and preserves the configured API base path.

Singular returned IDs and requested Space or organization scope must match the
request. Computed references and collection item IDs preserve decoded values,
even when those values would be invalid as inputs to a later lookup. Projection
does not reapply request-ID validation. Locale environment links must echo the
addressed Environment/alias context. Other links are diagnosed as unsupported
response identity, rather than proof that the service routed incorrectly. This
bound follows the retained [Locale/alias observations](../research/environment-aliases.md).
It introduces no alias resolution or parent preflight requests.

Locale resource ID, content code, and fallback code remain distinct. Explicit
null or omitted `fallbackCode` becomes Terraform null; an empty string remains
empty. No `internal_code`, audit metadata, billing metadata, or credentials enter
these schemas. Other advertised Locale fields are required by the curated
response decoder. The [Locale evidence](../research/locales.md) separates raw
observations from SDK projection.

Environment and Environment Alias use the same generated HTTP operations and
entity decoders as their managed counterparts. Data-source schemas and timeouts
are separate from resource planning and lifecycle ownership; singular and plural
sources share each family's item projection. Environment data sources expose the
optional decoded `aliasedEnvironment` link as `aliased_environment_id`. It never
replaces the addressing ID. Both published
`Environment Alias` and established `EnvironmentAlias` type spellings decode.

General Environment reads do not poll or classify statuses as failures. The
readiness waiter owns polling, failure classification, and its timeout contract.
The curated Environment decoder requires `name` and `sys.status`:
the [published detail example][environment-detail] omits both, while the pinned
SDK and [alias reference][aliases] include them. A reduced response fails decoding;
no defaults are synthesized. This discrepancy remains an evidence limit, not a
live-verified assertion that every collection/detail response contains those fields.

## Offset traversal

Teams, Spaces, Environments, Environment Aliases, and Locales share one collection
reader. Each endpoint decodes its generated response into a typed collection or
diagnostics before traversal. The collection element type and item projection
input must agree at compile time. As with
Terraform list resources, traversal advances by the number of returned items and
stops on an empty page or when the latest reported total is reached. Generated
collection decoders require `sys` and `items`; pagination metadata is optional.
Returned `skip` and `limit` do not control progress. Changing totals and duplicate
IDs are accepted, and duplicates are retained in the sorted result.

Empty successful collections produce empty lists. Request, decoding, and
projection errors abort the lookup without publishing partial state; not-found
and permission errors are not reclassified as empty inventories.

All pages share the data source's operation timeout and the existing
[read retry policy](contentful-http-retry-policy.md). No cursor or continuation
URLs are followed, and no N+1 detail reads fill list results. Offset support is
grounded in the [family endpoint evidence](../research/configuration-discovery.md).

Offset traversal does not establish a snapshot. Concurrent edits can cause
omissions or duplicates, and alias reads do not establish uninterrupted target
stability. A single-read Environment alias projection avoids a resolution race
but does not promise stability after that response.

## Verification boundary

Raw fixture tests check literal methods, paths, headers, page progression,
projection, error handling, and publication. Mocked Terraform tests cover all
eight sources, apply-time unknown inputs, stable repeat plans, localized Entry
composition, and cardinality postconditions that stop dependent mutations. Shared
decoder and resource/waiter regression checks guard compatibility. The in-process
server's explicit fixture stores and alias-context projection model these tested
contracts; they do not establish live endpoint conformance or access policy.

The read-only live acceptance test uses the shared acceptance Space to exercise
Space, Environment, and Locale detail and collection reads, plus the Environment
Alias collection. It checks collection membership and Locale projection without
assuming exact inventory sizes. Alias detail and alias-routed reads require an
existing alias fixture and are covered by the mocked tests. The live test does
not establish multipage behavior or stability during concurrent changes.

[environment-detail]: https://www.contentful.com/developers/docs/references/content-management-api/environments/get-an-environment/
[aliases]: https://www.contentful.com/developers/docs/references/content-management-api/environment-aliases/#new-environment-properties-when-using-aliases
