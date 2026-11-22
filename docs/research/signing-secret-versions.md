# Signing secret versions and timestamps

The recorded AppSigningSecret and WebhookSigningSecret responses omitted
`sys.version`, and the tested numeric version headers did not prevent replacement.
Both endpoints changed creation and update timestamps after replacement. These
observations establish neither a supported conditional-write mechanism nor a
reliable ordering of rapid writes.

The WebhookSigningSecret reference includes a version example that conflicts with
these responses and the first-party SDK. Public sources were reviewed on
2026-09-30; the observations below retain their experiment dates and scope.

## Public reference and first-party SDK

| Evidence | AppSigningSecret | WebhookSigningSecret |
| --- | --- | --- |
| GET response example | Includes timestamps and parent links, but no secret-specific `sys.version` or `sys.id`. [GET reference][app-get] | Includes `sys.version: 1`, timestamps, and a Space link, but no secret-specific `sys.id`. [GET reference][webhook-get] |
| PUT response example | No `sys.version`; request example supplies only `value` and no version header. [PUT reference][app-put] | Includes `sys.version: 1`; request example supplies only `value` and no version header. [PUT reference][webhook-put] |
| SDK response type | Explicitly excludes `version` and `id` from the common system metadata type. [Entity definition][app-entity] | Explicitly excludes `version` from the common system metadata type. [Entity definition][webhook-entity] |
| SDK mutation implementation | PUT and DELETE supply no version or conditional header. [REST adapter][app-adapter] | Signing-secret PUT and DELETE supply no version or conditional header. The separate WebhookDefinition update does send `X-Contentful-Version`; that does not establish a signing-secret requirement. [REST adapter][webhook-adapter] |

The SDK sources above are pinned to commit
`5ed646919c0663e2da7676979afae0fd15c11587`. Type declarations establish SDK
expectations, not every field the server could return. A version in one response
example does not establish
that it changes after replacement or can enforce a conditional mutation.

## Recorded direct observations

