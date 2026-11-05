# Organization settings and access configuration

Organization metadata, feature enablements, security contacts, memberships, and
access configuration are addressed separately.

## Scope and evidence

Published sources reviewed: 2026-11-07. The direct observations below come from
GET requests made with a personal management token on 2026-07-14. No write
requests to these organization endpoints were tested.

## Metadata and preferences

The [CMA organization reference][organizations] documents GET and PUT at
`/organizations/{organization_id}`. The [pinned SDK entity][organization-entity]
declares `name` and `defaultUserLanguage`. Its documentation describes
`defaultUserLanguage` as the language for the interface and email, with a per-user
override in profile settings. This preference is separate from environment
[Locales](locales.md). The type describes the returned organization; accepted
update fields and omission behavior were not tested.

## Feature enablements

[Organization Enablements][enablements] uses
`GET` and `PUT /organizations/{organization_id}/enablements`, separately from
[Space Enablements](space-enablements.md). GET creates a default document with
all feature values false when none exists: reading enablements can create remote
state. PUT requires the document's last version in `X-Contentful-Version`, or `1`
if no document exists yet.

The direct read returned a versioned singleton containing built-in AI actions,
custom AI actions, and vectorization objects. Whether that read created the document
was not determined. Write validation and omission behavior remain untested.

## Security contacts

The [CMA security-contact reference][security-contacts] documents collection GET
and POST at `/organizations/{organization_id}/security_contacts`, with PUT and
DELETE on individual contacts. The direct collection read returned `200` with a
paginated representation.

## Memberships and invitations

The [UMA organization-membership reference][memberships] documents list, get,
update, and delete operations. Creating an [organization invitation][invitations]
also creates a membership with status `pending`; memberships cannot be created
directly. A pending member can be added to teams and spaces, but must accept the
invitation before accessing the organization. Deleting the associated organization
membership removes its invitations; there is no direct invitation DELETE operation.
The invitation reference labels its endpoints alpha. The [UMA overview][uma]
describes the API's availability requirements.

For the relationship between individual and team access to a space, see
[direct space memberships and aggregate SpaceMember views](space-access.md#direct-memberships-and-aggregate-members).

## Access policy, identity provider, and certificates

The direct reads addressed three resources independently:

| GET path | Observed result |
| --- | --- |
| `/organizations/{organization_id}/access_policy` | `200`; versioned singleton with SSO and MFA enforcement, explicit-token authorization, SCIM-only management, token-expiration limit, and granular-environment-policy fields |
| `/organizations/{organization_id}/identity_provider` | `404 NotFound` |
| `/organizations/{organization_id}/certificates` | `200`; paginated collection |

The study did not determine why the identity-provider request returned 404 or
identify a published contract for these routes. Their create, update, delete,
and concurrency behavior remains unverified.

[organizations]: https://www.contentful.com/developers/docs/references/content-management-api/organizations/
[organization-entity]: https://github.com/contentful/contentful-management.js/blob/83d6a3f5fcbc9773c21b6779fe66c7fe59349f66/lib/entities/organization.ts
[enablements]: https://www.contentful.com/developers/docs/references/content-management-api/organization-enablements/
[security-contacts]: https://www.contentful.com/developers/docs/references/content-management-api/security-contacts/
[memberships]: https://www.contentful.com/developers/docs/references/user-management-api/organization-memberships/
[invitations]: https://www.contentful.com/developers/docs/references/user-management-api/invitations/
[uma]: https://www.contentful.com/developers/docs/references/user-management-api/overview/
