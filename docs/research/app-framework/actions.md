# App Actions: definitions, invocation, and outcomes

An AppAction defines a callable capability; AppActionCall records one execution.
Configuration, submission, and successful execution are separate API outcomes.

See the [shared study scope and sources](README.md#scope-and-evidence) and [API
inventory](README.md#api-inventory). The observations below cover action configuration;
invocation, execution, and result validation were not exercised.

## AppAction definitions

There is no separate publish/activate lifecycle in the reviewed AppAction API. The
endpoint form used in the live probe was:

```json
{
  "name": "Example HTTP action",
  "category": "Custom",
  "parameters": [],
  "type": "endpoint",
  "url": "https://example.invalid/action-probe"
}
```

Current action forms are `endpoint` with HTTPS URL and `function-invocation` with a
`function: Link<Function>` accepting `appaction.call`; legacy `type: function` is
deprecated. The SDK create shape permits omitted `type` for an endpoint and returns
explicit `type: "endpoint"`; the default was also observed directly. Built-in categories
include `Entries.v1.0` and `Notification.v1.0`. The live category endpoint returned
those two built-ins; their inputs describe entry IDs and notification message/recipient
respectively. [Action types][action-entity], [categories][categories].

The AppAction response carries the definition data and `sys.id`, plus
organization/app-definition links and timestamps/actors, without a `version` in the SDK.
AppActionCategory instead has `sys.id`, `sys.type`, a string-valued `sys.version`,
`name`, `description`, and optional parameter definitions. That category version is not
an AppAction concurrency token. [Action and category entities][action-entity].

The current action guide supports `parametersSchema` and `resultSchema` using JSON
Schema draft 4. These validate invocation inputs and successful structured results. The
input contract is an alternative, not two independent optional fields: direct
observations accept a custom action with `parameters` or `parametersSchema`, but reject
both together. The pinned SDK's custom-category intersection still requires `parameters`
while allowing `parametersSchema`; that type is not an accurate exclusive union for the
observed service. [App Actions guide][actions], [SDK type][action-entity].

Legacy parameter definitions require `id`, `name`, and `type`; the pinned source types
are `Boolean`, `Symbol`, `Number`, and `Enum`. Optional fields include `description`,
`required`, `default`, and `options`. AppAction parameters omit the shared parameter
type's `labels`; `Secret` belongs to installation parameters, not this AppAction type.
Type declarations do not establish null acceptance or server defaults. [Shared
parameter types][parameter-types].

The action adapter passes request fields directly, so construct a writable body rather
than sending response `sys`. Function actions use a raw `function` link; the CLI's
`functionId` is a manifest convenience converted before the request. The SDK only
declares caller-supplied `id` in its Function creation branch, while the CLI carries it
in both branches; chosen-ID support for every action form is therefore not established.
[Action adapter][action-sdk], [CLI conversion][action-conversion].

The schema-based alternative exercised successfully was:

```json
{
  "name": "Schema action",
  "category": "Custom",
  "type": "endpoint",
  "url": "https://example.invalid/actions",
  "parametersSchema": {
    "type": "object",
    "properties": {"message": {"type": "string"}},
    "required": ["message"],
    "additionalProperties": false
  },
  "resultSchema": {"type": "object"}
}
```

| Observed request | Result and read-back evidence |
| --- | --- |
| Custom with `parameters: []`, no schema | POST 201; empty array retained. |
| Custom with only `parametersSchema` | POST 201; schema retained without a parameters array. |
| Both `parameters: []` and `parametersSchema` | 422 `ValidationFailed`: `AppAction cannot have both parametersSchema and parameters. Please provide just a parametersSchema.` |
| Custom with neither input definition | 422 `ValidationFailed`: `Please provide a parametersSchema to validate the app action call parameters`. |
| Schema-based action PUT omitting existing description/resultSchema | 200; both absent from the response and subsequent GET. |
| PUT switching schema-based input to `parameters: []` | 200; parametersSchema absent on subsequent GET. |
| Null description, URL, or either schema | 422 type validation errors; null was not an omission/clear operation. |
| `Entries.v1.0` with no parameters array | POST 201; response supplied its `entryIds` parameter definition. |

Schema object key order changed on GET; compare JSON structurally, not as raw text. The
observations do not cover every JSON Schema keyword, built-in/schema combination, or
result validation during execution. The observed `Entries.v1.0` response supplied
parameter definitions that were absent from the request.

The observed AppAction errors for missing or conflicting input definitions use a string
at `details.errors`, while null-field type errors use an array of validation objects.

Live endpoint-action create returned 201 both without and with a signing secret. Update
without a version header and GET returned 200; DELETE returned
204. This proves configuration access, not invocation behavior. The docs'
signing-secret requirement matters to execution: the call reference describes 409 for an
absent or invalid signing secret for the app providing the action, 403 for invalid
calling-app access, and 404 for an absent action or installation of the app providing
it. Older overview wording restricts callers to app identities, while the current guide
describes manual user triggering; no caller-permission matrix was exercised here. [Call
reference][calls], [CMA App Actions][action-reference].

## AppActionCall request and outcome

An invocation sends
`POST /spaces/{space_id}/environments/{environment_id}/app_installations/{app_definition_id}/actions/{action_id}/calls`
with a `parameters` object. These are argument values, distinct from an AppAction's
parameter definitions or input schema.
For the schema-based action above, a synthetic request is:

```json
{"parameters": {"message": "Example"}}
```

The trigger reference documents 201 on submission; successful submission does not
establish successful execution. [Trigger endpoint][call-trigger].

The SDK call's `sys` contains its ID and links named `appDefinition`, `action`, `space`,
and `environment`, with no resource `version`. Structured outcome is also inside `sys`:

| `sys.status` | Outcome data |
| --- | --- |
| `processing` | No terminal result is declared. |
| `succeeded` | `sys.result` is a JSON value: scalar, array, object, or null. |
| `failed` | `sys.error` has `sys: {type: "Error", id: string}` and `message`, with optional `details` and `statusCode`. |

A successful GET of a failed call is distinct from failure of the GET request. Likewise,
a null successful result is not an absent terminal outcome. [Call entity][call-entity].

The `/response` route returns a `response` envelope with string `body` and optional
status/header data in the SDK. The SDK also declares an `AppActionCallResponse` identity
with links; the CMA example omits `sys` and adds `method`/`url`. These are differing
source representations, not a complete wire schema validated here. Structured
`sys.result`, raw `response.body`, and legacy webhook-shaped call details are separate
representations. [Raw response entity][call-entity], [raw response reference][call-raw].

SDK `createWithResult` polls the installation-scoped route; `createWithResponse` uses
older call details. The default two-second interval and 15 checks are SDK polling
policy, not a service execution deadline. [Polling implementation][call-sdk].

No call cancellation, deletion, replay, list endpoint, or retention guarantee was
established in the inspected AppActionCall contract. An execution ID should not be
treated as a configurable persistent action definition.

## Unresolved behavior

Every JSON Schema keyword and built-in/schema combination, chosen-ID support across
action forms, execution/result validation, caller permissions, and the complete
raw-response wire shape. No call retention, cancellation, replay, or list contract was
established.

No authoritative AppAction count quota was established.

[action-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/app-action.ts
[action-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-action.ts
[parameter-types]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/widget-parameters.ts
[action-conversion]: https://github.com/contentful/create-contentful-app/blob/909e37a3e55a1e5851bdc35f49ac9c5c34b64d4e/packages/contentful--app-scripts/src/upsert-actions/make-cma-payload.ts
[categories]: https://www.contentful.com/developers/docs/references/content-management-api/app-action-categories/get-app-action-categories/
[call-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/app-action-call.ts
[call-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-action-call.ts
[actions]: https://www.contentful.com/developers/docs/extensibility/app-framework/app-actions/
[calls]: https://www.contentful.com/developers/docs/references/content-management-api/app-action-calls/
[action-reference]: https://www.contentful.com/developers/docs/references/content-management-api/app-actions/
[call-trigger]: https://www.contentful.com/developers/docs/references/content-management-api/app-action-calls/trigger-an-action/
[call-raw]: https://www.contentful.com/developers/docs/references/content-management-api/app-action-calls/get-raw-response/
