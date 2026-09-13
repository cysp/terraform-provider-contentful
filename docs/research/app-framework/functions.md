# Functions: runtime, logs, usage, and availability

Functions are deployed app capabilities with execution, network, and resource limits.
Function logs and aggregate usage describe runtime activity; their retention and
pagination contracts differ.

See the [shared study scope and sources](README.md#scope-and-evidence) and [API
inventory](README.md#api-inventory). The read-only observations below inspected existing
Function logs; the experiment did not generate those executions.

## Runtime and availability

Functions are documented as Premium/Partner functionality. Premium includes 5 million
executions/month and Enterprise 10 million/month in the usage table. Marketplace
Function executions are excluded from Function quota/overage accounting; that exception
does not enable custom Function deployment on Free. [Function
availability][working-functions], [usage limits][usage-limits].

The reviewed [technical limits][limits] distinguish service ceilings from the executions
included in a plan:

| Limit | Published value |
| --- | --- |
| Functions per app | 50 |
| Executions per organization per month | 20 million |
| Memory per execution | 128 MB |
| Outbound requests per execution | 20 |
| Wall time per execution | 30 seconds; 10 seconds for `resource.*` |
| CPU time per execution | 10 seconds; 1 second for `graphql.*` |
| Log retention | 30 days |

The limits page spells the wall-time exception `resource.*`, while the invocation names
use `resources.*`. That source discrepancy remains unresolved; the table preserves the
published spelling.

The runtime is not complete Node.js and has no filesystem. Event payloads and the
built-in CMA client's request/response data must be below 32 MB. Resource limit or
timeout termination is documented as not retried. This retry statement concerns
Function termination; external HTTP App Event retries remain unverified. [Runtime
contract][functions].

## Function invocation context and logs

A Function handler receives `(event, context)`. For App Events, `event` contains `type`,
`headers`, and the topic-dependent `body`. App Actions use `type: "appaction.call"` with
headers and a parameter body; this is different from the CMA invocation request's
`parameters` wrapper. Context includes `spaceId`, `environmentId`, and
`appInstallationParameters`, with optional `cmaClientOptions`, `cma`, and
`originalRequest.headers` in the toolkit. [Function types][function-types].

The Function guide supplies CMA client options for management contexts such as App
Events, App Actions, and external-resource search/lookup. It does not promise CMA
credentials in GraphQL delivery contexts. The supplied identity is scoped to the
containing app's triggering space/environment; it does not authorize cross-space
mutation. [Function CMA access][functions].

### Function logs

FunctionLog list/detail requests send `x-contentful-enable-alpha-feature: function-logs`
in the pinned SDK. List parameters include `limit`, `pageNext` or `pagePrev`, and
`sys.createdAt[gt]`, `[gte]`, `[lt]`, or `[lte]`. The SDK's cursor and interval unions
reject both cursor directions or competing bounds. Its type composition also inherits
`accepts[all]`; the dedicated FunctionLog reference does not establish that filter's
server behavior. [Log adapter][log-sdk], [query types][common-types].

The SDK record exposes `requestId`, event data, severity counters (`info`, `warn`,
`error`), and messages containing timestamp, type (`INFO`, `WARN`, `ERROR`), and text.
`sys` identifies the log and links its space/environment/app definition. The public CMA
example instead separates `eventType` from `event: {headers, body}`, and represents
message timestamps and severity counts as strings where the SDK declares numbers. The
SDK's event type is GraphQL-shaped. Neither source establishes every log variant. The
public list example is only `{}`; it does not establish a complete collection envelope.
[Log entity][log-entity], [log detail][log-detail], [log list][log-list].

Observations from the [supplied read-only study](README.md#source-metadata) distinguish
list summaries from individual log details:

| Read | Observed representation |
| --- | --- |
| List, including empty lists and requests with `limit=1` | `{sys, pages, items}`, without `total`, `skip`, or `limit`. Items contained `sys`, numeric `severity.info`/`warn`/`error`, string `requestId`, and string `eventType`; `event` and `messages` were absent. |
| Individual detail | Added `event` and `messages`. Where messages existed, `timestamp` was a JSON number and `message`/`type` were strings; severity counts were numbers. |

Sampled `appevent.handler` and `appaction.call` event objects contained `headers`,
`body`, and `type`; the `resources.search` sample contained `type`, `resourceType`,
`query`, and `limit`. These shapes resolve numeric representation and summary/detail
presence only for the inspected records. They do not establish timestamp units, all
event variants, or GraphQL log shapes.

One returned next-page link was followed successfully. Reverse paging, filter behavior,
and stability under concurrent log creation were not tested. All requests included
`x-contentful-enable-alpha-feature: function-logs`, as the SDK and public endpoint
specify; the observations do not independently establish behavior without that header.

## Usage and observability

Aggregate Function usage requires inclusive `date[gte]` and `date[lte]` parameters. The
endpoint reference accepts `YYYY-MM-DD` or full ISO-8601 date-times. `P1D` granularity
supports up to 31 days per query; `P1M` supports up to 12 calendar months including the
current month. The documented history boundary is the last 12 months regardless of
granularity. A query without required dates returned 422; that
probe did not establish accepted date forms or history boundaries. Tenant usage results
are excluded from this reference. [Aggregated usage][usage-aggregate].

The Usage overview's Permissions paragraph limits Free-plan history to 45 days, while
its Aggregated usage section specifies the same 12-month boundary as the endpoint
reference. The overview and endpoint require an Organization Admin or Organization Owner
role. No older-history query resolved the Free-plan distinction. Legacy periodic usage
methods are deprecated with a 2027-02-28 sunset. [Usage overview][usage],
[usage types][usage-entity].

Enterprise Observability streams CDA, GraphQL, and audit logs to external storage. It
requires Enterprise and organization admin/owner access. Its export delivery status is
not AppEventSubscription delivery health. [Enterprise Observability][observability].

## Unresolved behavior

FunctionLog shapes beyond the sampled event types, timestamp units, reverse pagination,
filter behavior, and header necessity; Free-plan history beyond 45 days where usage
documentation disagrees. No tenant usage data is retained.

[log-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/function-log.ts
[log-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/function-log.ts
[functions]: https://www.contentful.com/developers/docs/extensibility/app-framework/functions/
[working-functions]: https://www.contentful.com/developers/docs/extensibility/app-framework/working-with-functions/
[function-types]: https://github.com/contentful/node-apps-toolkit/blob/64fa31b6b2223cd1c8b1798fa540e8aad5e2d319/src/requests/typings/function.ts
[limits]: https://www.contentful.com/developers/docs/platform/technical-limits/
[usage-limits]: https://www.contentful.com/help/admin/usage/usage-limit/
[usage-aggregate]: https://www.contentful.com/developers/docs/references/content-management-api/usage/get-usage-aggregated/
[usage]: https://www.contentful.com/developers/docs/references/content-management-api/usage/
[usage-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/usage.ts
[observability]: https://www.contentful.com/developers/docs/concepts/enterprise-observability/
[common-types]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/common-types.ts
[log-detail]: https://www.contentful.com/developers/docs/references/content-management-api/function-logs/get-a-function-log/
[log-list]: https://www.contentful.com/developers/docs/references/content-management-api/function-logs/get-all-function-logs/
