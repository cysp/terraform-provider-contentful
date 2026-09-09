# App Events: subscriptions, topics, and payloads

An AppEventSubscription routes selected events for an installed app to an HTTP target
or a Function pipeline. The subscription belongs to the app definition; installation
determines the triggering space and environment.

See the [shared study scope and sources](README.md#scope-and-evidence) and [API
inventory](README.md#api-inventory). The observations below cover subscription
configuration and topic validation; event delivery and payload fidelity were not
exercised.

## AppEventSubscription

The SDK request consists of `topics: string[]`, optional `targetUrl: string`, and
optional `functions`, whose `filter`, `transformation`, and `handler` members each
contain a Function link. The response `sys` has organization and app-definition links
but no subscription `id` or `version`. The SDK does not add a version header.
[Entity][event-entity], [adapter][event-sdk].

The ordinary HTTP form exercised successfully was:

```json
{
  "targetUrl": "https://example.invalid/event-probe",
  "topics": ["Entry.publish"]
}
```

An illustrative response projection follows. IDs are synthetic; timestamps and actor
links are omitted. This shape comes from the documented entity type, while the
no-subscription-ID/no-version behavior was also observed directly.

```json
{
  "sys": {
    "type": "AppEventSubscription",
    "organization": {
      "sys": {"type": "Link", "linkType": "Organization", "id": "organization-id"}
    },
    "appDefinition": {
      "sys": {"type": "Link", "linkType": "AppDefinition", "id": "app-definition-id"}
    }
  },
  "targetUrl": "https://example.invalid/event-probe",
  "topics": ["Entry.publish"]
}
```

The Function form is documented independently of account entitlement:

```json
{
  "topics": ["Entry.publish"],
  "functions": {
    "handler": {
      "sys": {"type": "Link", "linkType": "Function", "id": "handler-id"}
    }
  }
}
```

A handler replaces the external HTTP target. A filter decides whether an event
continues, while a transformation alters the outgoing request before signing. Their
invocation types are `appevent.filter`, `appevent.transformation`, and
`appevent.handler`. They return, respectively, `{result: boolean}`, a headers/body
object, and no result. Function context supplies triggering space/environment and
installation parameters; the provided CMA identity has that installation's scope.
[Functions][functions], [toolkit types][function-types].

## Observed wire behavior

These results describe the exact configuration requests exercised.

| Experiment | Result | Consequence |
| --- | --- | --- |
| Initial PUT / later PUT | 201 / 200 | Singleton upsert; submitted topics replaced prior topics. |
| PUT without version / with `X-Contentful-Version: 999999` | Both succeeded | No effective version check was demonstrated by this experiment. |
| Read response | No `sys.id` or `sys.version` | Identity comes from organization and app definition. |
| Successive replacements | Both `createdAt` and `updatedAt` changed | `createdAt` cannot be treated as stable original-creation evidence here. |
| Omitted `targetUrl` in HTTP form | 400, missing required property | Omission did not preserve the previous target. |
| Null / empty `targetUrl` | 422 type / regex error | Empty is not a usable target; validation reported `^https://.+`. |
| Omitted / null / empty `topics` | 422 | Required array, minimum one item. |
| Repeated topic | 422 uniqueness error | Duplicate topic strings are rejected. |
| `Entry.*` | 422 | Webhook wildcard syntax cannot be assumed for App Events. |
| Extra `filters: []`, `transformation: {}`, `unexpected: true` | 200, all three omitted in response | Acceptance does not establish support; no execution test checked their effect. |
| Nine Comment/Workflow/Task topics together | 200; identical topics on GET | Those exact arrays were accepted in the probe context. |
| DELETE / repeated DELETE | 204 / 404 | Deletion is not status-idempotent. |
| Delete definition with a subscription, then GET child | 204 / 404 | Child became inaccessible because the parent was absent; this does not inspect internal deletion storage. |

HTTP App Events do not expose the WebhookDefinition filter/transformation DSL, custom
method, multiple targets, basic-auth configuration, or a pause flag in the reviewed
request contract. The tested unknown-field behavior is particularly important: ignoring
a field is not implementation of that field.

## Event topics

The observed validation response to an invalid topic enumerated **88 allowed strings**.
Each action below is prefixed with its entity, for example `Entry.publish`. This is an
observed server allowlist, not a guarantee that the organization is entitled to every
underlying feature or that every event was generated and delivered.

| Entity | Actions |
| --- | --- |
| Entry | create, delete, save, auto_save, publish, unpublish, archive, unarchive |
| Asset | create, delete, save, auto_save, publish, unpublish, archive, unarchive |
| ContentType | create, delete, save, publish, unpublish |
| AppInstallation | create, delete, save |
| Task | create, save, delete |
| Comment | create, delete, save |
| Release | create, save, archive, unarchive, delete |
| ReleaseAction | create, execute |
| ReleaseAsset | save, auto_save |
| ReleaseEntry | save, auto_save |
| ScheduledAction | create, save, execute, delete |
| BulkAction | create, execute |
| TemplateInstallation | complete |
| Workflow | create, save, complete |
| Experience | create, save, publish, unpublish, delete |
| Template | create, save, publish, unpublish, delete |
| ComponentType | create, save, publish, unpublish, delete |
| DataAssembly | create, save, publish, unpublish, delete |
| DesignToken | create, save, publish, unpublish, delete |
| View | create, save, publish, unpublish, delete |
| Fragment | create, save, publish, unpublish, archive, unarchive, delete |

The [CMA subscription reference][event-reference] lists fewer topics and calls the
installation update event `AppInstallation.update`; the observed allowlist says
`AppInstallation.save`. Official changelogs separately document Workflow
create/save/complete and Comment create/delete, followed by Comment.save.
[Workflow/comment announcement][workflow-comments], [Comment.save][comment-save].

The toolkit's payload map is also incomplete: it lacks Comment.save and models
Workflow.delete where the observed allowlist and changelog use Workflow.complete. The
map does not define a complete server event enum, and its entry-field typing is not a
universal wire model of Contentful field values. [Payload map][payload-types].

## Event data and contextual headers

Subscription topics use `Entity.action`; the delivered `X-Contentful-Topic` header uses
`ContentManagement.Entity.action`. A topic identifies both an entity and an operation;
its body is not uniformly a complete CMA entity. [Function event types][function-types].

The pinned toolkit models Entry/Asset/ContentType delete and unpublish payloads as
`sys`-only records, with different metadata/content on create/save/publish. It models
Comment create/delete through `sys.newComment`/`sys.oldComment`, and Task save through
both `sys.oldTask` and `sys.newTask`. These distinctions explain why one generic entity
decoder is insufficient. They remain SDK declarations, subject to the payload-map
omissions and narrow field typing described above; no delivery experiment establishes
them as exhaustive wire schemas. [Payload declarations][payload-types].

The published contextual headers for entry and asset publish/unpublish events are:

| Header | Meaning |
| --- | --- |
| `x-contentful-bulk-action-id` | Triggering BulkAction `sys.id`. |
| `x-contentful-scheduled-action-id` | Triggering ScheduledAction `sys.id`. |
| `x-contentful-release-id` | Release `sys.id`. |
| `x-contentful-release-version-id` | Release `sys.version` used for the operation. |
| `x-contentful-release-action-id` | ReleaseAction `sys.id`. |

These are conditional context, not fields guaranteed on every event. The announcement
covers both App Events and webhooks. [Contextual headers][context-headers].

## Unresolved behavior

Handler/target coexistence, filter-only target requirements, invalid Function IDs or
invocation-role mismatches, and removal of individual Function roles. End-to-end payload
fidelity, delivery order, retries, duplicate identifiers, and pending events after edits
remain unverified.

No app-specific HTTP event call, log, health, or replay endpoint was identified in the
reviewed public surface. This does not establish the absence of internal records.

[event-reference]: https://www.contentful.com/developers/docs/references/content-management-api/app-event-subscriptions/
[event-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/app-event-subscription.ts
[event-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-event-subscription.ts
[functions]: https://www.contentful.com/developers/docs/extensibility/app-framework/functions/
[function-types]: https://github.com/contentful/node-apps-toolkit/blob/64fa31b6b2223cd1c8b1798fa540e8aad5e2d319/src/requests/typings/function.ts
[payload-types]: https://github.com/contentful/node-apps-toolkit/blob/64fa31b6b2223cd1c8b1798fa540e8aad5e2d319/src/requests/typings/event-payloads.ts
[workflow-comments]: https://www.contentful.com/developers/changelog/workflow-and-comment-events-updates/
[comment-save]: https://www.contentful.com/developers/changelog/new-webhook-event-comment-save/
[context-headers]: https://www.contentful.com/developers/changelog/new-contextual-appevent-and-webhook-headers-on-content-events/
