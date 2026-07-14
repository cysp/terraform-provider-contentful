# Adversarial review: CMA-manageable settings catalogue

Date reviewed: 2026-07-14

## Scope and standard of proof

This review treats [`contentful-cma-manageable-settings.md`](contentful-cma-manageable-settings.md) as an untrusted claim set. It checks the catalogue against Contentful's current first-party API references, first-party product documentation, and the provider code and OpenAPI description in this checkout.

Verdicts have deliberately narrow meanings:

- **Confirmed** — a current first-party API reference or checked-in source directly supports the claim.
- **Unsupported** — an authoritative source contradicts the claim, or the claimed API contract is not established by the cited evidence.
- **Ambiguous** — live behavior or product documentation supports part of the claim, but scope, mutation semantics, or public support remains unproved.

“No public API reference located” is bounded to the current [CMA reference](https://www.contentful.com/developers/docs/references/content-management-api/), [User Management API reference](https://www.contentful.com/developers/docs/references/user-management-api/), and first-party Contentful search results reviewed on the date above. It is not proof that an unpublished or separately contracted API does not exist.

No customer identifiers, names, URLs, email addresses, tokens, request identifiers, or response bodies are reproduced here.

## Findings ordered by severity

### High: organization memberships are not directly creatable

**Verdict: Unsupported in the initial catalogue; corrected in the refined catalogue.** The initial text said the User Management API documents list, get, create, update, and delete operations for organization memberships. Contentful instead states that an organization membership **cannot be created directly**; an organization invitation creates a pending membership as a side effect. List, get, update, and delete are documented. See [User Management API — Organization Memberships](https://www.contentful.com/developers/docs/references/user-management-api/#organization-memberships).

This matters to Terraform lifecycle design. A conventional create/read/update/delete membership resource cannot implement `Create` against the membership collection. It would need to model an invitation and its transition to membership, or support only import/update/delete of an existing membership. The catalogue's ranking should not present organization membership as ordinary public CRUD.

### Medium: the three private-looking endpoint families remain unsupported public contracts

**Verdict: Ambiguous as live observations; unsupported as public provider contracts.** The current public CMA navigation does not list endpoint families for:

```text
/spaces/{space_id}/preview_environments
/organizations/{organization_id}/access_policy
/organizations/{organization_id}/identity_provider
```

The live update and rollback of an existing preview configuration is good evidence that the web app currently uses the first endpoint with optimistic versioning. It does not prove supported personal-access-token authentication, create/delete behavior, long-term compatibility, or permission semantics. The refined catalogue correctly keeps provider implementation in an investigation tier and labels product fit as only potential.

For SSO, Contentful publicly documents web-app configuration, activation, deactivation, testing, and SAML fields in [Single sign-on](https://www.contentful.com/help/faq/sso/), and identifies the feature as a web-app module in its [SSO configuration module changelog](https://www.contentful.com/developers/changelog/sso-configuration-module/). Those sources do not document the observed `identity_provider` path as a public API.

The initial catalogue incorrectly described `access_policy` as space-scoped. A targeted Safari reinspection during this review confirmed an organization-scoped `GET`. The live object exposed fields for SSO enforcement, MFA enforcement, explicit token authorization, token-expiration limits, SCIM-only management, and granular environment policies. This establishes a broad organization access-policy read shape, not a narrow granular-permissions resource.

The public organization update example separately exposes an `accessPolicy.sso` value inside organization system metadata, but it does not document the observed `access_policy` endpoint or its mutation contract. See [Update an organization](https://www.contentful.com/developers/docs/references/content-management-api/organizations/put-an-organization-id-an-admin-or-owner-has-access-to/). Keep the live endpoint outside the public-contract tier until Contentful documents or confirms it.

### Medium: UI Config support and environment scope are confirmed, but the shortened path and replacement semantics are not

**Verdict: Confirmed for environment-scoped GET/PUT and for a live non-primary-environment read; ambiguous for the shortened primary-environment path and full-replacement claim.** Contentful publicly documents:

```text
GET /spaces/{space_id}/environments/{environment_id}/ui_config
PUT /spaces/{space_id}/environments/{environment_id}/ui_config
```

The reference says the shared configuration is visible to everyone in the environment and includes Home views, preview mode, publishing mode, and timeline controls. See [UI Config](https://www.contentful.com/developers/docs/references/content-management-api/ui-config/), [Get the UI Config](https://www.contentful.com/developers/docs/references/content-management-api/ui-config/get-the-ui-config/), and [Update the UI Config](https://www.contentful.com/developers/docs/references/content-management-api/ui-config/update-the-ui-config/).

A targeted Safari Web Inspector pass from a non-primary environment observed the explicit read path:

```text
/spaces/{space_id}/environments/{environment_id}/ui_config
```

The Home and Content preview panes both requested that path. This removes the earlier ambiguity about whether UI Config itself is space-wide. It does not turn the separately observed shortened primary-environment route into a documented contract.

The Environments reference documents one compatibility precedent: entries can omit the environment fragment to address the primary environment. That precedent makes a similar explanation for the observed `/spaces/{space_id}/ui_config` route plausible, but it does not document the UI Config route itself or establish that every environment-aware resource supports omission. See [Environments — Access content in an environment](https://www.contentful.com/developers/docs/references/content-management-api/environments/). Label this an observed shortened route with an unconfirmed compatibility interpretation, not “legacy” or “private,” unless Contentful confirms either classification.

The web app sent a complete document and used `X-Contentful-Version`; that proves client behavior for the observed mutation. The public UI Config update page describes the request only as a string-to-any map and does not state that omitted properties are deleted. Therefore “full-document replacement” should be expressed as a safe implementation assumption pending a controlled omission test, not as a confirmed endpoint contract. The overwrite-risk warning remains defensible: Contentful's general CMA guidance recommends fetch-modify-update to avoid losing unseen properties. See [CMA overview — Updating content](https://www.contentful.com/developers/docs/references/content-management-api/overview/).

### Medium: embargoed-assets modes are public product configuration, not a documented mode-management API

**Verdict: Confirmed product behavior; unsupported public mutation contract.** Contentful documents space-level configuration and the durable modes `preparation`, `unpublished assets protected`, and `all assets protected`. It also documents creating short-lived asset keys, which is a separate operation. See [Getting started with embargoed assets](https://www.contentful.com/developers/docs/tutorials/general/embargoed-assets-getting-started/), [Protection modes](https://www.contentful.com/help/media/embargoed-assets/embargoed-assets-modes/), and [CMA Asset keys](https://www.contentful.com/developers/docs/references/content-management-api/asset-keys/).

No public endpoint for changing the durable protection mode was located in the current CMA reference. The catalogue correctly separates mode configuration from asset keys and correctly withholds a Terraform resource recommendation. The phrase “public feature docs” should not be upgraded to Tier A API evidence.

### Low: non-primary settings traffic confirms environment-aware paths without proving pane ownership

**Verdict: Confirmed as live read paths; ambiguous as independent manageable settings.** The non-primary-environment pass observed these symbolic path families:

```text
/spaces/{space_id}/environments/{environment_id}/locales
/spaces/{space_id}/environments/{environment_id}/tags
/spaces/{space_id}/environments/{environment_id}/public/content_types
/spaces/{space_id}/environments/{environment_id}/editor_interfaces
/spaces/{space_id}/environments/{environment_id}/resources
/spaces/{space_id}/environments/{environment_id}/app_installations
/spaces/{space_id}/environments/{environment_id}/extensions
```

Locales and tags were the primary data requests of their respective panes. The remaining paths appeared among shared bootstrap traffic and should not be attributed to the Tags pane merely because that page was used for capture. The `public/content_types` spelling is a live observation, not a claim that the route is in Contentful's public API reference.

The non-primary Content preview pane also requested the space-scoped `/spaces/{space_id}/preview_environments` collection. This is evidence against rewriting that family as environment-scoped based solely on the selected environment in the UI.

The non-primary settings menu omitted General settings, Users, Roles, Embargoed assets, Webhooks, and Usage. This supports the catalogue's separation of environment-local panes from broader space or organization state, but menu visibility remains product-UI evidence rather than a formal API guarantee.

### Low: security-contact CRUD and organization update are public, with narrower contracts than the ranking implies

**Verdict: Confirmed.** Security contacts have public collection `GET`/`POST` and item `PUT`/`DELETE` operations under the organization scope. See [Security Contacts](https://www.contentful.com/developers/docs/references/content-management-api/security-contacts/), [create](https://www.contentful.com/developers/docs/references/content-management-api/security-contacts/post-an-organization-security-contacts-an-admin-or-owner-has-access-to/), [update](https://www.contentful.com/developers/docs/references/content-management-api/security-contacts/update-an-organization-security-contact-an-admin-or-owner-has-access-to/), and [delete](https://www.contentful.com/developers/docs/references/content-management-api/security-contacts/delete-an-organization-security-contact-an-admin-or-owner-has-access-to/). The catalogue's verb and scope claims are accurate.

The organization endpoint publicly supports `GET` and `PUT`, and the update example returns mutable organization name. See [Organizations](https://www.contentful.com/developers/docs/references/content-management-api/organizations/) and [Update an organization](https://www.contentful.com/developers/docs/references/content-management-api/organizations/put-an-organization-id-an-admin-or-owner-has-access-to/). This supports an organization-settings singleton or metadata update. It does not establish public organization creation/deletion, so “organization resource” is too broad unless its lifecycle is intentionally constrained.

### Low: locale, direct space membership, and aggregate space-member claims are accurate

**Verdict: Confirmed.** Locales are environment-scoped and have public list/create/get/update/delete operations; the provider does not register a locale resource. See [Locales](https://www.contentful.com/developers/docs/references/content-management-api/locales/) and [Update a locale](https://www.contentful.com/developers/docs/references/content-management-api/locales/update-a-locale/).

Direct space memberships have public collection `GET`/`POST` and item `GET`/`PUT`/`DELETE` operations. The create operation can invite a user to the space. See [Space memberships](https://www.contentful.com/developers/docs/references/content-management-api/space-memberships/) and [Create a space membership](https://www.contentful.com/developers/docs/references/content-management-api/space-memberships/create-a-space-membership/). The separate User Management API exposes space-member aggregation as a read model. The catalogue's recommendation to manage memberships rather than aggregate members is sound.

## Public-documentation verdict matrix

| Resource or setting | Verdict | Confirmed public surface | Correction or boundary |
| --- | --- | --- | --- |
| Preview configuration | **Ambiguous** | Web-app feature documentation only | Observed endpoint and update are not a supported public contract |
| Organization access policy | **Ambiguous** | Live organization-scoped read shape; public organization response exposes a separate `accessPolicy.sso` field | No public endpoint or mutation contract located |
| Organization identity provider | **Ambiguous** | SSO behavior and web-app setup are public | No public CMA endpoint contract located |
| Security contacts | **Confirmed** | Organization-scoped collection and item CRUD | Personally identifiable state and authorization still need design work |
| Organization memberships | **Unsupported as CRUD** | List/get/update/delete; invitation creates membership indirectly | No direct membership create operation |
| Organization metadata | **Confirmed, narrow** | Organization GET/PUT, including name | Not a full organization lifecycle |
| UI Config | **Confirmed, environment-scoped** | Shared and user-specific GET/PUT | Shortened path and omission behavior remain ambiguous |
| Embargoed-assets mode | **Confirmed feature; unsupported API mutation** | Modes and asset-key creation are public | Asset key creation does not manage durable mode |
| Locale | **Confirmed** | Environment-scoped CRUD | Default and fallback transition tests remain prudent |
| Space membership | **Confirmed** | Space-scoped CRUD/invite | Access-removal and invitation side effects are material |

## Repository coverage audit

**Verdict: Confirmed.** [`ContentfulProvider.Resources`](../../internal/provider/contentful_provider.go) registers exactly 21 managed resource types. The catalogue's names and count match the code. In particular, it already includes `contentful_space_enablements` and `contentful_team_space_membership`, but not locale, direct space membership, organization membership, security contact, organization metadata, or UI Config resources.

The checked-in [`openapi.yml`](../../internal/contentful-management-go/openapi/openapi.yml) includes space enablements and the provider's other generated-client operations, but none of the investigated locale, UI Config, security-contact, organization, direct membership, preview-configuration, access-policy, identity-provider, or embargoed-mode paths. That is a local implementation constraint, not evidence that the public APIs are absent: several are confirmed in Contentful's current public references but have not yet been added to this embedded specification.

## Adversarial assessment of the candidate ranking

The existing ranking mixes product value, API support, implementation effort, destructive risk, and state sensitivity without defining weights. Its exact order is therefore **ambiguous**, not evidence-backed. A more defensible readiness ordering is:

1. **Locale** — public environment-scoped CRUD and a conventional stable identity, with known default/fallback constraints.
2. **Security contact** — public CRUD and a clean web-setting mapping, but email data will be stored in Terraform state.
3. **Space name/settings singleton** — public read/update with a very small schema; avoid pretending Terraform owns space deletion.
4. **Direct space membership** — public CRUD, but creation can send an invitation and deletion removes access.
5. **Environment UI Config** — public GET/PUT, but it is an open-ended shared document with substantial ownership and drift risk.
6. **Organization membership** — public update/delete but no direct create; model invitations separately before calling it a managed lifecycle.

Keep preview configuration, access policy, identity-provider configuration, and embargoed-assets protection mode out of an implementation queue until a supported authentication and mutation contract is established. Their live mutability demonstrates technical possibility, not a stable third-party API commitment.

The ordering above is still a design recommendation rather than a fact. Before implementation, score candidates explicitly against public support, complete lifecycle, blast radius, sensitive state, importability, concurrency, and likely user demand.

## Catalogue corrections applied

The refined catalogue incorporates these review outcomes:

1. Organization memberships are described as list/get/update/delete, with invitation as the indirect creation workflow.
2. The observed space-scoped UI Config route is described as an unconfirmed compatibility route rather than private or legacy.
3. UI Config's complete web-app payload is recorded as client behavior, not proof that the server deletes omitted fields.
4. `preview_environments`, `access_policy`, and `identity_provider` remain outside the supported-public-API tier.
5. Organization metadata update is separated from full organization lifecycle.
6. Candidate readiness is explicitly an engineering recommendation with stated criteria.
7. `access_policy` is corrected to organization scope and characterized as broad security-sensitive shared state.
8. A non-primary-environment capture confirms explicit environment paths for UI Config, locales, tags, and shared bootstrap resources while preserving the space scope of preview-environment definitions.
