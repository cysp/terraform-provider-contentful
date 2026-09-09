# Environment creation and readiness

Creating an environment starts an asynchronous operation. A successful creation response
is separate from readiness to manage content in that environment. Contentful instructs
callers to query the environment after creation.

## Published contract

The [CMA environment
reference](https://www.contentful.com/developers/docs/references/content-management-api/environments/#environment-states)
documents four environment states:

| Status | Meaning |
| --- | --- |
| `queued` | Creation is waiting to start. |
| `inProgress` | Creation is in progress. |
| `ready` | The environment is ready. |
| `failed` | Creation failed. |

The first-party [environment
model](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/entities/environment.ts#L9-L25)
models the property as a `sys.status` link, while the published reference names it
`sys.state`. The cited sources therefore disagree on the property name. Read the addressed
environment at `/spaces/{space_id}/environments/{environment_id}` to observe its status.

## Interpretation and limits

The published states establish a distinction between creation acceptance, ongoing work,
readiness, and failure. They do not establish a fixed completion time or a polling
interval. No direct timing distribution, regional comparison, or undocumented status
behavior is retained in this reference. Client polling and test-fixture timing are
separate policies.
