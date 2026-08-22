# Taxonomy DELETE version behavior

## Published CMA contract

Contentful's CMA references for both taxonomy DELETE endpoints document
`X-Contentful-Version` as a required string header and include it in their cURL
examples:

- [Delete a concept](https://www.contentful.com/developers/docs/references/content-management-api/taxonomy/delete-a-concept/)
  describes it as the version of the concept. The page currently says "to
  update," despite documenting a DELETE operation.
- [Delete a concept scheme](https://www.contentful.com/developers/docs/references/content-management-api/taxonomy/delete-a-concept-scheme/)
  describes it as the version of the concept scheme to delete.

Both endpoint pages document HTTP 204 No Content as the successful response.
Neither page documents endpoint-specific responses for an omitted, zero,
malformed, or stale version header.

Contentful's general [CMA overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/#updating-and-version-locking)
documents version locking for updates. Its general
[error reference](https://www.contentful.com/developers/docs/references/errors/)
associates HTTP 409 `VersionMismatch` with an omitted or outdated version when
updating assets, entries, or content types. Those general statements do not
establish the runtime contract of the taxonomy DELETE endpoints.

## Direct CMA observations

Raw CMA requests against disposable resources established the following current
behavior for both concepts and concept schemes. Each resource was created by a
caller-defined ID and initially returned `sys.version: 1`. The stale case was
advanced to version `2` by a successful PATCH before DELETE.

| DELETE request | Concept | Concept scheme | Follow-up GET |
| --- | --- | --- | --- |
| Correct version | HTTP 204, empty body | HTTP 204, empty body | HTTP 404; object absent |
| Header omitted | HTTP 422 `ValidationFailed` | HTTP 422 `ValidationFailed` | HTTP 200 at version `1`; object remains |
| Version `0` | HTTP 422 `ValidationFailed` | HTTP 422 `ValidationFailed` | HTTP 200 at version `1`; object remains |
| Stale version `1` against version `2` | HTTP 409 `VersionMismatch` | HTTP 409 `VersionMismatch` | HTTP 200 at version `2`; object remains |

Header omission was repeated three times independently for each resource type;
all six requests returned the same status and error. The exact observed error
bodies were identical between concepts and concept schemes.

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

## Provider consequence

The documented required header matches current CMA behavior. Omitting the
header is not a safe deletion mode, and version `0` is not a valid lock token.
The provider therefore sends a positive private-state version directly when it
is available. Terraform omits prior private data from Delete for a tainted
replacement, both after an errored Create and after a resource is manually
marked tainted. Only in that genuinely absent case, Delete performs one GET for
the requested resource and uses its positive `sys.version`; a 404 means
deletion is already complete. Malformed, zero, or negative private data remains
an error and does not trigger the GET.

This fallback reads at apply time because Terraform did not supply private data
to Delete. It retains optimistic locking over the destructive operation: if the
resource changes after GET, the DELETE carries the now-stale version and CMA
rejects it with HTTP 409 `VersionMismatch`. Valid private data does not cause a
preliminary GET.

With no prior version available, the fallback cannot detect a remote change
that happened before GET. It deliberately authorizes deletion of the version
current at GET time; the CMA header protects only the interval from that GET to
DELETE.
