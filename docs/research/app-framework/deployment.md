# App bundles and Function deployment

AppUpload supplies temporary zip bytes, and AppBundle stores a deployment artifact.
Updating the AppDefinition selects a bundle and activates its frontend or Functions.

See the [shared study scope and sources](README.md#scope-and-evidence) and [API
inventory](README.md#api-inventory). Upload and bundle observations establish
configuration behavior; activation and Function execution were not exercised.
[App configuration](configuration.md) describes the definition and its installations.

## AppUpload and AppBundle

AppUpload accepts zip bytes with `Content-Type: application/octet-stream` on the upload
host. Its `sys.expiresAt` is authoritative; the reference describes a temporary lifetime
of roughly 24–48 hours. Uploading is distinct from creating a bundle and selecting that
bundle on an app definition. Its SDK response contains only `sys`: upload ID,
organization link, expiry and creation/update metadata, without a version or archive
bytes. [Uploads][uploads], [upload entity][upload-entity].

The hosting guide limits an upload to 10 MB and 500 files. Those limits come from the
hosting guide, not the generic asset limits or a bundle-count quota. Whether the size
bound concerns compressed or expanded data is not established. [Hosting an
app][hosting].

AppBundle creation takes `upload: Link<AppUpload>`, optional `comment`, and Function
manifests; the SDK additionally passes an older `actions` manifest field with executable
paths. That older shape is not the endpoint/Function-link AppAction contract; current
CLI action upsert is a separate operation. The raw request uses `upload: {sys: {type:
"Link", linkType: "AppUpload", id: "upload-id"}}`; `appUploadId` is an SDK helper
argument, not the corresponding JSON wire member. Bundles have generated identity and
file metadata, no update operation. Frontend assets need root `index.html` and relative
asset paths; functions-only bundles do not need a frontend. [Bundle reference][bundles],
[adapter][bundle-sdk], [manifest types][bundle-entity].

The response exposes `files: [{name, size, md5}]`, optional comment/functions, and
`sys`; it does not promise an echoed upload link or legacy actions. GET does not
reconstruct zip/source bytes. An upload expires independently of the durable bundle, so
backup and reconstruction must distinguish recoverable metadata from original artifact
inputs. Upstream tests reuse one upload for multiple bundles; do not assume single-use
consumption. [Bundle types][bundle-entity], [integration tests][bundle-tests].

Upload POST/GET returned 201/200, and ordinary bundle creation returned 201. These
configuration results do not establish activation or runtime behavior. Function
deployment and execution contracts in this reference are sourced from public
documentation and tooling; they were not validated end to end.

## Function manifests and activation

Function manifests identify `id`, `name`, `description`, `path`, `accepts`, and optional
outbound `allowNetworks`. Functions are managed through the bundle deployment workflow;
removing one from the manifest and deploying removes it, and the guide warns that no
Function history/backup is retained. A UI zip upload alone does not create Functions.
[Working with Functions][working-functions].

Raw SDK manifests require `id`, `name`, `description`, and `path`, and treat `accepts`
and `allowNetworks` as optional; the build tool imposes additional requirements. Source
manifests also contain build-only `entryFile`, which is removed before upload. CLI
defaults and network normalization are not raw API defaults. [Bundle
types][bundle-entity], [manifest conversion][manifest-conversion].

Activation is an AppDefinition update after upload and bundle creation. The CLI reads
the definition, assigns its bundle link, removes `src` when activating a frontend, and
submits the definition update. It handles a separate 400 `Function upload failed` error
at this stage. Bundle creation success therefore does not prove Function deployment
success. The CLI also excludes the selected bundle from cleanup; that policy is not
evidence of a server rejection for selected-bundle DELETE. [Activation
implementation][activation], [cleanup implementation][bundle-cleanup].

## Function discovery

The pinned SDK declares Function fields `name`, `description`, `path`, `accepts`, and
optional `allowNetworks`. Its `sys` has the Function ID and organization/app links,
timestamps and actor links, with no version. The installation discovery adapter uses
that same entity type; this declaration does not establish identical field presence in
every response scope. [Function entity][function-entity], [adapter][function-sdk].

The specialized Function list query declares `accepts[all]`, not skip/limit, although
its response uses the generic collection type. The CMA list example omits skip/limit and
shows `accepts`/`allowNetworks` as strings, including `appaction.event`; the SDK
declares arrays and the toolkit uses `appaction.call`. These example/type differences do
not establish a validated wire union or alternative invocation name. [Function
adapter][function-sdk], [Function example][function-list], [invocation
types][function-types].

The [supplied read-only study](README.md#source-metadata) found a separate installation
discovery projection at `GET I/functions`: `{sys, items, total}`, without `skip` or
`limit`. Nonempty items contained `sys`, `name`, `description`, and array-valued `accepts`,
but omitted `path` and `allowNetworks`. Their metadata linked the app definition and
organization and omitted `sys.version`. These app-owned references do not describe an
independently configured environment Function. The discovery projection is insufficient
to reconstruct a complete Function manifest.

For each sampled nonempty collection, `accepts[all]=appaction.call` returned exactly its
matching subset. This supports that single-value filter in installation discovery;
multi-value filtering and server defaults remain untested. No definition-scoped
Function read, deployment, or execution was performed by this experiment.

## Unresolved behavior

Frontend bundle promotion/rollback does not establish reliable historical Function
restoration. The available evidence does not establish selected-bundle deletion
behavior, old-bundle Function rollback, the effects of omitted versus empty Function
manifests, activation failure atomicity, or effects on existing Function references. A
selected bundle link alone does not establish deployment health or repair of references
to removed Functions.

No authoritative AppBundle count quota was established.

[bundle-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/app-bundle.ts
[bundle-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-bundle.ts
[bundle-tests]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/test/integration/app-bundle-integration.test.ts
[activation]: https://github.com/contentful/create-contentful-app/blob/909e37a3e55a1e5851bdc35f49ac9c5c34b64d4e/packages/contentful--app-scripts/src/activate/activate-bundle.ts
[bundle-cleanup]: https://github.com/contentful/create-contentful-app/blob/909e37a3e55a1e5851bdc35f49ac9c5c34b64d4e/packages/contentful--app-scripts/src/clean-up/clean-up-bundles.ts#L113-L117
[manifest-conversion]: https://github.com/contentful/create-contentful-app/blob/909e37a3e55a1e5851bdc35f49ac9c5c34b64d4e/packages/contentful--app-scripts/src/utils.ts#L118-L174
[function-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/function.ts
[working-functions]: https://www.contentful.com/developers/docs/extensibility/app-framework/working-with-functions/
[function-types]: https://github.com/contentful/node-apps-toolkit/blob/64fa31b6b2223cd1c8b1798fa540e8aad5e2d319/src/requests/typings/function.ts
[uploads]: https://www.contentful.com/developers/docs/references/content-management-api/app-uploads/
[hosting]: https://www.contentful.com/developers/docs/extensibility/app-framework/hosting-an-app/
[bundles]: https://www.contentful.com/developers/docs/references/content-management-api/app-bundles/
[function-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/function.ts
[function-list]: https://www.contentful.com/developers/docs/references/content-management-api/functions/get-all-functions/
[upload-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-upload.ts
