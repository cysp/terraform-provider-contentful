# Delivery API keys: environment selection and versions

A Delivery API key selects environments through an `environments` array of Contentful
Environment links. Omission, an empty array, explicit links, and JSON null have distinct
request representations. The observed service defaulted both omission and an empty array
to the built-in `master` environment identifier.

## Addressing and representation

Create uses `POST /spaces/{space_id}/api_keys`; GET, PUT, and DELETE address
`/spaces/{space_id}/api_keys/{api_key_id}`. An environment selection has this synthetic
representation:

```json
{"environments": [{"sys": {"type": "Link", "linkType": "Environment", "id": "environment-id"}}]}
```

## Direct CMA observations

Observed against the Contentful CMA on 2026-07-28 using temporary Delivery API keys:

| Request value | Create | Update from `[environment-id]` |
|---|---|---|
| property omitted | returns `[master]` | returns `[master]` |
| `[]` | returns `[master]` | returns `[master]` |
| explicit environment links | returns those links | returns those links |
| `null` | HTTP 400 | HTTP 400; existing value is unchanged |

Successful create, update, and subsequent get responses contained an `environments`
array. The linked Preview API key reflected the same environment selection as its
Delivery API key.

These are point-in-time observations rather than a published compatibility guarantee. In
particular, the API could later preserve an empty list or choose a default environment
with a different identifier.

## Interpretation and limits

An omitted member on update did not preserve the prior selection. Explicit null was
rejected rather than treated as omission or clearing. Consumers comparing requested and
returned configuration need to retain those request distinctions and inspect the
returned links instead of reconstructing an assumed default.

## Supporting documentation

- Contentful documents CMA updates as full replacements rather than merges:
  [Updating content](https://www.contentful.com/developers/docs/references/content-management-api/overview/#updating-content).
- Contentful's official Ruby SDK defaults an omitted create argument to `[]` and
  notes that an empty value defaults to master:
  [`ApiKey`](https://github.com/contentful/contentful-management.rb/blob/46060ea8341350ebb15753ff9a66990cefbc8e71/lib/contentful/management/api_key.rb).
- The REST examples show `environments` represented as an array of Environment
  links:
  [create](https://www.contentful.com/developers/docs/references/content-management-api/api-keys/create-a-delivery-api-key/),
  [update](https://www.contentful.com/developers/docs/references/content-management-api/api-keys/update-a-delivery-api-key/), and
  [get](https://www.contentful.com/developers/docs/references/content-management-api/api-keys/get-a-delivery-api-key/).

## Versioning observations

Experiment date unrecorded; results retained by 2026-08-24.

A disposable Delivery API key was created in an existing disposable test space, updated
once, sent one stale update, and deleted. Only status codes, `sys.type`, `sys.version`,
response-member presence, and the error classification were recorded; complete response
bodies were not retained.

| Request | Status | Structural observation |
| --- | ---: | --- |
| `POST /spaces/{space_id}/api_keys` | 201 | `sys.type: ApiKey`, `sys.version: 0`; delivery and preview token members were present but not read or recorded |
| `PUT /spaces/{space_id}/api_keys/{api_key_id}` with `X-Contentful-Version: 0` | 200 | `sys.type: ApiKey`, `sys.version: 1`; token members remained present but were not read or recorded |
| Repeated PUT with stale `X-Contentful-Version: 0` | 409 | `sys.type: Error`, `sys.id: Conflict`, nonempty message, no details member |
| `DELETE /spaces/{space_id}/api_keys/{api_key_id}` | 204 | Empty response; the disposable key was removed |

The initial zero version and `Conflict` error ID are specific to this observed
endpoint. They must not be replaced with assumptions from Entry or Taxonomy examples.
The official [update
reference](https://www.contentful.com/developers/docs/references/content-management-api/api-keys/update-a-delivery-api-key/)
and pinned [API-key
adapter](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/api-key.ts#L26-L76)
establish the versioned update request.
