# Contentful CMA-manageable settings: validated findings

Date observed: 2026-07-14

## Purpose

This document catalogues settings and resource families exposed by the Contentful web app that are not all represented by the provider's currently registered Terraform resources. Its findings were adversarially checked against:

- read-only observation of the signed-in Contentful web app and Safari Web Inspector network traffic;
- the provider registrations in [`internal/provider/contentful_provider.go`](../../internal/provider/contentful_provider.go);
- first-party Contentful API and Help Center documentation available on the observation date; and
- the provider's generated-client OpenAPI description.

This is the final evidence catalogue, not an assertion that every undocumented web-app endpoint is a supported public API. Most observations were read-only. Two low-impact settings were changed and immediately restored to characterize their write contracts. No access-control, membership, SSO, deletion, billing, secret, embargoed-asset, or publishing setting was changed.

All organization, space, environment, user, account, and content identifiers are symbolic placeholders. Examples preserve exact API field names and protocol enum values only where their spelling is part of the finding. No response bodies containing customer or personal data are reproduced.

## Standard of proof and evidence tiers

Conclusions use deliberately narrow standards:

- **Confirmed public contract** means a current first-party API reference or checked-in source directly supports the stated API claim.
- **Web-app implementation** means the behavior is implemented in a current first-party Contentful web-app bundle. It describes what the web app does, not a supported third-party API contract.
- **Observed live behavior** means direct web-app evidence supports the stated technical fact without establishing a public third-party contract.
- **Unsupported as a public contract** means the web app demonstrates technical behavior but no matching public API contract was established.
- **Ambiguous** means the evidence establishes only part of the scope, lifecycle, authentication, or mutation semantics.

| Tier | Meaning |
| --- | --- |
| A — publicly documented | Contentful documents a public API contract for the resource or setting. This is the strongest basis for provider support. |
| B — web-app implementation or live observation | The first-party web app implements or calls an endpoint, but no matching public CMA reference was found. Treat public support and stability as unknown until Contentful documents or confirms them. |
| C — read-only/supporting | The endpoint appears to supply entitlement, enforcement, membership aggregation, billing, license, or usage context. Observation alone does not establish a manageable resource. |

