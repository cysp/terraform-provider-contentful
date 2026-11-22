# Contentful API research

Use these notes to check Contentful request shapes, version preconditions, and observed lifecycle behavior. For Terraform support, see the [provider reference](../index.md). For decisions based on this evidence, see [provider design](../design/README.md).

## Reading map

| Topic | Reference and questions covered |
| --- | --- |
| App Framework | [App Framework reference](app-framework/README.md): definitions, installations, deployment, events, actions, identity, request signing, Functions, and external resources |
| Signing secret metadata | [Versions and timestamps](signing-secret-versions.md): AppSigningSecret and WebhookSigningSecret metadata, conflicting examples, and conditional requests |
| Webhooks | [Configuration, delivery, and observability](webhooks.md): topics, headers, Basic authentication, signing settings, retries, and activity logs |
| Entries and Content Types | [PUT header semantics](entry-and-content-type-put-headers.md): create/update selection and absent-target behavior |
| Entry lifecycle | [Publication, versions, and deletion](entry-lifecycle.md): version locking, published versions, unpublish, and delete preconditions |
| Entry fields | [Null and omission](entry-fields.md): defaults, full-body replacement, raw null, and localized null |
| Entry metadata | [Tag and concept ordering](entry-metadata.md): immediate versus later representations and duplicate links |
| Taxonomy | [Concept representations, pagination, and versions](taxonomy.md): response fields, relationships, cursor traversal, required headers, conflicts, and deletion |
| Delivery API keys | [Environment selection and versions](delivery-api-keys.md): omission, empty arrays, null, Preview API key relationships, and locking |
| UI Extensions | [Source values](ui-extensions.md): `src`, `srcdoc`, empty sources, and switching |
| Editor Interfaces | [Response identity and sidebar values](editor-interfaces.md): list/detail metadata, collection envelope, and absent versus explicit `disabled: false` |
| UI Config | [Identity, configuration, and updates](ui-config.md): metadata discrepancies, views, Timeline, preview-mode mutations, and omission limits |
| Configuration discovery | [Spaces and scoped configuration endpoints](configuration-discovery.md): detail/list paths, organization scope, pagination evidence, and response discrepancies |
| Locales | [Identity, configuration, and lifecycle](locales.md): identity, discovery, code changes, API flags, fallbacks, and publication |
| Space access | [Roles and memberships](space-access.md): permission unions, direct versus aggregate memberships, and static list/detail comparisons |
| Organization settings | [Metadata, enablements, and access configuration](organization-settings.md): preferences, contacts, invitation lifecycle, and access-policy reads |
| Embargoed assets | [Protection modes and asset keys](embargoed-assets.md): space-wide protection, publication across environments, and signing-key lifetime |
| Environments | [Creation readiness](environment-readiness.md): asynchronous status and completion boundaries |
| Environment aliases | [Routing and response identity](environment-aliases.md): alias/target comparisons and returned environment links |
| Space Enablements | [Request values](space-enablements.md): routes, default-document creation, coupled features, and validation |
| Content preview | [Preview environments](content-preview-environments.md): space-level URL configuration, merge/disable behavior, selected IDs, and concurrency |
| Custom preview tokens | [Live preview variables](live-preview-variables.md): environment-level storage, locale values, normalization, and versioning |
| Collections and errors | [Representations and endpoint differences](collections-and-errors.md): CMA/UMA pagination and error envelopes |

Use the official [Content Management API
reference](https://www.contentful.com/developers/docs/references/content-management-api/)
for its full published surface and the [User Management API
reference](https://www.contentful.com/developers/docs/references/user-management-api/)
for organization-team contracts. These notes contain focused research, not an exhaustive API
specification. A missing operation is not proof that it does not exist.

## Terminology and scope

**Content Management API (CMA)** names the management HTTP API. **Content Preview API
(CPA)** serves unpublished content; preview environments configure preview URLs; live
preview variables store custom preview token values. These are separate interfaces
despite their shared preview terminology.

Use Contentful's resource and property names, including AppDefinition,
AppEventSubscription, Content Type, Entry, `sys.version`, and `sys.publishedVersion`. In
a scoped version table, `version` and `publishedVersion` refer to those `sys` members.
Use *activation* for a Content Type and *publication* for an Entry; the shared
`/published` suffix does not make them the same resource operation. *Delete* describes
an HTTP resource operation; Terraform *destroy* is a client workflow.

Paths are relative to the named API host. Braced snake_case names such as `{space_id}`
are path placeholders; exact JSON properties and SDK argument names retain their native
spelling. The App Framework overview defines its abbreviated paths. A Contentful Link is
a JSON reference with `sys.type: "Link"`, `sys.linkType`, and `sys.id`. `Link<T>`
abbreviates a Link whose `sys.linkType` is `T`; the enclosing field determines the
relationship. A resource's addressing identity must be determined from its endpoint,
not inferred solely from the presence of `sys.id`.

## Evidence and redaction

Record endpoint behavior and its sources in the relevant research note. Keep decisions about how the provider uses that behavior in [provider design](../design/README.md), and deliberate fixture behavior in [test-server conformance](../design/cma-test-server-conformance.md).

| Evidence class | What it establishes |
| --- | --- |
| Published Contentful reference | The contract stated by the cited page; examples and prose can disagree. |
| Pinned first-party source | Paths, types, headers, and client behavior at that revision; a type is not proof of server acceptance. |
| Direct observation | Results for the exact requests and lifecycle exercised, with an observation date when recorded. |
| Interpretation or unresolved behavior | A conclusion or limit explicitly distinguished from the underlying evidence. |

Keep source citations beside the facts they support and pin source-code references to
inspected commits. Record known experiment dates once per evidence set, next to its
scope or observation matrix. If the date is unknown, mark it as unrecorded; retain a
record date only when it supplies otherwise unavailable provenance. Never infer an
experiment date from a documentation commit.

Date mutable sources where limits or contracts may change. Keep disagreements and unverified cases beside the affected operation. An editorial change or mock test does not refresh service evidence.

Examples and retained observations exclude tenant identifiers, credentials and secret
fragments, names, content, URLs, local paths, actor/timestamp metadata, inventory, usage,
region, billing plans, and entitlement outcomes. This includes linked traces and error details
that could expose the same information indirectly. Use typed path placeholders,
synthetic resource names, and reserved `example.invalid` URLs. Preserve the
relationships, value types, distinctions between omission, null, and empty values,
status codes, and version comparisons needed to reproduce the reasoning. The built-in `master` identifier,
documented public plan limits, experiment counts, observation dates, and public source
revisions are not tenant inventory. Locale examples must be synthetic or describe a
configured locale without disclosing the original locale inventory.
