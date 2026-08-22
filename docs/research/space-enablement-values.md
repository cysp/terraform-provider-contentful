# Space Enablements request values

Contentful's [Space Enablements CMA reference](https://www.contentful.com/developers/docs/references/content-management-api/space-enablements/)
currently says that `spaceTemplates` and `crossSpaceLinks` must be enabled or
disabled together. The update endpoint describes the request as a map rather
than publishing a required-member schema. The provider's
[generated request schema](../../internal/contentful-management-go/openapi/schemas/space-enablement/data.yml)
likewise defines `crossSpaceLinks`, `spaceTemplates`, `studioExperiences`, and
`suggestConcepts` as independent optional members.

On 2026-08-22, isolated live CMA `PUT` requests established the current server
validation behavior. Each request used the document's current
`X-Contentful-Version`:

| Request body | Result | Validation detail |
| --- | --- | --- |
| `{}` | HTTP 422 | `spaceTemplates` is required. |
| `{"crossSpaceLinks":{"enabled":true}}` | HTTP 422 | `spaceTemplates` is required. |
| `{"spaceTemplates":{"enabled":false}}` | HTTP 422 | `crossSpaceLinks` is required. |
| `{"crossSpaceLinks":{"enabled":true},"spaceTemplates":{"enabled":false}}` | HTTP 422 | Both fields must be enabled or disabled. |

The failed requests did not change the enablement document. A final GET verified
that its original four-field representation was unchanged.

Provider impact: current member-presence and equality checks are CMA server
validation, not structural Terraform configuration invariants. All four
Terraform attributes remain independently Optional+Computed. Request conversion
sends every known Plan value, including values preserved from prior state and
explicit `false`; omits response-owned null or unknown Plan values; and rejects
configuration-owned unknown Plan values before I/O. It does not infer a sibling
or reject known unequal values, so future CMA support for partial or independent
updates does not require a provider schema change. Current CMA validation errors
are surfaced as remote API failures.
