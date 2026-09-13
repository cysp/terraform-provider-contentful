# Preview environments: observed CMA behavior

A preview environment manages Content Type preview URLs at space level. In the recorded
requests, updates merged configurations by Content Type identity: disabling a
configuration removed it from the active preview settings, while omission left it
unchanged. Disabled configurations remained readable and could be re-enabled.

## Scope and evidence

Contentful documents content previews as a product feature. No public endpoint reference
for `PreviewEnvironment` or `/preview_environments` was identified in the source review.
The HTTP behavior below comes from direct observations rather than a published endpoint
contract.

A preview environment is space-level configuration for a content preview platform. It
maps content types to preview URL templates. It is not a Contentful sandbox environment
and is not the Content Preview API that serves draft entries.

Primary sources:

- [Set up content preview](https://www.contentful.com/developers/docs/tutorials/preview/content-preview/)
- [Content Management API reference](https://www.contentful.com/developers/docs/references/content-management-api/)
- [Content Preview API overview](https://www.contentful.com/developers/docs/references/content-preview-api/overview/)
- Direct authenticated observations of the production Content Management API and Contentful Web App on 2026-07-14, 2026-07-15, and 2026-07-17

Product documentation establishes the purpose of preview platforms; it does not
establish the HTTP mutation, normalization, or concurrency behavior recorded below.

## Observed endpoints

All paths are relative to the configured Content Management API base URL.

| Operation | Method and path | Success response |
| --- | --- | --- |
| List | `GET /spaces/{space_id}/preview_environments` | `200`; offset collection |
| Read | `GET /spaces/{space_id}/preview_environments/{preview_environment_id}` | `200` |
| Create with generated ID | `POST /spaces/{space_id}/preview_environments` | `201`; version `0` |
| Create with selected ID | `PUT /spaces/{space_id}/preview_environments/{preview_environment_id}` | `200`; version `0` |
| Update | `PUT /spaces/{space_id}/preview_environments/{preview_environment_id}` | `200` |
| Delete | `DELETE /spaces/{space_id}/preview_environments/{preview_environment_id}` | `204` |

List requests honor `skip` and `limit`. Cursor pagination, filtering, sorting, and
behavior above the documented product limit were not verified.

## Representation and request normalization

The resource contains `name`, `description`, a `configurations` array whose order is
observable, and `sys` metadata. Each configuration represents a content type, URL
template, and enabled state. The observed configuration identity is the pair
`entityType` and `entityId`; the URL and enabled state are mutable values for that
identity.

The service accepts either `contentType` or `entityType: "ContentType"` with `entityId`
when creating a configuration. The Contentful Web App uses this request representation:

```json
{
  "url": "https://example.invalid/preview/{entry.sys.id}",
  "entityId": "content-type-a",
  "entityType": "ContentType",
  "enabled": true,
  "example": false
}
```

The Web App sends `example: false`, but direct API probes established that `example` may
be omitted from create and update requests; responses normalize it to `false`. The
tested requests accepted the field but did not require it. Responses also add a
`contentType` field for the same content-type identity. Update requests containing both
`contentType` and the `entityType`/`entityId` identity form are rejected with `400
ContentPreviewChangeInvalid`.

Request and response representations are therefore asymmetric. An update can use
`entityType` and `entityId` while omitting the response alias `contentType` and the
optional `example` member. Serializing an entire response configuration back into PUT
can submit both identities and trigger validation failure. The create-time `contentType`
alias is observed behavior, not evidence of a second resource identity.

Omitting `description` normalized it to an empty string; sending JSON `null` produced
`503 UnknownError`. Empty configuration lists are accepted on create; on update, an
empty or omitted list does not remove existing configurations. Duplicate content-type
identities are rejected with `400 ContentPreviewChangeInvalid` on both create and update.

## Configuration lifecycle and Web App requests

Contentful merges configuration updates by `entityType` and `entityId`:

- changing `url` or `enabled` updates the existing configuration for that identity;
- a previously unseen identity is appended;
- omitted identities are retained, so an empty update list does not clear configurations; and
- duplicate identities are rejected with `400 ContentPreviewChangeInvalid`.

The Contentful Web App updates a preview environment with `PUT
/spaces/{space_id}/preview_environments/{preview_environment_id}` and a body containing all mutable
top-level fields:

```json
{
  "name": "Example preview",
  "description": "",
  "configurations": []
}
```

The observed UI actions populated that array as follows:

| UI action | Configurations sent |
| --- | --- |
| Select the URL for `content-type-a` | One `content-type-a` record with the selected URL and `enabled: true` |
| Also tick `content-type-b` | Enabled `content-type-a` and `content-type-b` records with the same URL |
| Untick `content-type-a` | Disabled `content-type-a` and enabled `content-type-b` records |
| Remove the preview URL | Disabled `content-type-a` and `content-type-b` records; neither is omitted |
| Re-enable `content-type-a` | The existing identity and URL with `enabled: true` |
| Assign another URL to disabled `content-type-a` | The same identity, replacement URL, and `enabled: true` |

Each record also contained `entityType: "ContentType"` and `example: false`. For
example, unticking `content-type-a` sent:

```json
{
  "name": "Example preview",
  "description": "",
  "configurations": [
    {
      "url": "https://example.invalid/preview/{entry.sys.id}",
      "entityId": "content-type-a",
      "entityType": "ContentType",
      "enabled": false,
      "example": false
    },
    {
      "url": "https://example.invalid/preview/{entry.sys.id}",
      "entityId": "content-type-b",
      "entityType": "ContentType",
      "enabled": true,
      "example": false
    }
  ]
}
```

Disabled configurations are durable resource state. They remained present in a fresh CMA
`GET`, survived a full Web App reload while being hidden by its active-configuration UI,
and were available for later re-enablement.

On a single-content-type probe, assigning a different URL to a previously disabled
identity sent:

```json
{
  "name": "Example preview",
  "description": "",
  "configurations": [
    {
      "url": "https://example.invalid/replacement/{entry.sys.id}",
      "entityId": "content-type-a",
      "entityType": "ContentType",
      "enabled": true,
      "example": false
    }
  ]
}
```

The response contained only the replacement configuration, not separate old and new
configurations. This confirms that URL is mutable data and not part of configuration
identity.

Unticking a content type or removing its preview URL is therefore an in-place update
represented by `enabled: false`; omission is a no-op. The probes establish persistence
across subsequent reads, reload, and updates, but not a guaranteed retention period. The
undocumented API exposes no expiry or cleanup metadata, so maintainers should treat
disabled configurations as persistent until direct evidence establishes another
lifecycle.

## Update, ordering, and replacement behavior

Creation preserves submitted configuration order. Submitting existing configurations in
another order on update does not reorder them. Order is therefore observable on the
wire, but Contentful's [content preview
documentation](https://www.contentful.com/developers/docs/tutorials/preview/content-preview/)
assigns it no product meaning, and no practitioner-visible consequence was established.

Selected-ID recreation has a separate service-side history constraint after the preview
environment itself is deleted:

- deleting an empty platform permits immediate recreation under the same ID with either an empty or non-empty configuration list;
- deleting a platform with non-empty configurations causes immediate non-empty recreation to fail with `400 ContentPreviewChangeInvalid`;
- recreating that ID with an empty list succeeds, but subsequently adding a configuration still returned `400` after 0, 5, 15, and 30 seconds.

The failed reuse indicates separate backend residue associated with the deleted selected
ID; unlike disabled configurations on a live resource, that residue was inferred rather
than returned in a representation. Empty recreation is not a verified migration path
back to a configured platform. Using a new ID avoids the observed reuse condition;
successful configured reuse of the deleted ID was not established.

Create with a selected ID uses the same `PUT` operation as update. A PUT intended to
create at an existing version-0 address succeeded as an update, retained an omitted
existing configuration, changed metadata, and incremented the version. The observed PUT
did not enforce create-only intent. Callers cannot use it to claim a fresh resource at
an address that may already exist.

The undocumented route does not fully follow the generic CMA selected-ID rules. The [CMA
overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/)
documents 1–64 characters and alphanumeric, dot, hyphen, or underscore characters for
resource IDs. Direct preview-environment probes established this accepted envelope: 1–64
ASCII alphanumeric, hyphen, or underscore characters. Uppercase letters and leading or
trailing hyphens and underscores were accepted. Outside that envelope:

- 65 characters, spaces, and `@` were rejected with `400`; and
- a dot produced `404 UnknownRoute`.

In particular, the generic CMA allowance for dots does not apply to this path as routed
in production.

## Concurrency and errors

The [CMA
overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/)
describes optimistic locking through `X-Contentful-Version`. Preview-environment
metadata updates increment `sys.version`, and a request with an older version then
returns `409 Conflict`. Configuration-only updates do not increment the version:
multiple configuration changes using the original version `0` succeeded. Version locking
therefore protects metadata changes but cannot detect concurrent configuration-only
changes. The tested API also accepted updates without the header.

Sending only changed configuration identities limits the set of values a PUT can
overwrite. It cannot detect another writer changing the same identity:
configuration-only changes did not advance `sys.version`. A response or later read may
reveal a concurrent change, but the observed version header alone does not make
configuration updates conflict-safe.

Observed error behavior:

| Scenario | Observed result |
| --- | --- |
| Read or delete missing item | 404 |
| Stale metadata update version | 409 `Conflict` |
| Both configuration identity forms on update | 400 `ContentPreviewChangeInvalid` |
| Duplicate content-type configurations | 400 `ContentPreviewChangeInvalid` |
| Null description | 503 `UnknownError` |
| Non-empty recreation after deleting a configured selected ID | 400 `ContentPreviewChangeInvalid` |

Deletion can be briefly read-after-delete inconsistent. A successful DELETE did not
always make the next GET return `404`; immediate read visibility is not guaranteed by
these observations.

## Unresolved behavior

The evidence does not establish a complete URL-scheme or placeholder grammar,
content-type existence validation, authorization and entitlement matrix, all ID
characters, regional parity, or additional entity types. Product limits do not establish
the exact behavior above those limits.

Platform ordering and default selection, [environment UI preview mode](https://www.contentful.com/developers/docs/references/content-management-api/ui-config/get-the-ui-config/), and custom preview
tokens are separate concerns. The token storage observations are in [Live preview
variables](live-preview-variables.md). Disabled-record retention, deleted-ID reuse after
longer intervals, and stronger concurrency mechanisms remain unresolved.
