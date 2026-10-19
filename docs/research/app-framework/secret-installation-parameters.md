# Secret installation parameters

An AppInstallation's management representation need not contain the values its app
receives. The distinction matters when comparing a configured parameter object with
the object returned by Contentful, or when building a later write from a management
read.

This reference separates published contracts and inspected first-party client
behavior from service observations. See [configuration and installation](configuration.md)
for ordinary parameter replacement and [research evidence conventions](../README.md#evidence-and-redaction).
Mutable documentation was reviewed on 2026-10-20. SDK references are pinned to
`883e2b9dc1c76413d5c24e45f74243da699071e4`; their behavior is client evidence, not proof
of server acceptance. Provider policy belongs in [provider design](../../design/README.md).

## Published Contentful contract

An AppDefinition declares installation parameters by `id` and `type`; `Secret` is an
installation-only type. Contentful identifies protected values by matching parameter
keys against those declaration IDs. When installation parameter definitions are
omitted from the AppDefinition, the guide describes values as a free-form object;
it does not establish that extra undeclared keys are accepted alongside configured
declarations. The guide describes redaction in CMA reads using a personal access token
and in the App SDK inside the Contentful UI; App Identities, Events, and Functions can
receive raw values. It does not define the wire representation of redaction or the
meaning of omitted, null, empty, or placeholder Secret values on update.
[App Parameters][parameters]

AppInstallation reflects changes to its AppDefinition immediately. Definition metadata
therefore describes current visibility, not necessarily the visibility when a value
was written. The documentation also says deleting a definition removes its
installations and cloning an environment copies installations and their parameters.
It does not explain storage migration or disclosure behavior when a declaration
changes between `Secret` and an ordinary type. [App Installation][installation]

The installation write endpoint is
`PUT /spaces/{space_id}/environments/{environment_id}/app_installations/{app_definition_id}`.
Its example submits and returns a `parameters` object. The endpoint reference does
not specify Secret-specific replacement, preservation, deletion, or optimistic
concurrency behavior. [Install or update an app][put]

App Identity uses short-lived app access tokens obtained using app keys. The identity
can operate on its own AppInstallation. Combined with the documented Secret
visibility distinction, this provides an independent observation channel for the
experiments below: compare a PAT read with the app's own read without changing the
definition to expose its secrets.
[App Identity][identity], [App Parameters][parameters]

## First-party SDK evidence

The inspected `contentful-management.js` REST adapter copies the supplied installation
payload and sends it via PUT. It does not merge parameters with an earlier read,
look up Secret declarations, or generate a version precondition. Caller-supplied
headers can be forwarded; the adapter can additionally send Marketplace agreement
headers. Its GET simply retrieves the addressed installation.
[Installation REST adapter][sdk-adapter]

The object-oriented `AppInstallation.update()` method sends the object's plain
representation to that upsert operation. The entity has an optional `parameters`
property; its type is not a separate read/write Secret model. These methods provide
no client-side reconstruction of values withheld by a preceding read. They also do
not prove that a server would remove or retain a missing Secret during the write.
[Installation entity][sdk-entity]

The SDK distinguishes installation declaration types from ordinary parameter types
by adding `Secret`. Parameter declarations contain `id`, `name`, `type`, and optional
configuration metadata. The value type remains free-form and contains no explicit
redaction discriminator. [Parameter types][sdk-parameters]

## Direct CMA observations

Experiments on 2026-10-20 used a personal access token, a disposable AppDefinition,
and its installation in one concrete environment. The definition initially declared two
optional `Secret` parameters (`secret`, `otherSecret`) and one optional `Symbol`
(`ordinary`). Every submitted value was synthetic.

Selected PAT reads were paired with reads using an AppAccessToken for the same app.
These independently distinguished deletion, rotation, literal asterisks, and
preservation. Initial creation and whole-object null rejection were checked through
PAT responses only; the tables distinguish those cases.

All PUT/GET installation successes below returned HTTP 200. Invalid parameter
requests returned 422 `ValidationFailed`. For an existing installation, the following
GET retained the prior successful value. No installation response exposed
`sys.version`. These are scoped observations, not a promise about every tenant, token type, or future API version.

### Read and mutation representations

| Operation | PAT representation | App Identity observation |
| --- | --- | --- |
| First successful PUT to an absent installation, with Secret string | Returned submitted plaintext; repeated after deletion and with Terraform | Not separately read before the next mutation |
| PUT to an existing installation, including first value after clearing | Returned asterisks for Secret strings | GET returned submitted plaintext |
| GET with `secret = "alpha"`, then rotate to `"bravo"` | Both returned `"*****"` | Returned `"alpha"`, then `"bravo"` |
| Mixed Secret and ordinary parameters | Secret masked; ordinary value unchanged | Both real values returned |
| Both declared Secret keys, no ordinary key | Both Secret keys retained and masked | Both real values returned |
| Empty Secret string | `""`, key retained | `""`, key retained |
| `secret = "a🔑é"` | `"****"` | Exact submitted string |
| `secret = "a*b"`, `ordinary = "***"` | Both returned `"***"` | Secret remained `"a*b"`; ordinary remained literal `"***"` |

The observed mask length matches UTF-16 code units: `a🔑é` is three Unicode code
points, seven UTF-8 bytes, and four UTF-16 code units. ASCII examples also matched
this rule. This is evidence for that example and rule, not an exhaustive Unicode
normalization specification. A mask is not a hash and does not establish equality.
The ordinary-asterisks case proves that the returned string alone cannot classify a
key as Secret.

A PUT with an undeclared key returned 422, followed by GET 404 when the installation
was previously absent. Successful first installations returned 200, as did updates;
HTTP status alone did not distinguish creation from update. The repeated absent/present
sequence produced plaintext creation responses and masked update responses.

### Indicators in the installation response

A separate PAT-only probe on the same date inspected complete response bodies and
header names, rather than only extracting `parameters`. It used a disposable
definition with `alpha` declared Secret and `beta` declared Symbol, then exercised
creation, GET, rotation, key removal, and an empty Secret. It did not read definition
metadata during those installation operations.

| Response surface | Observed contents |
| --- | --- |
| Top-level body | Only `sys` and `parameters` |
| `sys` | `id`, `type`, `createdBy`, `updatedBy`, `createdAt`, `updatedAt`, and links named `appDefinition`, `organization`, `space`, and `environment` |
| Parameter values | Plain JSON strings, including masks; no per-key type, redacted flag, companion Secret-key list, or redaction object |
| Response headers | Content type, cache/transport/security/CORS metadata, request identification, and rate limits; no Secret classification or redaction indicator |
| Version/precondition metadata | No `sys.version` or `ETag` |

For submitted `alpha = "alpha"` and `beta = "*****"`, GET returned:

```json
{"parameters":{"alpha":"*****","beta":"*****"}}
```

This is the parameter projection; `sys` is omitted from the example. The response
contained no discriminator between the masked Secret and the literal ordinary
string. `sys.appDefinition` was a link, not embedded parameter declarations. Actor
and timestamp metadata described the installation, not which key was Secret.

GET with no explicit `Accept`, with `Accept: application/json`, and with
`Accept: application/vnd.contentful.management.v1+json` produced the same parameter
projection and metadata shape. Successful creation echoed plaintext, while subsequent
PUTs and GETs masked the Secret; no accompanying field or header explained that
transformation. Removing the key and setting it to an empty string introduced no
additional metadata.

The mask pattern and its length can assist comparison with a previously known value,
but cannot establish Secret classification from this response alone. This probe does
not rule out other API versions or undocumented response options; it establishes no
usable discriminator in the exercised detail GET/PUT representations.

### Replacement, deletion, and placeholders

| Request on an existing installation | Observed result |
| --- | --- |
| Send a new non-placeholder Secret string | Replaced the old value |
| Send one Secret and omit the second Secret | Replaced the first and removed the second |
| Send only an ordinary key | Removed the remaining Secret key |
| `parameters: {}` with optional declarations | Removed all keys; both reads returned `{}` |
| Omit `parameters` entirely | Removed the whole parameter property; both reads omitted it |
| `parameters: null` | 422; PAT GET still omitted the previously absent property; no paired App Identity check |
| Secret `null`, number, or object | 422 expecting String; previous value unchanged |
| Secret `""` | Stored empty string, distinct from absent key |
| Existing real Secret `"rotated"`; send `"*******"` | Retained `"rotated"` while ordinary sibling changed |
| Existing real Secret `"rotated"`; send `"***"` | Replaced it with the literal string `"***"` |
| Secret key absent; send `"*******"` | Stored literal `"*******"` |

Except for whole-object null rejection as noted, the table's stored-value outcomes
were verified through App Identity. The exact current mask acted as a preservation
placeholder in these probes. This creates an ambiguity:
a caller trying to set a literal all-asterisk string equal to the current mask cannot
express replacement with that PUT. A different-length mask is ordinary input. An
imported or stale mask must not be treated as recoverable plaintext or as a generally
safe replay token: another actor changing the secret length can turn that stale mask
into a new literal secret.

The observed API replaces the supplied object; it does not retain omitted Secret
keys. A client that blindly merges missing keys would hide real deletion. Conversely,
a client that drops masks before updating ordinary keys would delete those secrets.

### Required declarations and definition changes

Changing `secret.required` to true yielded:

| Request | Result |
| --- | --- |
| Present object omitting `secret`, including `{}` | 422 required-property error; previous installation unchanged |
| `secret: null` | 422 type error |
| `secret: ""` | 200; stored empty string |
| Omit the entire `parameters` property | 200; removed the entire property despite the required declaration |
| Supply a new Secret and subsequently its exact mask | 200; new string stored, then preserved by the mask |

Changing the existing `secret` declaration from `Secret` to `Symbol` made the
previously stored plaintext visible in PAT GET without rewriting the installation.
Changing back to `Secret` masked it again. Removing that declaration also exposed the
existing plaintext while leaving the stored parameter present. These experiments
establish visibility changes for this sequence; they do not establish every type
migration, default, or validation policy.

The same stored value can therefore have different visibility after a declaration
change, without an installation write.

### Concurrency evidence and limits

PUTs containing `X-Contentful-Version: 0` and subsequently
`X-Contentful-Version: 2147483647` both succeeded and replaced the stored Secret,
verified through App Identity. No returned `sys.version` was available. These probes
did not establish a usable version precondition for this endpoint.

A fresh GET can reduce the age of a preservation placeholder but cannot make a
GET/PUT sequence atomic. If an external rotation changes secret length between those
requests, replaying the old mask can store literal asterisks. If the rotation keeps
the same length, the PAT read cannot distinguish the two real secrets. No local hash,
sensitivity marker, or equality algorithm can recover that missing information.

### Free-form parameters and deletion

A follow-up PAT probe on 2026-10-21 used a new disposable definition and
installation, with synthetic values. With installation declarations absent,
creation, GET, and update preserved an object containing a string, number,
Boolean, null, array, and nested object. Setting the declaration list to `[]`
also accepted that object. These cases returned 200.

DELETE returned 404 before installation, 204 for an existing installation, and
404 on repetition; GET after deletion returned 404. The probe also repeated
plaintext Secret creation and masked GET/update with `a🔑é`, and verified that a
number supplied for a `Symbol` parameter returned 422 without changing the
installation. The disposable installation and definition were deleted and their
absence verified.

The [shared test server](../../../internal/contentful-management-go/testing/app_installation.go)
models these PAT parameter and deletion rules, with
[HTTP conformance tests](../../../internal/contentful-management-go/testing/handler_app_installation_test.go)
using explicit expected values. Its scope excludes App Identity authorization,
complete response metadata, defaults, and validation of every declaration type.

### Remaining boundaries

The experiments did not exercise App Events or Functions, restricted-role tokens,
shared or Marketplace app metadata permissions, UI behavior, every validation type,
Secret defaults, all Unicode sequences, network-failure recovery, or actual
simultaneous writers. The stale-mask race follows from the separately verified mask
replacement behavior and absence of an established atomic precondition; it was not a
concurrent load test. No claim is made that PAT access can detect same-length secret
rotation. Only app-authenticated reads verified exact stored values.

## Provider reproduction

See the [provider diagnosis and solution comparison](../../design/app-installation-secret-parameters.md)
for a pinned Terraform reproduction and the unresolved design choices. It does not
define an adopted reconciliation contract.

[parameters]: https://www.contentful.com/developers/docs/extensibility/app-framework/app-parameters/#secret-installation-parameters
[installation]: https://www.contentful.com/developers/docs/extensibility/app-framework/app-installation/
[put]: https://www.contentful.com/developers/docs/references/content-management-api/app-installations/install-or-update-an-app/
[identity]: https://www.contentful.com/developers/docs/extensibility/app-framework/app-identity/
[sdk-adapter]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/app-installation.ts
[sdk-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-installation.ts
[sdk-parameters]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/widget-parameters.ts
