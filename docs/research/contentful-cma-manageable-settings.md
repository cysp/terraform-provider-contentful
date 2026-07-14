# Contentful CMA-manageable settings observed in the web app

Date observed: 2026-07-14

## Purpose

This note catalogues settings and resource families exposed by the Contentful web app that are not all represented by the provider's currently registered Terraform resources. It combines:

- read-only observation of the signed-in Contentful web app and Safari Web Inspector network traffic;
- the provider registrations in [`internal/provider/contentful_provider.go`](../../internal/provider/contentful_provider.go); and
- first-party Contentful API and Help Center documentation available on the observation date.

This is the final evidence catalogue, not an assertion that every private web-app endpoint is a supported public API. Most observations were read-only. Two low-impact settings were changed and immediately restored to characterize their write contracts. No access-control, membership, SSO, deletion, billing, secret, embargoed-asset, or publishing setting was changed.

All organization, space, environment, user, account, and content identifiers are symbolic placeholders. Examples preserve exact API field names and protocol enum values only where their spelling is part of the finding. No response bodies containing customer or personal data are reproduced.

## Evidence tiers

| Tier | Meaning |
| --- | --- |
| A — publicly documented | Contentful documents a public API contract for the resource or setting. This is the strongest basis for provider support. |
| B — live-observed only | The web app exposed a mutable setting and/or called an endpoint, but no matching public CMA reference was found. Treat public support and stability as unknown until Contentful documents or confirms them. |
| C — read-only/supporting | The endpoint appears to supply entitlement, enforcement, membership aggregation, billing, license, or usage context. Observation alone does not establish a manageable resource. |

“Public docs not found” means no matching endpoint was found in Contentful's public CMA and User Management API references during this survey; it does not prove that no documentation exists elsewhere.

## Current provider baseline

The provider currently registers these 21 Terraform resource types:

`app_definition`, `app_signing_secret`, `app_installation`, `content_type`, `delivery_api_key`, `editor_interface`, `environment_alias`, `environment`, `entry`, `extension`, `personal_access_token`, `resource_provider`, `resource_type`, `role`, `space_enablements`, `tag`, `taxonomy_concept`, `taxonomy_concept_scheme`, `team`, `team_space_membership`, and `webhook`.

The primary-environment settings menu exposed:

`Locales`, `Tags`, `Home`, `Content preview`, `General settings`, `Users`, `Roles`, `Environments`, `API keys`, `CMA tokens`, `Embargoed assets`, `Webhooks`, and `Usage`.

The non-primary-environment menu exposed only `Locales`, `Tags`, `Home`, `Content preview`, `Environments`, `API keys`, and `CMA tokens`. `General settings`, `Users`, `Roles`, `Embargoed assets`, `Webhooks`, and `Usage` were absent. This is UI evidence that the omitted panes are not environment-local; it is not by itself an API contract.

The organization settings UI also exposed mutable organization name, granular environment permissions, default language, security-notification email addresses, and SSO configuration.

## Catalogue by settings surface

