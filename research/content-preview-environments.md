# Content preview environment CMA contract

Last verified: 2026-07-17

## Status and terminology

Contentful documents content previews as a product feature, but its public Content Management API reference does not document the `PreviewEnvironment` entity or `/preview_environments` endpoints. This note records the production behavior on which the provider depends.

A preview environment is space-level configuration for a content preview platform. It maps content types to preview URL templates. It is not a Contentful sandbox environment and is not the Content Preview API that serves draft entries.

Primary sources:

- [Set up content preview](https://www.contentful.com/developers/docs/tutorials/preview/content-preview/)
- [Content Management API reference](https://www.contentful.com/developers/docs/references/content-management-api/)
- [Content Preview API overview](https://www.contentful.com/developers/docs/references/content-preview-api/overview/)
- Direct authenticated observations of the production Content Management API and Contentful Web App on 2026-07-14, 2026-07-15, and 2026-07-17

Because the HTTP API is undocumented, the generated client isolates its wire contract and the live acceptance suite should retain focused contract coverage.

## Observed endpoints

All paths are relative to the configured Content Management API base URL.

| Operation | Method and path | Success response |
| --- | --- | --- |
| List | `GET /spaces/{space_id}/preview_environments` | `200`; offset collection |
| Read | `GET /spaces/{space_id}/preview_environments/{id}` | `200` |
| Create with generated ID | `POST /spaces/{space_id}/preview_environments` | `201`; version `0` |
| Create with selected ID | `PUT /spaces/{space_id}/preview_environments/{id}` | `200`; version `0` |
| Update | `PUT /spaces/{space_id}/preview_environments/{id}` | `200` |
| Delete | `DELETE /spaces/{space_id}/preview_environments/{id}` | `204` |

List requests honor `skip` and `limit`. Cursor pagination, filtering, sorting, and behavior above the documented product limit were not verified.

## Representation and request normalization

The resource contains `name`, `description`, a `configurations` array whose order is observable, and `sys` metadata. Each configuration represents a content type, URL template, and enabled state. The observed configuration identity is the pair `entityType` and `entityId`; the URL and enabled state are mutable values for that identity.

The service accepts either `contentType` or `entityType: "ContentType"` with `entityId` when creating a configuration. The Contentful Web App uses this request representation:

```json
{
  "url": "https://example.invalid/preview/{entry.sys.id}",
  "entityId": "author",
  "entityType": "ContentType",
  "enabled": true,
  "example": false
}
```

The Web App sends `example: false`, but direct API probes established that `example` may be omitted from create and update requests; responses normalize it to `false`. The field is accepted on requests but is not required or user-managed provider input. Responses also add a `contentType` field for the same content-type identity. Update requests containing both `contentType` and the `entityType`/`entityId` identity form are rejected with `400 ContentPreviewChangeInvalid`.

The provider therefore:

- uses operation-specific create, update, and response models;
- uses only `entityType` and `entityId` as request identity fields;
- normalizes either response identity to `content_type_id`;
- omits `contentType` and `example` from requests; and
- never serializes a response object directly into an update request.

Omitting `description` normalizes it to an empty string. Sending JSON `null` produced `503 UnknownError`, so requests always send a string. Empty configuration lists are accepted on create; on update, an empty or omitted list does not remove existing configurations. Duplicate content-type identities are rejected with `400 ContentPreviewChangeInvalid` on both create and update.

## Configuration lifecycle and Web App requests

Contentful merges configuration updates by `entityType` and `entityId`:

- changing `url` or `enabled` updates the existing configuration for that identity;
- a previously unseen identity is appended;
- omitted identities are retained, so an empty update list does not clear configurations; and
- duplicate identities are rejected with `400 ContentPreviewChangeInvalid`.

The Contentful Web App updates a preview environment with `PUT /spaces/{space_id}/preview_environments/{id}` and a body containing all mutable top-level fields:

```json
{
  "name": "Probe",
  "description": "",
  "configurations": []
}
```

The observed UI actions populated that array as follows:

| UI action | Configurations sent |
| --- | --- |
| Select the URL for `author` | One `author` record with the selected URL and `enabled: true` |
| Also tick `centre` | Enabled `author` and `centre` records with the same URL |
| Untick `author` | Disabled `author` and enabled `centre` records |
| Remove the preview URL | Disabled `author` and `centre` records; neither is omitted |
| Re-enable `author` | The existing identity and URL with `enabled: true` |
| Assign another URL to disabled `author` | The same identity, replacement URL, and `enabled: true` |

Each record also contained `entityType: "ContentType"` and `example: false`. For example, unticking `author` sent:

```json
{
  "name": "Probe",
  "description": "",
  "configurations": [
    {
      "url": "https://example.invalid/preview/{entry.sys.id}",
      "entityId": "author",
      "entityType": "ContentType",
      "enabled": false,
      "example": false
    },
    {
      "url": "https://example.invalid/preview/{entry.sys.id}",
      "entityId": "centre",
      "entityType": "ContentType",
      "enabled": true,
      "example": false
    }
  ]
}
```

Disabled configurations are durable resource state. They remained present in a fresh CMA `GET`, survived a full Web App reload while being hidden by its active-configuration UI, and were available for later re-enablement.

On a single-content-type probe, assigning a different URL to a previously disabled identity sent:

```json
{
  "name": "Probe",
  "description": "",
  "configurations": [
    {
      "url": "https://example.invalid/replacement/{entry.sys.id}",
      "entityId": "author",
      "entityType": "ContentType",
      "enabled": true,
      "example": false
    }
  ]
}
```

The response contained only the replacement configuration, not separate old and new configurations. This confirms that URL is mutable data and not part of configuration identity.

Unticking a content type or removing its preview URL is therefore an in-place update represented by `enabled: false`; omission is a no-op. The probes establish persistence across subsequent reads, reload, and updates, but not a guaranteed retention period. The undocumented API exposes no expiry or cleanup metadata, so maintainers should treat disabled configurations as persistent until direct evidence establishes another lifecycle.

The provider exposes only active content-type configurations as a map keyed by content type ID. It keeps `entityType` and `enabled` behind the provider interface: adding a key or changing its URL sends `enabled: true`, removing a key sends `enabled: false`, and unchanged identities are omitted from the update payload. Reads and imports filter disabled configurations from Terraform state. Re-adding a disabled key either re-enables the retained identity or recreates it if the service no longer retains that disabled record.

## Update, ordering, and replacement behavior

Creation preserves submitted configuration order. Submitting existing configurations in another order on update does not reorder them. Order is therefore observable on the wire, but Contentful's [content preview documentation](https://www.contentful.com/developers/docs/tutorials/preview/content-preview/) assigns it no product meaning, and no practitioner-visible consequence was established. The provider models content-type configurations as a map and sorts changed keys only to make request construction deterministic.

Selected-ID recreation has a separate service-side history constraint after the preview environment itself is deleted:

- deleting an empty platform permits immediate recreation under the same ID with either an empty or non-empty configuration list;
- deleting a platform with non-empty configurations causes immediate non-empty recreation to fail with `400 ContentPreviewChangeInvalid`;
- recreating that ID with an empty list succeeds, but subsequently adding a configuration still returned `400` after 0, 5, 15, and 30 seconds.

The failed reuse indicates separate backend residue associated with the deleted selected ID; unlike disabled configurations on a live resource, that residue was inferred rather than returned in a representation. Empty recreation is not a verified migration path back to a configured platform. A replacement whose old selected-ID object had configurations and whose target is non-empty requires a new ID.

Create with a selected ID uses the same `PUT` operation as update. A simulated create-before-destroy request against an existing version-0 object succeeded as an update, retained an omitted existing configuration, changed metadata, and incremented the version. The endpoint cannot distinguish creation from update, so selected-ID replacement must not issue its create request before the old object is deleted.

The undocumented route does not fully follow the generic CMA selected-ID rules. The [CMA overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/) documents 1–64 characters and alphanumeric, dot, hyphen, or underscore characters for resource IDs. Direct preview-environment probes established this accepted envelope: 1–64 ASCII alphanumeric, hyphen, or underscore characters. Uppercase letters and leading or trailing hyphens and underscores were accepted. Outside that envelope:

- 65 characters, spaces, and `@` were rejected with `400`; and
- a dot produced `404 UnknownRoute`.

In particular, the generic CMA allowance for dots does not apply to this path as routed in production.

## Concurrency and errors

The [CMA overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/) describes optimistic locking through `X-Contentful-Version`. Preview-environment metadata updates increment `sys.version`, and a request with an older version then returns `409 Conflict`. Configuration-only updates do not increment the version: multiple configuration changes using the original version `0` succeeded. Version locking therefore protects metadata changes but cannot detect concurrent configuration-only changes. The API also accepts updates without the header, but the provider sends the last observed version and surfaces conflicts instead of refreshing and replaying automatically.

The provider constructs configuration updates from the Terraform state-to-plan delta without an additional read. This prevents unrelated concurrent configuration changes from being included in the request; a later refresh exposes them as drift. A concurrent change to the same identity Terraform is updating remains last-writer-wins because the service does not advance `sys.version` for configuration-only changes.

Observed error behavior:

| Scenario | Status | Provider behavior |
| --- | ---: | --- |
| Read missing item | `404` | Remove the resource from state |
| Delete missing item | `404` | Treat it as already absent |
| Stale update version | `409` | Return the conflict diagnostic |
| Both configuration identity forms on update | `400` | Prevent through request normalization |
| Duplicate content-type configurations | `400` | Validate before request construction |
| Null description | `503` | Prevent by sending an empty string |
| Non-empty recreation after deleting configured selected ID | `400` | Require a new selected ID |

Deletion can be briefly read-after-delete inconsistent. Live-test cleanup polls until reads return `404`.

## Maintenance constraints

- Keep live coverage focused on the undocumented contract and perform broader lifecycle cases against the deterministic mock server.
- Do not add provider-side validation for URL schemes, placeholder grammar, remote content-type existence, entitlements, names, or unverified ID characters without published or directly verified constraints.
- Keep `sys.version`, timestamps, response aliases, `example`, platform ordering, preview mode, and custom preview tokens out of this resource.
- Treat platform ordering/default selection, space-wide preview mode, and custom preview tokens as separately owned concerns requiring their own API research.
- Reverify this contract before expanding scope, particularly for authorization requirements, EU data-residency endpoints, new entity types, or ordering APIs.
