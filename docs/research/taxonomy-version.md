# Taxonomy version behavior

## Published CMA contract

Contentful's CMA references for taxonomy PATCH and DELETE endpoints document
`X-Contentful-Version` as a required string header and include it in their cURL
examples:

- [Update a concept](https://www.contentful.com/developers/docs/references/content-management-api/taxonomy/update-a-concept/)
  requires the current version of the concept.
- [Update a concept scheme](https://www.contentful.com/developers/docs/references/content-management-api/taxonomy/update-a-concept-scheme/)
  requires the current version of the concept scheme.
- [Delete a concept](https://www.contentful.com/developers/docs/references/content-management-api/taxonomy/delete-a-concept/)
  describes it as the version of the concept. The page currently says "to
  update," despite documenting a DELETE operation.
- [Delete a concept scheme](https://www.contentful.com/developers/docs/references/content-management-api/taxonomy/delete-a-concept-scheme/)
  describes it as the version of the concept scheme to delete.

The DELETE endpoint pages document HTTP 204 No Content as the successful
response. The endpoint pages do not document specific responses for an omitted,
zero, negative, malformed, or stale version header.

Contentful's general [CMA overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/#updating-and-version-locking)
documents version locking for updates. Its general
[error reference](https://www.contentful.com/developers/docs/references/errors/)
associates HTTP 409 `VersionMismatch` with an omitted or outdated version when
updating assets, entries, or content types. Those general statements do not
establish the runtime contract of the taxonomy endpoints.

## Direct CMA observations

Raw CMA requests against disposable resources established the following current
behavior for both concepts and concept schemes. Each resource was created by a
caller-defined ID and initially returned `sys.version: 1`. A PATCH with version
`1` succeeded and advanced each resource to version `2`.

| Operation and version | Concept | Concept scheme | Observed effect |
| --- | --- | --- | --- |
| PATCH, header omitted | HTTP 422 `ValidationFailed` | HTTP 422 `ValidationFailed` | Version remained `1` |
| PATCH, version `0` | HTTP 422 `ValidationFailed` | HTTP 422 `ValidationFailed` | Version remained `1` |
| PATCH, version `-1` | HTTP 422 `ValidationFailed` | HTTP 422 `ValidationFailed` | Version remained `1` |
| PATCH, exact version `1` | HTTP 200, version `2` | HTTP 200, version `2` | Requested change applied |
| PATCH, stale version `1` against version `2` | HTTP 409 `VersionMismatch` | HTTP 409 `VersionMismatch` | Version remained `2` |
| DELETE, header omitted | HTTP 422 `ValidationFailed` | HTTP 422 `ValidationFailed` | Object remained at version `2` |
| DELETE, version `0` | HTTP 422 `ValidationFailed` | HTTP 422 `ValidationFailed` | Object remained at version `2` |
| DELETE, version `-1` | HTTP 422 `ValidationFailed` | HTTP 422 `ValidationFailed` | Object remained at version `2` |
| DELETE, stale version `1` against version `2` | HTTP 409 `VersionMismatch` | HTTP 409 `VersionMismatch` | Object remained at version `2` |
| DELETE, exact version `2` | HTTP 204, empty body | HTTP 204, empty body | Follow-up GET returned HTTP 404 |

DELETE header omission was repeated three times independently for each resource
type; all six requests returned the same status and error. The exact observed
DELETE error bodies were identical between concepts and concept schemes.

Omitted header:

```json
{
  "sys": {"type": "Error", "id": "ValidationFailed"},
  "message": "Validation error",
  "details": {
    "flatten": {
      "formErrors": [],
      "fieldErrors": {
        "x-contentful-version": ["Invalid input: expected number, received NaN"]
      }
    },
    "errors": [
      {
        "name": "invalid_type",
        "path": ["x-contentful-version"],
        "details": "Invalid input: expected number, received NaN"
      }
    ]
  }
}
```

Version `0`:

```json
{
  "sys": {"type": "Error", "id": "ValidationFailed"},
  "message": "Validation error",
  "details": {
    "flatten": {
      "formErrors": [],
      "fieldErrors": {
        "x-contentful-version": ["Too small: expected number to be >0"]
      }
    },
    "errors": [
      {
        "name": "too_small",
        "path": ["x-contentful-version"],
        "details": "Too small: expected number to be >0"
      }
    ]
  }
}
```

Stale version:

```json
{
  "sys": {"type": "Error", "id": "VersionMismatch"},
  "message": "Version mismatch",
  "details": "Version mismatch, expected 2, got 1."
}
```

Every disposable resource was then deleted with its correct version. A final
GET of each identifier returned HTTP 404.

## Provider contract

The provider policy informed by these observations is defined in
[Terraform value semantics](../design/terraform-value-semantics.md#taxonomy-optimistic-version-locking).