“Public docs not found” is bounded to the current [CMA reference](https://www.contentful.com/developers/docs/references/content-management-api/), [User Management API reference](https://www.contentful.com/developers/docs/references/user-management-api/), and first-party Contentful documentation reviewed on the date above. It does not prove that an unpublished or separately contracted API does not exist.

## Provider and generated-client baseline

The provider currently registers these 21 Terraform resource types:

`app_definition`, `app_signing_secret`, `app_installation`, `content_type`, `delivery_api_key`, `editor_interface`, `environment_alias`, `environment`, `entry`, `extension`, `personal_access_token`, `resource_provider`, `resource_type`, `role`, `space_enablements`, `tag`, `taxonomy_concept`, `taxonomy_concept_scheme`, `team`, `team_space_membership`, and `webhook`.

The checked-in [`openapi.yml`](../../internal/contentful-management-go/openapi/openapi.yml) does not contain the investigated locale, UI Config, security-contact, organization-metadata, direct-membership, preview-configuration, access-policy, identity-provider, or embargoed-mode paths. It does contain other organization-scoped resource families. This is a generated-client implementation gap, not evidence that the confirmed public APIs are absent.

The primary-environment settings menu exposed:

`Locales`, `Tags`, `Home`, `Content preview`, `General settings`, `Users`, `Roles`, `Environments`, `API keys`, `CMA tokens`, `Embargoed assets`, `Webhooks`, and `Usage`.

The non-primary-environment menu exposed only `Locales`, `Tags`, `Home`, `Content preview`, `Environments`, `API keys`, and `CMA tokens`. `General settings`, `Users`, `Roles`, `Embargoed assets`, `Webhooks`, and `Usage` were absent. This is UI evidence that the omitted panes are not environment-local; it is not by itself an API contract.

The organization settings UI also exposed mutable organization name, granular environment permissions, default language, security-notification email addresses, and SSO configuration.

## Catalogue by settings surface

| Web-app surface | API/resource interpretation | Evidence | Verdict | Provider coverage | Assessment |
| --- | --- | --- | --- | --- | --- |
| Locales | Environment-scoped locale resources | A | Confirmed public contract | Missing | Strong candidate. Public CMA supports listing, creation, update, and deletion of locales. |
| Tags | Environment-scoped tags | A | Confirmed public contract | Covered by `contentful_tag` | No additional resource implied. |
| Home | App-provided/built-in home widgets, app installations, and UI config | A, plus live mapping | Confirmed public UI Config contract; observed Home mapping | Partly covered by `app_installation` and `extension`; UI config missing | The documented UI Config includes `homeViews` entries. A dedicated UI-config resource would cover the environment-visible arrangement; app installation/extension lifecycle is already represented. |
| Content preview | Preview-environment configuration | B | Web-app lifecycle implementation and observed update; unsupported as a public contract | Missing | The web app implements create, update, delete, and rank operations. Only update and rollback were exercised live; external authentication and API support remain unverified. |
| General settings | Space object, especially mutable space name | A | Confirmed public contract, narrow metadata scope | Missing | Strong candidate for a space resource or narrowly scoped space-settings resource. Public CMA documents space creation, name update, deletion, and unarchive. |
| Users | Space memberships are mutable; space members are an aggregate view | A and C | Confirmed public contracts with distinct lifecycles | Direct user/space membership missing; team-space membership covered | Model `space_memberships`, not the read-only `space_members` aggregation. |
| Organization members | Organization memberships created indirectly through invitations | A | Unsupported as conventional CRUD | Missing | Public list/get/update/delete operations exist, but a membership cannot be created directly. Model invitation and acceptance or support only import/update/delete. |
| Roles | Space roles and policies | A | Confirmed public contract | Covered by `contentful_role` | No additional top-level resource implied. |
| Environments | Environments and aliases | A | Confirmed public contract | Covered | No additional top-level resource implied by the menu item. |
| API keys | Delivery/preview API-key pairs | A | Confirmed public contract | Delivery key covered; preview key is a data source | Existing coverage is substantially aligned. |
| CMA tokens | Personal access tokens | A | Confirmed public contract | Covered by `contentful_personal_access_token` | Existing coverage is aligned. |
| Embargoed assets | Space-level protection mode; asset keys are short-lived signing credentials | B, plus public feature docs and asset-key API | Web-app mutation implementation; unsupported as a public protection-mode contract | Missing | The web app uses a space singleton with a narrow protection-mode body. Do not equate it with the documented asset-key endpoint. |
| Webhooks | Space-scoped webhook definitions | A | Confirmed public contract | Covered by `contentful_webhook` | Existing coverage is aligned. |
| Usage | Usage, entitlements, enforcement, account, and license views | C | Observed read-only/supporting behavior | Missing | Treat as informational unless a documented mutation contract and a stable Terraform lifecycle emerge. |
| Organization name | Organization metadata | A for organization retrieval/update; live UI | Confirmed public contract, narrow metadata scope | Missing | Possible organization-settings singleton. Public evidence supports metadata read/update, not full organization lifecycle. |
| Organization access policy | Organization-wide access controls, including granular environment permissions | B; public Help Center describes some behavior | Web-app partial-update implementation; unsupported as a public contract | Missing | Two independent web-app flows use narrow versioned `PUT` bodies. This is security-sensitive shared state, and arbitrary omission semantics remain unverified. |
| Default language | Organization-level default for the web app and email communications | B | Observed UI behavior; unsupported as a public contract | Missing | Users can override it in their profiles. Candidate only after its endpoint and update contract are captured and supported; it is not a space environment locale. |
| Security notification emails | Organization security contacts | A | Confirmed public contract | Missing | Strong candidate. Public CMA documents get, create, update, and delete operations for organization security contacts. |
| SSO | Organization identity provider, certificates, and access-policy state | B | Web-app create/update implementation and observed absence behavior; unsupported as a public contract | Missing | The web app splits identity-provider, certificate, and MFA mutations across separate resources. The live GET returned `404` when unset; no delete call was found. |

## Non-primary-environment path survey

Read-only Safari Web Inspector captures from a non-primary environment confirmed that the environment shown in the web-app route is also present in the CMA paths for environment-local resources; the environment was not merely client-side navigation context.

All paths below are symbolic. Query values are retained only when they describe the observed collection request rather than a customer identifier.

| Web-app surface or request role | Live-observed CMA path | Scope conclusion |
| --- | --- | --- |
| Locales pane | `/spaces/{space_id}/environments/{environment_id}/locales?limit=100&skip=0` | Explicitly environment-scoped |
| Tags pane | `/spaces/{space_id}/environments/{environment_id}/tags?limit=1000&skip=0` | Explicitly environment-scoped |
| Home pane | `/spaces/{space_id}/environments/{environment_id}/ui_config` | Explicitly environment-scoped |
| Content preview pane: shared UI settings | `/spaces/{space_id}/environments/{environment_id}/ui_config` | Explicitly environment-scoped |
| Content preview pane: preview definitions | `/spaces/{space_id}/preview_environments?limit=100` | Space-scoped; selecting a non-primary environment did not add an environment path component |
| Shared settings-page bootstrap: content types | `/spaces/{space_id}/environments/{environment_id}/public/content_types?limit=1000` | Explicitly environment-scoped; note the observed `public` path component |
| Shared settings-page bootstrap: editor interfaces | `/spaces/{space_id}/environments/{environment_id}/editor_interfaces` | Explicitly environment-scoped |
| Shared settings-page bootstrap: resources | `/spaces/{space_id}/environments/{environment_id}/resources` | Explicitly environment-scoped |
| Shared settings-page bootstrap: app installations | `/spaces/{space_id}/environments/{environment_id}/app_installations` | Explicitly environment-scoped |
| Shared settings-page bootstrap: extensions | `/spaces/{space_id}/environments/{environment_id}/extensions` | Explicitly environment-scoped |

The bootstrap requests were observed while loading an environment settings page, but they are not all owned by that pane. The web app preloads navigation, editor, app, and entitlement context. Their presence establishes environment-aware read paths, not a new settings resource or a pane-specific mutation contract.

A reversible UI Config mutation subsequently confirmed that the environment-explicit non-primary path accepts `PUT`. No other path in the table was mutated. The survey does not establish whether the observed `public/content_types` route is a supported third-party contract or whether a personal access token can call undocumented routes used by the web app.

## Reversible mutation evidence

### Preview mode through primary and non-primary UI Config routes

Reversible preview-mode changes succeeded in both the primary and a non-primary environment, and restoration was verified in the UI. The web app used the shortened path for the primary environment and the documented environment-explicit path for the non-primary environment:

```http
PUT /spaces/{space_id}/ui_config
PUT /spaces/{space_id}/environments/{environment_id}/ui_config
Content-Type: application/vnd.contentful.management.v1+json
X-Contentful-Version: {current_version}
```

The web app sent the complete shared UI Config document, not a one-field patch. That establishes observed client behavior, but not server-side replacement semantics; treatment of an omitted property remains unverified. The relevant value changed between:

```json
{"livePreview":{"previewMode":"legacyPreview"}}
```

and:

```json
{"livePreview":{"previewMode":"livePreview"}}
```

The same document also contained publishing mode, Home views, and entry-list views. This reinforces the ownership concern for Terraform: a conservative client should fetch, merge, and update the document so it preserves unknown and unmanaged fields.

The non-primary change advanced the UI Config version, and restoration used that successor version. The tested preview-mode effect was environment-specific: after a full reload, the primary environment retained its original mode while the non-primary environment retained the temporary mode. This differs from the dialog's all-environments warning and is a bounded two-environment result, not proof about every Contentful account or future web-app version.

The shortened primary path may be a compatibility route for the primary environment. Provider work should use the documented environment-explicit endpoint rather than depend on that alias.

### Preview-environment update

An existing preview platform description was changed temporarily and restored to its exact original value. Both `PUT` requests returned HTTP 200, and the final UI state was verified.

The web app used:

```http
PUT /spaces/{space_id}/preview_environments/{preview_environment_id}
Content-Type: application/vnd.contentful.management.v1+json
X-Contentful-Version: {current_version}
```

The request body was a full replacement-shaped document containing:

```text
{
  "name": "{name}",
  "description": "{description}",
  "configurations": [
    {
      "url": "{preview_url_template}",
      "entityType": "ContentType",
      "entityId": "{content_type_id}",
      "enabled": {enabled_boolean},
      "example": {example_boolean}
    }
  ]
}
```

Multiple content types using the same URL were represented as separate configuration objects. Responses additionally included `contentType` and `sys` metadata.

Each update sent the current resource version and returned its successor; the restoration used the version returned by the preceding change. This establishes optimistic versioning for updates without retaining concrete observed values, but does not establish a supported public contract. Browser-session authorization, server validation, error behavior, and personal-access-token compatibility remain unverified. Create, delete, and ordering behavior were established only from the web-app implementation described below.

## Mutation semantics by configuration object

The CMA overview recommends fetching a resource, modifying it, and sending the complete expected update body. The public SDK and live web app show complete-object behavior for UI Config. That does not independently establish how the UI Config server treats omitted fields, and it must not be projected onto undocumented endpoints where the first-party client demonstrably uses a partial body or a purpose-built action.

| Configuration object | Mutation shape | Concurrency | Established behavior |
| --- | --- | --- | --- |
| UI Config | Environment singleton `PUT`; public SDK sends the complete object without `sys` | `X-Contentful-Version` | Complete-object client behavior; live primary and non-primary update/rollback confirmed; server omission semantics unverified |
| Preview environment | Collection `POST`; item `PUT` and `DELETE`; collection rank `PATCH` | Version header on item update; per-item versions in rank body; no delete version in inspected client | Undocumented web-app implementation; live item update/rollback confirmed |
| Organization access policy | Singleton `PUT` with narrow bodies used by independent UI flows | `X-Contentful-Version` | Undocumented partial-update behavior for verified fields only |
| Organization identity provider | Singleton `POST`, `PUT`, and `PATCH` | Current version on existing-object `PUT` and `PATCH` | Undocumented create/update implementation; no delete call found |
| Identity-provider certificate | Collection `POST`; item `PUT` and `PATCH` | Current version on existing-object mutation | Separate from identity-provider and MFA writes; no delete call found |
| Embargoed-assets protection | Space singleton `PUT` with exactly `{ "protectionMode": value }` | No version supplied by inspected client | Undocumented web-app implementation; high-impact confirmation dialog observed in UI |
| Locale | Environment collection `POST`; versioned item `PUT`; item `DELETE` | Version header on update; no delete version in public SDK | Public lifecycle; referenced-fallback deletion is multi-resource orchestration |

### UI Config complete-object client behavior

The public management SDK deep-copies the supplied UI Config, removes `sys`, and sends the remaining object with the current version. Its example modifies a fetched object rather than constructing a one-field request. The provider should therefore preserve unknown mutable fields and avoid claiming exclusive ownership of this shared document. The live client also sent the complete shared document, including unrelated Home, publishing, and entry-list configuration. Server behavior for an intentionally omitted field remains unverified. [UI Config endpoint source](https://github.com/contentful/contentful-management.js/blob/main/lib/adapters/REST/endpoints/ui-config.ts) [UI Config API source](https://github.com/contentful/contentful-management.js/blob/main/lib/create-ui-config-api.ts)

### Preview-environment lifecycle and rank action

The first-party web app implements the following undocumented surface:

```text
POST   /spaces/{space_id}/preview_environments
PUT    /spaces/{space_id}/preview_environments/{preview_environment_id}
DELETE /spaces/{space_id}/preview_environments/{preview_environment_id}
PATCH  /spaces/{space_id}/preview_environments/rank
```

Create sends mutable fields without `sys`. Update removes `sys` and sends the current version as a header. Delete supplies no version in the inspected client. Configuration rows omit resolved UI-only `contentType` objects before transmission.

Rank is an action rather than resource replacement. Its body wraps items containing the object ID, current version, and zero-based rank; concurrency travels per item rather than in a request header. Client-side URL and configuration validation does not establish the server's complete validation contract. [Contentful web app](https://app.contentful.com/)

### Access policy, identity provider, and certificates

The undocumented access-policy helper performs versioned `PUT`. At least two independent UI flows send narrow bodies: `{ "explicitTokenAuthorization": value }` and `{ "mfaEnforcement": value }`. This establishes field-specific partial-update behavior for those flows, not a general merge guarantee for every field. No create or delete call was found.

The identity-provider client exposes unversioned `POST` creation and versioned `PUT` or `PATCH` for an existing object. The setup save flow can issue `PUT`, while edits to an enabled configuration use `PATCH`; the inspected implementation did not establish which preceding server state selects `POST` rather than the setup-flow `PUT`. The client catches every identity-provider read failure and returns absence, so callers must not treat all swallowed failures as `404`.

Certificates are separate objects created with `{ "content": value }` and updated through versioned `PUT` or `PATCH`. The form removes certificate content and `mfaEnforcement` before writing the identity provider: certificate content goes to the certificate endpoint, and MFA enforcement goes to access policy. Connection testing calls a separate authentication service and is not a CMA mutation. No identity-provider or certificate delete call was found. [Contentful web app](https://app.contentful.com/)

### Embargoed-assets protection mode

The web app uses a space-scoped singleton:

```text
GET /spaces/{space_id}/embargoed_assets
PUT /spaces/{space_id}/embargoed_assets
```

The `PUT` body is exactly `{ "protectionMode": value }`, and the inspected client supplies no version. For display it treats an absent mode as `disabled`; on write it maps the disabled UI state to JSON `null`, the enabled UI state to `migrating`, and passes other selected modes through. These are undocumented client mappings, not a public wire contract. Asset keys remain a separate public, environment-scoped, short-lived credential resource. [Contentful web app](https://app.contentful.com/) [Asset keys](https://www.contentful.com/developers/docs/references/content-management-api/asset-keys/)

### Locale fallback deletion

The public SDK sends a complete mutable locale representation on update, removing `sys`, the read-only `default` property, and internal fields. The web app cannot directly delete a locale referenced as another locale's fallback. It first updates every dependent locale with its own current version, then deletes the target. The dependent updates run concurrently and no transaction or rollback is evident, so partial failure can leave some dependents changed while the target remains. This should be modeled as a multi-resource transition with explicit recovery, not as one atomic delete. [Locale endpoint source](https://github.com/contentful/contentful-management.js/blob/main/lib/adapters/REST/endpoints/locale.ts) [Locales](https://www.contentful.com/developers/docs/references/content-management-api/locales/) [Contentful web app](https://app.contentful.com/)

### Unresolved organization preference mutation

The public organization representation contains `defaultUserLanguage`, and the public CMA exposes organization update. The exact web-app save request, body ownership, and concurrency behavior for this field were not established. It remains unresolved rather than inheriting semantics from unrelated organization resources. [Organization entity source](https://github.com/contentful/contentful-management.js/blob/main/lib/entities/organization.ts) [Organizations](https://www.contentful.com/developers/docs/references/content-management-api/organizations/)

## Tier A: publicly documented gaps

### Locales

Contentful publicly documents locale collection and individual-locale operations under the CMA. Locales are environment-aware and are copied when an environment is created. The provider has no registered locale resource in this checkout.

Suggested Terraform shape: an environment-scoped `contentful_locale` resource. Default-locale changes, fallback cycles, and deletion constraints require live lifecycle tests before finalizing the schema.

### Space metadata and lifecycle

The public CMA documents spaces as manageable resources, including updating a space name. This directly backs the mutable name shown under General settings.

A provider resource should avoid accidentally coupling a routine name change to destructive space deletion. A narrowly scoped `contentful_space_settings` resource may be safer than claiming the complete space lifecycle.

### Space memberships

Contentful publicly documents mutable `/spaces/{space_id}/space_memberships` resources. These represent direct user membership/invitation and role assignments. The live web app also called `/spaces/{space_id}/space_members`, but Contentful documents space members as a paginated aggregate of users with access and their related direct or team memberships.

Therefore:

- `space_memberships` are plausible Terraform-managed resources;
- `space_members` are better suited to a data source or internal read model; and
- the existing `team_space_membership` resource covers only team-derived access, not direct user membership.

### Organization memberships

The live web app called:

```text
/organizations/{organization_id}/organization_memberships
```

Contentful's public User Management API documents list, get, update, and delete operations for organization memberships. A membership cannot be created directly: creating an organization invitation produces a pending membership as a side effect. This is a supported API family, although it is documented under the User Management API rather than the main CMA reference.

Provider support would therefore need to model invitations and their transition to accepted membership, or intentionally support only import/update/delete of an existing membership. It also needs a stable identity key that does not place personal email addresses unnecessarily in logs or diagnostics.

### Organization security contacts

The live web app called:

```text
/organizations/{organization_id}/security_contacts
```

The public CMA reference documents get, create, update, and delete operations for organization security contacts. This maps cleanly to the mutable security-notification email list in organization settings.

Because the data is personally identifiable contact information, any provider design should explicitly consider state sensitivity and import behavior.

### Environment UI Config

Contentful publicly documents environment-scoped UI Config get/update operations and describes the configuration as views, folders, Home widgets, preview mode, publishing mode, and timeline settings visible to everyone in an environment.

The live web app called a shortened path for the primary environment:

```text
/spaces/{space_id}/ui_config
```

The non-primary-environment web app called:

```text
/spaces/{space_id}/environments/{environment_id}/ui_config
```

The public contract and the non-primary live observation are therefore aligned on environment scope. Provider work should prefer the documented environment-explicit endpoint rather than depend on the observed shortened primary-environment path, which may be a compatibility alias.

## Tier B: live-observed endpoints without a public contract found

### Preview environments

The Content preview page called:

```text
/spaces/{space_id}/preview_environments?limit={page_limit}
```

The UI exposes preview configuration as mutable, but no matching public CMA reference was found. The similarly named Content Preview API is a content-delivery API and is not documentation for this configuration endpoint.

The web-app implementation establishes create, update, delete, and rank request shapes; update and rollback were also observed live. Before provider implementation, verify supported non-browser authentication, server validation and error behavior, and whether Contentful supports the endpoint for third-party use.

### Embargoed-assets protection mode

Contentful publicly documents that embargoed assets are enabled at space level using the durable modes `preparation`, `unpublished assets protected`, and `all assets protected`. It separately documents creation of short-lived asset keys through Contentful APIs.

The web-app implementation identifies a space-level `embargoed_assets` singleton and its narrow `PUT` body, but no matching public endpoint contract was located. This and asset keys are distinct concerns:

- the protection mode is durable space configuration shown in the UI and mutated by an undocumented web-app endpoint;
- an asset key is an expiring secret used to sign protected URLs and has a public creation API.

Protection-mode management therefore remains in the investigation tier. An asset-key resource is technically better supported, but would still need write-only or ephemeral handling because the returned secret is short-lived.

### Organization access policy

The web app called:

```text
/organizations/{organization_id}/access_policy
```

The observed organization-scoped `GET` returned an access-policy object with fields for SSO enforcement, MFA enforcement, explicit token authorization, token-expiration limits, SCIM-only management, and granular environment policies. The object is broad, security-sensitive shared state rather than a narrow granular-permissions toggle.

The web-app client performs versioned partial-body `PUT` requests for at least explicit-token authorization and MFA enforcement. The public organization update example separately exposes an `accessPolicy.sso` value inside organization system metadata, but it does not document the observed `access_policy` endpoint or its mutation contract.

No matching public CMA endpoint reference was found. The implementation establishes those narrow client mutations, not a supported general merge contract, authentication outside the web app, or safe rollback behavior. Keep this object outside the implementation-ready tier.

### Organization identity provider

The SSO UI mapped to:

```text
/organizations/{organization_id}/identity_provider
```

An observed `GET` returned `404` when no identity provider was configured. The web-app implementation uses `POST`, versioned `PUT`, and versioned `PATCH`, and splits certificate and MFA writes into separate resources. No delete call or matching public CMA endpoint reference was found.

### Organization-level preferences

The UI exposed mutable organization name, granular environment permissions, and default language. Organization retrieval/update is publicly documented, granular environment permissions are described in the Help Center, and the default language is described in the UI as applying to the web app and email communications unless a user overrides it. Access-policy implementation semantics are recorded above; the exact default-language save request remains unresolved.

Further capture should distinguish:

- organization object fields;
- feature enablements or migration flags;
- user-local UI preferences; and
- values computed from license or entitlement state.

## Tier C: read-only and supporting endpoints

The following live-observed requests appear to support navigation, authorization, entitlements, billing, or usage displays rather than establish independently manageable resources:

```text
/spaces/{space_id}/space_entitlement_set
/spaces/{space_id}/enforcements
/spaces/{space_id}/space_members
/organizations/{organization_id}/space_entitlement_sets
```

The session also observed request families containing `customer_accounts`, `licenses`, and `usage`. Their exact customer-specific URLs and response bodies are intentionally not recorded here.

Classification notes:

- `space_members` has a public read API, but mutations should target memberships instead.
- Entitlement sets and licenses likely describe what the account may configure, not desired configuration themselves.
- Enforcements likely explain restrictions derived from plan or policy state.
- Usage is naturally read-only and time-varying; if useful to Terraform, it belongs in data sources or validation diagnostics rather than managed resources.
- Customer-account and billing endpoints may have different authentication, stability, and contractual boundaries from the public CMA.

## Provider readiness assessment

This ordering is an engineering recommendation, not an API fact. It weighs public support, completeness of lifecycle, blast radius, sensitive state, importability, concurrency, and shared-document ownership. User demand has not been measured.

| Readiness | Candidate | Basis | Main constraint |
| --- | --- | --- | --- |
| 1 | Locale | Public CRUD; conventional environment-scoped identity | Default/fallback transitions and deletion constraints |
| 2 | Security contact | Public CRUD; direct organization-settings mapping | Contact data becomes sensitive Terraform state |
| 3 | Space name/settings singleton | Public read/update; narrow schema | Must not imply ownership of space deletion |
| 4 | Direct space membership | Public CRUD/invite; clear access-management gap | Creation can invite; deletion removes access |
| 5 | Environment UI Config | Public read/update; covers Home and other shared settings | Open-ended shared document and drift/overwrite risk |
| 6 | Organization membership | Public read/update/delete | No direct create; invitations and acceptance are separate lifecycle states |
| Investigate | Preview environment | Web-app lifecycle known; live update/rollback characterized | No supported public contract or non-browser authentication established |
| Investigate | Embargoed-assets protection mode | Public feature behavior and undocumented client mutation known | No supported mode-mutation contract established |
| Investigate | Organization access policy | Live read schema and undocumented partial-update behavior known | No public mutation contract; security-critical shared state |
| Investigate | Identity provider/SSO | Undocumented create/update implementation and live absence behavior known | No public contract; no delete call found; security-critical and lockout-prone |
| Defer/read-only | Entitlements, enforcement, licenses, usage, customer accounts | Live supporting requests | Not demonstrated as user-managed desired state |

## Recommended follow-up capture

Use a disposable organization/space and record only redacted request metadata for one intentional change at a time:

1. Verify preview-environment authentication, validation, and error behavior outside the browser; the client lifecycle and live update/rollback are already characterized.
2. Change Home configuration and determine whether the documented environment UI Config fully represents it.
3. Capture the exact default-language save request in a disposable organization. Exercise access-policy writes only if Contentful confirms support and an explicit lockout and access-removal safety plan exists.
4. Exercise identity-provider lifecycle only with explicit authorization and a test IdP; SSO mistakes can lock users out.
5. Move embargoed assets through supported modes only in a disposable space, because activation changes asset URL behavior.
6. For every undocumented candidate, verify behavior using a personal access token outside the web app. A browser-session endpoint being callable does not prove that it is supported by public CMA credentials.

Do not copy browser cookies, bearer tokens, customer IDs, emails, or full response bodies into repository fixtures or notes.

## First-party sources

- [Content Management API overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/) — CMA scope, authentication, and read/write status.
- [Content Management API reference index](https://www.contentful.com/developers/docs/references/content-management-api/) — spaces, organizations, security contacts, locales, tags, webhooks, roles, memberships, API keys, access tokens, app resources, asset keys, and other public CMA resource families.
- [UI Config](https://www.contentful.com/developers/docs/references/content-management-api/ui-config/) — environment and per-user UI configuration, including Home views.
- [Get the UI Config](https://www.contentful.com/developers/docs/references/content-management-api/ui-config/get-the-ui-config/) and [Update the UI Config](https://www.contentful.com/developers/docs/references/content-management-api/ui-config/update-the-ui-config/) — documented environment-scoped read and update contracts.
- [Environments](https://www.contentful.com/developers/docs/references/content-management-api/environments/) — environment lifecycle and environment-aware resources.
- [Locales](https://www.contentful.com/developers/docs/references/content-management-api/locales/) — environment-scoped locale lifecycle.
- [Tags](https://www.contentful.com/developers/docs/references/content-management-api/tags/) — environment-scoped tag management.
- [Webhooks](https://www.contentful.com/developers/docs/references/content-management-api/webhooks/) — space webhook CRUD.
- [Roles](https://www.contentful.com/developers/docs/references/content-management-api/roles/) — space settings permissions and role policies.
- [Space memberships: create](https://www.contentful.com/developers/docs/references/content-management-api/space-memberships/create-a-space-membership/) — direct user invitation/membership creation.
- [User Management API](https://www.contentful.com/developers/docs/references/user-management-api/) — organization membership lifecycle.
- [Space members: get all](https://www.contentful.com/developers/docs/references/user-management-api/space-members/get-all-space-members/) — read-only aggregated space-access view.
- [Organizations](https://www.contentful.com/developers/docs/references/content-management-api/organizations/) and [Update an organization](https://www.contentful.com/developers/docs/references/content-management-api/organizations/put-an-organization-id-an-admin-or-owner-has-access-to/) — organization metadata read and update scope.
- [Security Contacts](https://www.contentful.com/developers/docs/references/content-management-api/security-contacts/) — organization security-contact lifecycle.
- [Authentication](https://www.contentful.com/developers/docs/references/authentication/) — API keys and personal access tokens surfaced in space settings.
- [Environment permissions](https://www.contentful.com/help/environments/environments-permissions/) — organization-level granular environment permissions behavior.
- [Single sign-on](https://www.contentful.com/help/faq/sso/) and [SSO configuration module changelog](https://www.contentful.com/developers/changelog/sso-configuration-module/) — public SSO behavior and web-app configuration surface, but not an identity-provider API contract.
- [Getting started with embargoed assets](https://www.contentful.com/developers/docs/tutorials/general/embargoed-assets-getting-started/) — space-level protection modes and asset-key use.
- [Protection modes](https://www.contentful.com/help/media/embargoed-assets/embargoed-assets-modes/) and [Asset keys](https://www.contentful.com/developers/docs/references/content-management-api/asset-keys/) — durable protection behavior and the separate short-lived signing-key API.
- [Embargoed assets](https://www.contentful.com/help/media/embargoed-assets/) — feature scope and space-level configuration.
