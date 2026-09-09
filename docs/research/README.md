# Contentful API research

These references describe Contentful resource models, HTTP contracts, lifecycle
behavior, and the evidence supporting them. They are for engineers integrating with
Contentful, operating those integrations, or evaluating API behavior. Their conclusions
apply independently of whether a Terraform resource or any other client implements the
API.

## Reading map

| Topic | Reference and questions covered |
| --- | --- |
| App Framework | [Resource relationships and API inventory](app-framework/README.md): ownership, installation scope, identity, transport, and collection forms |
| App configuration | [Definitions, installations, and sharing](app-framework/configuration.md): AppDefinition, AppInstallation, AppDetails, AppAccessGrant, and installation eligibility |
| App deployment | [Uploads, bundles, and activation](app-framework/deployment.md): AppUpload, AppBundle, selected bundles, and Function deployment |
| App Events | [Subscriptions, topics, and payloads](app-framework/events.md): HTTP targets, Function roles, topic validation, and event data |
| App Actions | [Definitions, invocation, and outcomes](app-framework/actions.md): parameter schemas, calls, status, and raw responses |
| App Identity | [Keys and access tokens](app-framework/identity.md): AppKey, AppAccessToken, credential exchange, and installation scope |
| App request signing | [Secrets, signed requests, and verification](app-framework/request-signing.md): AppSigningSecret, AppSignedRequest, rotation, and reconstruction limits |
| Native external references | [Providers, types, and resources](app-framework/native-external-references.md): ResourceProvider, ResourceType, Resource, and Function resolution |
| Functions and availability | [Runtime, logs, usage, and availability](app-framework/functions.md): execution limits, plan distinctions, log shapes, and usage queries |
| Webhooks | [Configuration, delivery, and observability](webhooks.md): topics, headers, Basic authentication, signing settings, retries, and activity logs |
| Entries and Content Types | [PUT header semantics](entry-and-content-type-put-headers.md): create/update selection and absent-target behavior |
| Entry lifecycle | [Publication, versions, and deletion](entry-lifecycle.md): version locking, published versions, unpublish, and delete preconditions |
| Entry fields | [Null and omission](entry-fields.md): defaults, full-body replacement, raw null, and localized null |
| Entry metadata | [Tag and concept ordering](entry-metadata.md): immediate versus later representations and duplicate links |
| Taxonomy | [Concept and concept scheme versions](taxonomy.md): required headers, validation, conflicts, and deletion |
| Delivery API keys | [Environment selection and versions](delivery-api-keys.md): omission, empty arrays, null, Preview API key relationships, and locking |
| UI Extensions | [Source values](ui-extensions.md): `src`, `srcdoc`, empty sources, and switching |
| Editor Interfaces | [Sidebar values](editor-interfaces.md): absent versus explicit `disabled: false` |
| Environments | [Creation readiness](environment-readiness.md): asynchronous status and completion boundaries |
| Space Enablements | [Request values](space-enablements.md): routes, default-document creation, coupled features, and validation |
| Content preview | [Preview environments](content-preview-environments.md): space-level URL configuration, merge/disable behavior, selected IDs, and concurrency |
| Custom preview tokens | [Live preview variables](live-preview-variables.md): environment-level storage, locale values, normalization, and versioning |
| Collections and errors | [Representations and endpoint differences](collections-and-errors.md): CMA/UMA pagination and error envelopes |

Use the official [Content Management API
reference](https://www.contentful.com/developers/docs/references/content-management-api/)
for its full published surface and the [User Management API
reference](https://www.contentful.com/developers/docs/references/user-management-api/)
for organization-team contracts. This corpus is focused research, not an exhaustive API
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

For mutable published sources, keep useful review dates in compact source metadata,
especially when documenting changing limits or source disagreements. A pinned code
revision identifies source evidence without a separate evaluation date. Editorial and
re-examination dates belong in Git history. Neither an edit nor a mock test refreshes
service evidence. Dates that form part of a contract, such as a sunset deadline, remain
in the relevant contract. Keep contradictory sources visible at the affected operation
and state which questions remain unverified.

Examples and retained observations exclude tenant identifiers, credentials and secret
fragments, names, content, URLs, local paths, actor/timestamp metadata, inventory, usage,
region, billing plans, and entitlement outcomes. This includes linked traces and error details
that could expose the same information indirectly. Use typed path placeholders,
synthetic resource names, and reserved `example.invalid` URLs. Preserve the
relationships, value types, omitted/null/ empty distinctions, status codes, and version
comparisons needed to reproduce the reasoning. The built-in `master` identifier,
documented public plan limits, experiment counts, observation dates, and public source
revisions are not tenant inventory. Locale examples must be synthetic or describe a
configured locale without disclosing the original locale inventory.

## Document organization

Keep one authoritative account of each API behavior. Name files for the entity or
concept they document; use headings to delimit focused coverage. Split when distinct
reader questions, representations, or lifecycle boundaries benefit from separate
references. Do not split every entity automatically or retain a split solely because
its observations came from separate experiments.

A focused reference leads with its subject, then states scope and evidence, addressing
and representation, mutation/lifecycle behavior, and unresolved cases as relevant. Keep
observation matrices near the behavior they explain. Group related references under a
family directory when a shared relationship map is useful. This README owns corpus
conventions; a family overview owns its shared study scope and sources. Individual
references link to that context and retain their specific limitations, definitions, and
citations. Summaries may orient readers, but detailed contract explanations belong in
one reference. Update this index and inbound links when moving material.

Provider schemas, implementation status, private state, recovery policy, test coverage,
and mock conventions belong in [provider design](../design/README.md). The [CMA
test-server conformance reference](../design/cma-test-server-conformance.md) links those
choices back to API evidence. Practitioner configuration belongs in [provider
documentation](../index.md). These separate audiences can share links without
duplicating each other's contracts or turning research into a work plan.
