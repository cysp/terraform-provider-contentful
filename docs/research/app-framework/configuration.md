# App configuration and installation

AppDefinition holds shared app configuration; AppInstallation places that app in an
environment. AppDetails supplies presentation metadata, and AppAccessGrant controls
which organizations may install the app.

See the [shared study scope and sources](README.md#scope-and-evidence) and [API
inventory](README.md#api-inventory). The configuration probes used uninstalled apps;
installation and access-grant behavior below comes from the cited sources.

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

Installation parameter size and shape have conflicting source descriptions: the CMA
overview describes an object limited to **16 kB** after stringification; the pinned SDK
comment says **32 KB**, and its free-form type also permits arrays and scalars. Neither
the exact byte boundary nor non-object acceptance was tested. Cloning an environment is
documented to copy installations and parameters. [App installations][installations],
[installation entity][installation-entity], [free-form type][parameter-types].

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

Clearing/defaults for each optional field, direct bundle selection at definition POST,
effects of changed parameter declarations on existing installations, exact parameter
byte limit and non-object acceptance, and a complete role/concurrency matrix.

Grant revocation effects on existing installations, exact targeted-grant enum, and
derivation of shared metadata.

[definition-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/app-definition.ts
[installation-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/app-installation.ts
[parameter-types]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/widget-parameters.ts
[details-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-details.ts
[details-tests]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/test/integration/app-details-integration.test.ts
[definitions]: https://www.contentful.com/developers/docs/references/content-management-api/app-definitions/
[installations]: https://www.contentful.com/developers/docs/references/content-management-api/app-installations/
[details]: https://www.contentful.com/developers/docs/references/content-management-api/app-details/
[definition-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-definition.ts
[installation-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-installation.ts
[installation-org]: https://www.contentful.com/developers/docs/references/content-management-api/app-installations/get-all-installations-of-an-app-within-an-organization/
[grant-create]: https://www.contentful.com/developers/docs/references/content-management-api/app-access-grants/create-one-access-grant/
[grant-concept]: https://www.contentful.com/developers/docs/extensibility/app-framework/access-grant/
[limits]: https://www.contentful.com/developers/docs/platform/technical-limits/
