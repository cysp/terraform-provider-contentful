# Space Enablements: request values

Space Enablements configure capabilities on a space. The published contract couples
`spaceTemplates` and `crossSpaceLinks`; direct requests also found that both members
were required and their `enabled` values had to agree.

## Addressing and operations

Published endpoint references reviewed: 2026-09-09.

| Operation | Path | Published behavior |
| --- | --- | --- |
| GET for a space | `/spaces/{space_id}/enablements` | Returns the document; if absent, creates a default document with features false. |
| PUT for a space | `/spaces/{space_id}/enablements` | Updates using `X-Contentful-Version`; the reference specifies the last version, or `1` when no document exists. |
| GET for an organization | `/organizations/{organization_id}/space_enablements` | Lists space enablement documents. |

Sources: [space
GET](https://www.contentful.com/developers/docs/references/content-management-api/space-enablements/get-the-enablements-for-a-space/),
[space
PUT](https://www.contentful.com/developers/docs/references/content-management-api/space-enablements/update-the-enablements-for-a-space/),
and [organization
collection](https://www.contentful.com/developers/docs/references/content-management-api/space-enablements/get-all-enablements-for-an-organization/).
A space GET is therefore documented to initialize an absent document; it must not be
assumed to be a side-effect-free existence check.

## Published evidence

Contentful's [Space Enablements CMA
reference](https://www.contentful.com/developers/docs/references/content-management-api/space-enablements/)
says that `spaceTemplates` and `crossSpaceLinks` must be enabled or disabled together.
The update endpoint describes the request as a map rather than publishing a
required-member schema. An open-ended request map does not establish that individual
members can be omitted in every service configuration.

## Direct CMA observations

Observed: 2026-08-22.

Isolated live CMA `PUT` requests established the server validation behavior for the
tested requests. Each request used the document's current
`X-Contentful-Version`:

| Request body | Result | Validation detail |
| --- | --- | --- |
| `{}` | HTTP 422 | `spaceTemplates` is required. |
| `{"crossSpaceLinks":{"enabled":true}}` | HTTP 422 | `spaceTemplates` is required. |
| `{"spaceTemplates":{"enabled":false}}` | HTTP 422 | `crossSpaceLinks` is required. |
| `{"crossSpaceLinks":{"enabled":true},"spaceTemplates":{"enabled":false}}` | HTTP 422 | Both fields must be enabled or disabled. |

The failed requests did not change the enablement document. A final GET verified that
its original four-field representation was unchanged.

## Interpretation and limits

The four retained failures establish member-presence and equality checks for those exact
requests. They do not establish independent enablement support, successful partial
updates, or defaults for `studioExperiences` and `suggestConcepts`. No client schema or
permissive SDK type overrides server validation. Preserve explicit `false` when forming
requests; omission and `{"enabled": false}` are different representations.
