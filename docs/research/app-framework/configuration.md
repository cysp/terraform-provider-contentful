# App configuration and installation

AppDefinition holds shared app configuration; AppInstallation places that app in an
environment. AppDetails supplies presentation metadata, and AppAccessGrant controls
which organizations may install the app.

See the [shared study scope and sources](README.md#scope-and-evidence) and [API
inventory](README.md#api-inventory). The configuration probes used uninstalled apps;
the separately scoped installation reads and parameter experiments below supplement
the published installation contract. Access-grant behavior comes from cited sources.

## AppDefinition and AppInstallation data

The table describes pinned SDK shapes using the [shared Link
notation](../README.md#terminology-and-scope); optional declarations alone do not establish null
acceptance, clearing, or server defaults.

| Resource | Writable data | Returned identity and relationships |
| --- | --- | --- |
| AppDefinition | Required `name`; optional `src`, `locations`, `parameters`. Update can select `bundle: Link<AppBundle>`; the SDK create type excludes `bundle` and `sys`. | `sys.id`, organization link, `sys.shared`, and version/timestamp/actor metadata in the SDK. [Entity][definition-entity], [adapter][definition-sdk]. |
| AppInstallation | Optional `parameters` on PUT. These are installation values, distinct from the declarations on AppDefinition. | Definition, space, environment, and organization links; the SDK omits `sys.id` but retains version metadata. Addressing uses the app definition ID. [Entity][installation-entity]. |

The Free plan's published limits are 10 app definitions per organization and 10
installations per environment. The older Extensibility FAQ's undifferentiated 250/50
limits should not override the explicit current Free rows. [Technical limits][limits].

AppDefinition locations include `app-config`, `entry-sidebar`, `entry-editor`,
`entry-field`, `dialog`, `page`, `home`, and `experience-toolbar` in the pinned SDK.
`entry-field` has `fieldTypes`; `page` can have `navigationItem: {name, path}`. The
`agent` location and `agent: {id}` property are explicitly internal-only in that source,
not established public app capabilities. Instance and installation parameter
declarations use `id`, `name`, and `type`, with optional `description`, `required`,
`default`, `options`, and `labels`. Installation declarations additionally support
`Secret`. [Definition entity][definition-entity], [parameter types][parameter-types].

`src` points to an external frontend; `bundle` selects Contentful assets. The definition
reference permits both when the bundle contains only Functions. Definition changes
propagate to installations. The SDK's create/update type difference does not
independently prove whether POST can select a bundle. [App definitions][definitions],
[definition entity][definition-entity].

Installation parameter size has conflicting source descriptions: the CMA AppInstallation
reference describes an object limited to **16 kB** after stringification; the App Parameters guide
and pinned SDK comment say **32 kB**. The exact byte boundary and counting rules remain
unverified. The guide describes undeclared parameters as a free-form object, while the
SDK type also permits arrays and scalars; the observations below establish rejection of
those shapes only for the tested app with declarations. Cloning an environment is
documented to copy installations and parameters. [App installations][installations],
[App Parameters][app-parameters], [installation entity][installation-entity],
[free-form type][parameter-types].

The App Parameters guide documents redaction of values whose keys match `Secret`
installation declarations when read through a personal access token in CMA or the App
SDK in the Contentful UI. App Identities, App Events, and Functions can receive the raw
values. Such a management read therefore cannot reconstruct the original secret for a
backup or subsequent write. The redaction marker and Secret omission/replacement
behavior were not exercised by the parameter experiment below. [Secret installation
parameters][secret-parameters].

Cross-environment discovery has a special SDK response: `{sys: {type: "Array"}, items:
AppInstallation[], includes: {Environment: Environment[]}}`. It does not declare the
ordinary `total`/`skip`/`limit` fields. The adapter sends `sys.organization.sys.id[in]`
as a query parameter; the generated CMA example places it in a header. This is a source
discrepancy, not independent evidence for two supported encodings. [Discovery
type][definition-entity], [adapter][installation-sdk], [endpoint
example][installation-org].

Marketplace installation terms acceptance uses `X-Contentful-Marketplace` with
`i-accept-end-user-license-agreement,i-accept-marketplace-terms-of-service,i-accept-privacy-policy`
when the SDK's `acceptAllTerms` is true. Looking up a definition does not supply that
installation header. [Installation adapter][installation-sdk].

### Observed installation reads

The [supplied read-only study](README.md#source-metadata) covered environment-scoped
collection and individual installation GETs. Both returned `sys.id` equal to the app
definition ID and omitted `sys.version`, unlike the SDK declaration in the table above.
Individual responses also contained actor links. Some reads omitted `parameters` and
others returned objects; neither the SDK's inherited version nor parameter presence is
a required wire field established by these samples.

The collection included `includes.AppDefinition` and `includes.ResolvedAppDefinition`.
The resolved projection exposed `src` and `sys.expiresAt`; its expiry behavior and a
separately manageable resource were not established. Included definitions lacked the
full definition's version metadata, so the includes do not reconstruct the writable
organization-owned definition. A request with `limit=1&skip=1` returned the corresponding
offset envelope and one item; it did not establish every filter or pagination default.

Installation metadata read through an environment alias retained that alias in
`sys.environment`. That link did not identify the concrete target independently of the
request route. The alias target agreed at three checkpoints, which were separate
snapshots rather than a guarantee against intervening retargeting. See [environment
alias routing](../environment-aliases.md).

### Installation parameter replacement

The [supplied installation-parameter study](README.md#source-metadata) used one existing
installation addressed through a concrete environment. Its AppDefinition declared only
`Symbol` installation parameters, including one required parameter and no `Secret`
parameters. Every comparison began with the same nonempty parameter object and used a
GET after the PUT; accepted changes were restored before the next comparison.

| Submitted PUT body | Outcome | Following GET |
| --- | --- | --- |
| `{}` | 200 | `parameters` absent; the previous object was removed. |
| `{"parameters":null}` | 422, expected object | Original object unchanged. |
| `{"parameters":[]}` | 422, expected object | Original object unchanged. |
| `parameters` set to a string, Boolean, or number | 422, expected object for each request | Original object unchanged. |
| `{"parameters":{}}` | 422, required parameter missing | Original object unchanged. |
| Object omitting one existing optional key, retaining the required key and other values | 200 | Omitted key absent; remaining values preserved. |

Validation errors used an array at `details.errors`. Object-type errors had
`name: "type"` and `path: ["parameters"]`; the missing-required error had
`name: "required"` and a path to the declared parameter, illustrated synthetically as
`["parameters", "requiredParameter"]`.

Accepted PUT responses and subsequent GETs agreed on parameter presence and value. In
this installation, whole-field omission cleared the object, and a supplied object
replaced the map rather than merging its keys. Keeping the required key in the partial
object separated replacement behavior from required-field validation.

Absence and an empty object were observably different: omission succeeded even with a
required declaration, while a present empty object failed that declaration's
validation. This is not evidence that every app rejects an empty object or accepts
omission under every parameter declaration. Undeclared parameters, other declaration
types, Secret values, accepted-object coercions, size boundaries, and concurrency
guarantees remain untested.

## AppDetails

AppDetails is a separate optional singleton. Its current `icon` shape is `{type:
"base64", value: "data:image/png;base64,..."}`; the reference describes PNG/JPG icons up
to 128 by 128 pixels. An empty `{}` PUT succeeded with 201, and deletion returned 204.
The data-URI form is an upstream fixture, not proof that bare base64 is rejected; icon
clearing and normalization were not probed. The SDK response adds
organization/app-definition links and timestamps/actors under `sys`, with no singleton
id/version. [Details reference][details], [entity][details-entity], [upstream
integration fixtures][details-tests].

## AppAccessGrant and sharing

By default, an AppDefinition can be installed only in spaces belonging to its owning
organization; a suitable AppAccessGrant permits installation by another organization.
It is app distribution authorization, not a content-management token. [Access Grant
concept][grant-concept].

Public sharing is documented with
`granteeType: "all"` and `granteeId: "all"`, returning `sys.type: "AppAccessGrant"`. No
grants were created during this research. [Create reference][grant-create].

The public create reference exposes a permissive map and an all/all example; the
inspected sources do not establish the exact targeted-organization enum or what
revocation does to existing installations. Those details limit what this evidence
establishes about targeted grants and revocation. There is no AccessGrant entity/adapter
in the pinned management SDK; the public CMA reference is the source for its endpoint
family.

The published technical limit is 1,000 access grants per app definition. A collection
example's pagination limit must not be confused with that quota. [Technical
limits][limits].

`AppDefinition.sys.shared` is response metadata, not a writable definition property in
the SDK. Its exact derivation from grants was not established. Sharing by a grant/link
is separate from Marketplace submission and review; installing a shared app links the
existing definition rather than transferring ownership or creating an owned copy.
[Definition
entity](https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-definition.ts),
[Marketplace
publication](https://www.contentful.com/developers/docs/extensibility/app-framework/publishing-an-app/),
[sharing
announcement](https://www.contentful.com/developers/changelog/app-sharing-easy-app-installation-outside-the-contentful-marketplace/).

## Unresolved behavior

The evidence does not establish:

- Clearing and defaults for other optional definition fields, or bundle selection during
  definition creation.
- The effects of changing parameter declarations on existing installations, the exact
  parameter size limit, behavior beyond the tested declarations, or Secret replacement.
- Complete permission and concurrency rules.
- The effects of grant revocation on existing installations, the exact targeted-grant
  enum, or how Contentful derives shared metadata.

[definition-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/app-definition.ts
[installation-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/app-installation.ts
[parameter-types]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/widget-parameters.ts
[details-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-details.ts
[details-tests]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/test/integration/app-details-integration.test.ts
[definitions]: https://www.contentful.com/developers/docs/references/content-management-api/app-definitions/
[installations]: https://www.contentful.com/developers/docs/references/content-management-api/app-installations/
[app-parameters]: https://www.contentful.com/developers/docs/extensibility/app-framework/app-parameters/
[secret-parameters]: https://www.contentful.com/developers/docs/extensibility/app-framework/app-parameters/#secret-installation-parameters
[details]: https://www.contentful.com/developers/docs/references/content-management-api/app-details/
[definition-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-definition.ts
[installation-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-installation.ts
[installation-org]: https://www.contentful.com/developers/docs/references/content-management-api/app-installations/get-all-installations-of-an-app-within-an-organization/
[grant-create]: https://www.contentful.com/developers/docs/references/content-management-api/app-access-grants/create-one-access-grant/
[grant-concept]: https://www.contentful.com/developers/docs/extensibility/app-framework/access-grant/
[limits]: https://www.contentful.com/developers/docs/platform/technical-limits/