| Web-app surface | API/resource interpretation | Evidence | Provider coverage | Assessment |
| --- | --- | --- | --- | --- |
| Locales | Environment-scoped locale resources | A | Missing | Strong candidate. Public CMA supports listing, creation, update, and deletion of locales. |
| Tags | Environment-scoped tags | A | Covered by `contentful_tag` | No additional resource implied. |
| Home | App-provided/built-in home widgets, app installations, and UI config | A, plus live mapping | Partly covered by `app_installation` and `extension`; UI config missing | The documented UI Config includes `homeViews` entries. A dedicated UI-config resource would cover the environment-visible arrangement; app installation/extension lifecycle is already represented. |
| Content preview | Preview-environment configuration | B | Missing | Potential product fit. Update and rollback traffic is characterized, but creation, deletion, ordering, authentication outside the web app, and API support remain unverified. |
| General settings | Space object, especially mutable space name | A | Missing | Strong candidate for a space resource or narrowly scoped space-settings resource. Public CMA documents space creation, name update, deletion, and unarchive. |
| Users | Space memberships are mutable; space members are an aggregate view | A and C | Direct user/space membership missing; team-space membership covered | Model `space_memberships`, not the read-only `space_members` aggregation. |
| Roles | Space roles and policies | A | Covered by `contentful_role` | No additional top-level resource implied. |
| Environments | Environments and aliases | A | Covered | No additional top-level resource implied by the menu item. |
| API keys | Delivery/preview API-key pairs | A | Delivery key covered; preview key is a data source | Existing coverage is substantially aligned. |
| CMA tokens | Personal access tokens | A | Covered by `contentful_personal_access_token` | Existing coverage is aligned. |
| Embargoed assets | Space-level protection mode; asset keys are short-lived signing credentials | Public feature docs; toggle contract not identified | Missing | Do not equate the documented asset-key endpoint with management of the space's protection mode. The UI makes the mode mutable, but this survey did not identify a supported public mutation contract for that setting. |
| Webhooks | Space-scoped webhook definitions | A | Covered by `contentful_webhook` | Existing coverage is aligned. |
| Usage | Usage, entitlements, enforcement, account, and license views | C | Missing | Treat as informational unless a documented mutation contract and a stable Terraform lifecycle emerge. |
| Organization name | Organization metadata | A for organization retrieval/update; live UI | Missing | Possible organization-settings singleton. Public evidence supports metadata read/update, not full organization lifecycle. |
| Organization access policy | Organization-wide access controls, including granular environment permissions | B; public Help Center describes some behavior | Missing | The live read schema is characterized, but mutation and public-support contracts are not. This is security-sensitive shared state, not a narrow feature toggle. |
| Default language | Organization-level default for the web app and email communications | B | Missing | Users can override it in their profiles. Candidate only after its endpoint and update contract are captured and supported; it is not a space environment locale. |
| Security notification emails | Organization security contacts | A | Missing | Strong candidate. Public CMA documents get, create, update, and delete operations for organization security contacts. |
| SSO | Organization identity-provider singleton/configuration | B | Missing | Potentially high-value but high-risk. The live GET returned `404` when unset; no public endpoint contract was found. |

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

No mutation was performed during this non-primary-environment pass. In particular, it does not establish whether every environment-explicit read path accepts writes, whether the observed `public/content_types` route is a supported third-party contract, or whether a personal access token can call private-looking routes used by the web app.

## Reversible mutation evidence

### Preview mode through the observed shortened UI Config route

A reversible primary-environment preview-mode change succeeded through the shortened UI Config route, and restoration was verified in the UI. The public contract and a non-primary-environment capture both use the environment-explicit UI Config path. UI Config is therefore environment-scoped, while the mutation evidence characterizes only the shortened primary-environment route.

The web app used:

```http
PUT /spaces/{space_id}/ui_config
Content-Type: application/vnd.contentful.management.v1+json
X-Contentful-Version: {current_version}
```

The web app sent the complete shared UI Config document, not a one-field patch. That establishes observed client behavior, but not server-side replacement semantics: this survey did not test whether an omitted property is preserved or removed. The relevant value changed between:

```json
{"livePreview":{"previewMode":"legacyPreview"}}
```

and:

```json
{"livePreview":{"previewMode":"livePreview"}}
```

The same document also contained publishing mode, Home views, and entry-list views. This reinforces the ownership concern for Terraform: a conservative client should fetch, merge, and update the document so it preserves unknown and unmanaged fields.

The mutation endpoint omitted the environment component, while both the public UI Config documentation and the non-primary live read are environment-scoped. Contentful documents shortened paths for some environment-aware resources as addressing the primary environment, so this may be a compatibility route. Do not classify the shortened route as public, private, or legacy until Contentful confirms its status.

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

Each update sent the current resource version and returned its successor; the restoration used the version returned by the preceding change. This establishes optimistic versioning for updates without retaining concrete observed values, but does not establish a supported public contract. Browser-session authorization, create/delete semantics, ordering, error behavior, and personal-access-token compatibility remain unverified.

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

Update and rollback request/response shapes are recorded in the reversible-mutation section. Before provider implementation, capture creation, reordering, and deletion traffic in a disposable test configuration; verify supported non-browser authentication; and confirm the endpoint is supported for third-party use.

### Embargoed-assets protection mode

Contentful publicly documents that embargoed assets are enabled at space level and configured using durable protection modes. It separately documents creation of short-lived asset keys through Contentful APIs.

Those are distinct concerns:

- the protection mode is durable space configuration shown in the UI, but no public mode-mutation endpoint was located;
- an asset key is an expiring secret used to sign protected URLs and has a public creation API.

Protection-mode management therefore remains in the investigation tier. An asset-key resource is technically better supported, but would still need write-only or ephemeral handling because the returned secret is short-lived.

### Organization access policy

The web app called:

```text
/organizations/{organization_id}/access_policy
```