The [WebhookSigningSecret experiment](webhook-signing-secret.md#direct-observations)
on **2026-09-14** explicitly examined version metadata after creation and repeated
replacement. Successful PUT responses and subsequent GETs omitted `sys.version`.
Replacement succeeded with `X-Contentful-Version` values `1`, `0`, and
`2147483647`; the last two also failed to prevent deletion. Tested `If-Match`
and `If-None-Match` conditions likewise did not prevent the documented mutations.
This demonstrates the behavior of those tested headers on that endpoint at that
time. Without a returned version, the experiment cannot classify `1` as a current
or stale version, and it does not exclude other conditional mechanisms.

### AppSigningSecret metadata and replacement

On **2026-09-30**, an experiment inspected raw JSON and response headers using the
standard-library HTTP client, bypassing the generated provider and SDK models.
Requests were sequential, with no automatic retries or redirects. A read-only GET
of an existing AppSigningSecret returned 200. A separate disposable AppDefinition
was then created for the mutation experiment.

| Request | Response | Structural result |
| --- | --- | --- |
| First signing-secret PUT | 201 | No `sys.version` or version-like response header |
| GET after first PUT | 200 | Same metadata shape; redacted suffix matched the supplied value's suffix |
| PUT with a different fresh value | 200 | Replacement acknowledged; both `createdAt` and `updatedAt` changed |
| GET after replacement | 200 | Same metadata shape; redacted suffix matched the replacement's suffix |
| PUT with a third fresh value and `X-Contentful-Version: 2147483647` | 200 | Replacement acknowledged; both `createdAt` and `updatedAt` changed again |
| GET after conditional-header probe | 200 | Same metadata shape; redacted suffix matched the third value's suffix |
| DELETE signing secret, then GET | 204, then 404 | Secret absent after deletion |

All successful signing-secret PUT and GET responses, including the initial
read-only GET, had these `sys` keys: `appDefinition`, `createdAt`, `createdBy`,
`organization`, `type`, `updatedAt`, and `updatedBy`. The JSON omitted
`sys.version`. Response headers included neither `ETag` nor `Last-Modified`, nor
any header whose name contained `version`. The suffix comparisons above used the
subsequent GET responses.

The replacement writes were separated by 1.1 seconds. Timestamp changes are
therefore observed for these separated writes only; this does not establish
uniqueness for rapid writes, monotonicity, conditional-request support, or
same-value PUT behavior. In particular, `createdAt` changing on replacement
prevents treating it as a stable singleton creation timestamp on this evidence.

The numeric request header did not prevent the tested replacement. With no
returned version, it cannot be called a stale-version test. No app `If-Match`,
`If-None-Match`, or version-conditioned DELETE test was performed. The suffix
checks confirm the expected redacted representation, not equality of complete
stored secret bytes.

### WebhookSigningSecret metadata and same-value replacement

On **2026-09-30**, an experiment inspected raw JSON and response headers for a
temporary webhook signing secret. It bypassed the generated provider and SDK
models, used sequential standard-library HTTP requests without automatic retries
or redirects, and confirmed absence with an initial 404 `NotFound` response before
creating the secret.

| Request | Response | Structural result |
| --- | --- | --- |
| Initial GET | 404 `NotFound` | No signing secret present |
| First PUT, then GET | 201, then 200 | Secret created; successful-response metadata had no version |
| PUT with the same value, then GET | 200, then 200 | Both `createdAt` and `updatedAt` changed despite unchanged supplied bytes |
| PUT with a different fresh value, then GET | 200, then 200 | Both timestamps changed again |
| PUT with another fresh value and `X-Contentful-Version: 2147483647`, then GET | 200, then 200 | Header did not prevent replacement; both timestamps changed again |
| DELETE, then GET | 204, then 404 `NotFound` | Secret absent after deletion |

All eight successful PUT and GET responses had exactly these `sys` keys:
`createdAt`, `createdBy`, `space`, `type`, `updatedAt`, and `updatedBy`. None
included `sys.version`, an `ETag` or `Last-Modified` response header, or a
response header whose name contained `version`. Each PUT's `sys` object equalled
the following GET's `sys` object. Both PUT and GET redacted-suffix comparisons
matched the supplied value after each write.

Writes were separated by 1.1 seconds. The timestamps changed after all three
subsequent PUTs, including the same-value PUT. They therefore did not distinguish
a secret-value change from resubmission in this experiment. This does not
establish uniqueness for rapid writes, monotonicity, or support for timestamp
preconditions. As with the app experiment, the numeric header cannot be classified
as stale without a returned version. Its observed failure to prevent this PUT
does not establish the behavior of other conditions or future service versions.

## Limits of the evidence

The effect of same-value AppSigningSecret PUT on timestamps remains unverified.
The [signature experiment](app-framework/request-signing.md#complete-key-verification-after-rotation-and-an-uncertain-write)
confirmed acceptance and signing with the resubmitted key, without examining its
timestamp effects. Neither metadata experiment establishes a supported
conditional-write mechanism or reliable timestamp ordering for rapid writes.
App conditional DELETE was not tested.

[app-get]: https://www.contentful.com/developers/docs/references/content-management-api/app-signing-secret/get-the-current-app-signing-secret/
[app-put]: https://www.contentful.com/developers/docs/references/content-management-api/app-signing-secret/create-or-overwrite-the-app-signing-secret/
[webhook-get]: https://www.contentful.com/developers/docs/references/content-management-api/webhook-security/get-a-webhook-signing-secret/
[webhook-put]: https://www.contentful.com/developers/docs/references/content-management-api/webhook-security/create-a-webhook-signing-secret/
[app-entity]: https://github.com/contentful/contentful-management.js/blob/5ed646919c0663e2da7676979afae0fd15c11587/lib/entities/app-signing-secret.ts
[app-adapter]: https://github.com/contentful/contentful-management.js/blob/5ed646919c0663e2da7676979afae0fd15c11587/lib/adapters/REST/endpoints/app-signing-secret.ts
[webhook-entity]: https://github.com/contentful/contentful-management.js/blob/5ed646919c0663e2da7676979afae0fd15c11587/lib/entities/webhook.ts
[webhook-adapter]: https://github.com/contentful/contentful-management.js/blob/5ed646919c0663e2da7676979afae0fd15c11587/lib/adapters/REST/endpoints/webhook.ts
