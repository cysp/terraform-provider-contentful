# Specified-ID PUT header semantics

Reviewed 2026-09-02. This note distinguishes Contentful's documented contract,
first-party client behavior, and sanitized direct observations of the standard
global Content Management API (CMA).

## Contract

Contentful uses the same Entry `PUT` endpoint to create an Entry with a
specified ID and to update an existing Entry. The official
[Entry reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/)
says that Create requires `X-Contentful-Content-Type`, while Update requires the
last `X-Contentful-Version` and does not require the Content Type header. The
[endpoint reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/create-an-entry-with-a-specified-id/)
describes the same distinction.

Contentful's current first-party management SDK implements that contract at
commit `cc096a337f0e1db6114e8da645d69bb6eb90f11c`. Entry
[`createWithId`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L255-L281)
sends `X-Contentful-Content-Type` and no version, while
[`update`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L155)
sends the observed `sys.version` and no Content Type header. Content Type
[`createWithId`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/content-type.ts#L71-L80)
and
[`update`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/content-type.ts#L82-L96)
likewise distinguish Create from Update by omitting or supplying the version.

## Entry direct observation

The probe used credentials from the ignored example repository. The available
relevant variable names were `CONTENTFUL_MANAGEMENT_ACCESS_TOKEN`,
`CONTENTFUL_SPACE_ID`, `CONTENTFUL_ENVIRONMENT_ID`, and
`CONTENTFUL_ORGANIZATION_ID`; no values were logged or retained. It selected the
first active Content Type from a read-only environment listing. The Content
Type identity was not retained, and no Content Type was created, activated, or
modified.

Every request used the member path
`PUT /spaces/{space_id}/environments/{environment_id}/entries/{entry_id}`, the
management JSON `Content-Type`, and the independent body `{"fields":{}}`.
Every target had a randomized disposable ID and returned `404 NotFound` to a
preflight `GET`. Successful requests returned the selected Content Type
identity, an empty `fields` object, and no `publishedVersion`; the table labels
the Content Type symbolically as `fixture`.

| Target before `PUT` | `X-Contentful-Content-Type` | `X-Contentful-Version` | `PUT` result | Follow-up `GET` |
| --- | --- | --- | --- | --- |
| Absent | `fixture` | absent | `201`, version `1` | `200`, version `1` |
| Absent | absent | `1` | `400 BadRequest`: “You should provide a content type in X-Contentful-Content-Type request header.” | `404`, still absent |
| Absent | `fixture` | `1` | `201`, version `1` | `200`, version `1` |
| Absent | absent | absent | same `400 BadRequest` | `404`, still absent |
| Absent | absent | `7` | same `400 BadRequest` | `404`, still absent |
| Absent | `fixture` | `7` | `201`, version `1` | `200`, version `1` |
| Existing disposable draft at returned version `1` | `fixture` | exact returned version `1` | `200`, version `2` | `200`, version `2` |
| Existing disposable draft at returned version `1` | absent | exact returned version `1` | `200`, version `2` | `200`, version `2` |

Every failed `PUT` returned error `sys.id` `BadRequest`, the message quoted in
the table, and null `details`. Each failed request's follow-up `GET` returned
error `sys.id` `NotFound`, message “The resource could not be found.”, and
details identifying the symbolic Entry, space, and environment target.

For the observed absent targets, Content Type presence selected the create
behavior: versions `1` and `7` did not prevent creation when Content Type was
present, and neither version authorized creation without Content Type. For the
observed existing drafts, the exact returned version selected Update and the
Content Type header made no difference. This makes omission of
`X-Contentful-Content-Type` from provider Update a safety boundary as well as a
smaller request: a stale Update against a remotely deleted target fails instead
of recreating it.

All disposable Entries created for these observations were deleted, subsequent
`GET` requests returned `404`, none was published, and no probe Entry remained.

### Earlier Entry and Content Type version observation

An earlier sanitized probe kept the Entry Content Type header present and
compared only version-header behavior for Entries and Content Types:

| Target before `PUT` | Version header absent | `X-Contentful-Version: 1` |
| --- | --- | --- |
| Absent | Both resources: `201`, created at version `1` | Both resources: `201`, created at version `1` |
| Existing draft at version `1` | Both resources: `409 VersionMismatch`; subsequent `GET` unchanged | Both resources: `200`, updated to version `2` |

The new Entry matrix narrows the unresolved point from that probe: on an absent
Entry, the observed Content Type header is necessary for creation and an
otherwise valid version does not substitute for it.

## Provider invariants

- Specified-ID Entry Create sends `X-Contentful-Content-Type` and omits
  `X-Contentful-Version`. A collision must fail without mutating, publishing, or
  adopting the existing Entry.
- Entry Update sends the exact prior `sys.version` and omits
  `X-Contentful-Content-Type`. Ordinary existing-target behavior is unchanged;
  an absent target now fails rather than being recreated and then published.
- Specified-ID Content Type Create omits `X-Contentful-Version`; Content Type
  Update sends the exact observed version.
- Entry specified-ID Create and Update and Content Type Create and Update do not
  transparently replay 429, transport, or 5xx outcomes. Only a complete,
  identity-valid, plan-consistent draft response can authorize its exact
  returned positive version for Publish or Activate; a later `GET` cannot do so.
- A preflight read cannot provide the same create/update safety because another
  writer could change target existence between the read and the `PUT`.

The CMA test handler should preserve this request-selection boundary explicitly:

- absent Entry plus Content Type creates, regardless of an accompanying
  version header;
- absent Entry without Content Type returns the observed `400 BadRequest` and
  remains absent;
- existing Entry plus an exact version updates, with or without Content Type;
  and
- existing Entry without an exact version returns `409 VersionMismatch` and
  remains unchanged.

## Limits

The direct Entry matrix covered one CMA account, space, environment, active
Content Type, and the standard global endpoint. It exercised empty-field draft
Entries, absent targets with positive version values `1` and `7`, and existing
drafts at version `1`. It did not cover the EU endpoint, published or archived
Entries, nonempty field validation, or malformed, zero, negative, or stale
versions. Exact statuses, messages, and `1` to `2` arithmetic are observed
behavior rather than permanent API guarantees. The durable compatibility rule
is the create/update header distinction supported independently by Contentful's
documentation and first-party SDK.