The observed organization-scoped `GET` returned an access-policy object with fields for SSO enforcement, MFA enforcement, explicit token authorization, token-expiration limits, SCIM-only management, and granular environment policies. The object is broad, security-sensitive shared state rather than a narrow granular-permissions toggle.

No matching public CMA endpoint reference was found. The live response establishes read shape, not supported mutation methods, field coupling, authentication outside the web app, or safe rollback behavior. Keep this object outside the implementation-ready tier.

### Organization identity provider

The SSO UI mapped to:

```text
/organizations/{organization_id}/identity_provider
```

An observed `GET` returned `404` when no identity provider was configured. That is useful absence behavior, but it does not establish supported create/update/delete methods or payload stability. No matching public CMA endpoint reference was found.

### Organization-level preferences

The UI exposed mutable organization name, granular environment permissions, and default language. Organization retrieval/update is publicly documented, granular environment permissions are described in the Help Center, and the default language is described in the UI as applying to the web app and email communications unless a user overrides it. This survey did not capture supported mutation contracts for the access-policy or language settings.

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
| Investigate | Preview environment | Live update/rollback characterized | No supported public contract or non-browser authentication established |
| Investigate | Embargoed-assets protection mode | Public feature behavior | No supported mode-mutation contract established |
| Investigate | Organization access policy | Live read schema and mutable UI | No public mutation contract; security-critical shared state |
| Investigate | Identity provider/SSO | Live mutable UI and absence behavior | No public contract; security-critical and lockout-prone |
| Defer/read-only | Entitlements, enforcement, licenses, usage, customer accounts | Live supporting requests | Not demonstrated as user-managed desired state |

## Recommended follow-up capture

Use a disposable organization/space and record only redacted request metadata for one intentional change at a time:

1. Create, reorder, and delete a preview environment; update and rollback behavior is already characterized.
2. Change Home configuration and determine whether the documented environment UI Config fully represents it.
3. In a disposable organization, test one organization access-policy field and the default language independently, recording the endpoint, method, concurrency header, and minimal symbolic diff. Access-policy changes require an explicit lockout and access-removal safety plan.
4. Exercise identity-provider create/update/delete only with explicit authorization and a test IdP; SSO mistakes can lock users out.
5. Move embargoed assets through supported modes only in a disposable space, because activation changes asset URL behavior.
6. For every candidate, verify behavior using a personal access token outside the web app. A browser-session endpoint being callable does not prove that it is supported by public CMA credentials.

Do not copy browser cookies, bearer tokens, customer IDs, emails, or full response bodies into repository fixtures or notes.

## First-party sources

- [Content Management API overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/) — CMA scope, authentication, and read/write status.
- [Content Management API reference index](https://www.contentful.com/developers/docs/references/content-management-api/) — spaces, organizations, security contacts, locales, tags, webhooks, roles, memberships, API keys, access tokens, app resources, asset keys, and other public CMA resource families.
- [UI Config](https://www.contentful.com/developers/docs/references/content-management-api/ui-config/) — environment and per-user UI configuration, including Home views.
- [Environments](https://www.contentful.com/developers/docs/references/content-management-api/environments/) — environment lifecycle and environment-aware resources.
- [Tags](https://www.contentful.com/developers/docs/references/content-management-api/tags/) — environment-scoped tag management.
- [Webhooks](https://www.contentful.com/developers/docs/references/content-management-api/webhooks/) — space webhook CRUD.
- [Roles](https://www.contentful.com/developers/docs/references/content-management-api/roles/) — space settings permissions and role policies.
- [Space memberships: create](https://www.contentful.com/developers/docs/references/content-management-api/space-memberships/create-a-space-membership/) — direct user invitation/membership creation.
- [User Management API](https://www.contentful.com/developers/docs/references/user-management-api/) — organization membership lifecycle.
- [Space members: get all](https://www.contentful.com/developers/docs/references/user-management-api/space-members/get-all-space-members/) — read-only aggregated space-access view.
- [Authentication](https://www.contentful.com/developers/docs/references/authentication/) — API keys and personal access tokens surfaced in space settings.
- [Environment permissions](https://www.contentful.com/help/environments/environments-permissions/) — organization-level granular environment permissions behavior.
- [Getting started with embargoed assets](https://www.contentful.com/developers/docs/tutorials/general/embargoed-assets-getting-started/) — space-level protection modes and asset-key use.
- [Embargoed assets](https://www.contentful.com/help/media/embargoed-assets/) — feature scope and space-level configuration.
