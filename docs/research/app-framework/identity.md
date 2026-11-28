# App Identity: keys and access tokens

An app signs a JWT with its private key and exchanges it for an app access token
scoped to one installation. [Request signing](request-signing.md)
and [app sharing](configuration.md#appaccessgrant-and-sharing) have separate purposes.

See the [shared study scope and sources](README.md#scope-and-evidence) and [API
inventory](README.md#api-inventory).

## App keys and app access tokens

An app key holds the public key Contentful uses to verify the JWT. The CMA uses
RSA/RS256 public JWKs (`kty: RSA`, `alg: RS256`, `use: sig`), permits up to three keys
per app, and does not allow sharing a key pair across apps. The API can generate a key
pair and return private material once, or accept a supplied public JWK. There is no
update endpoint; rotation is create, deploy, delete. [App keys][keys].

Key creation sends `{generate: true}` or `{jwk: ...}`. Responses contain `jwk` and, in
generated mode, one-time `generated.privateKey` (described as base64 PEM by the SDK).
`sys.id` is the key/fingerprint used in its URL; metadata links the organization and app
definition, without a version. The documented public key representation uses base64 DER
public-key bytes in `x5c[0]` and their base64url SHA-256 fingerprint for both `kid` and
`x5t`. This is the Contentful representation described by its example, not a general JWK
certificate-chain normalization rule. [Key entity][key-entity], [key reference][keys].

An app access token is valid for 10 minutes and scoped to one installation. For the JWT
that requests it, Contentful lists `iss` (the app definition ID), `iat` and `exp` as
mandatory claims. It also specifies `typ: JWT` and `alg: RS256` in the header.
The official toolkit sets the app JWT's lifetime with `expiresIn: '10m'`. The resulting
`exp` claim gives the expiration time in seconds since the Unix epoch.
[Token reference][tokens], [token implementation][token-toolkit],
[JWT expiration claim](https://www.rfc-editor.org/rfc/rfc7519.html#section-4.1.4).

The management SDK accepts `spaceId`, `environmentId`, `appDefinitionId` and a
`{jwt}` argument. It sends a bodyless `POST /spaces/{space_id}/environments/{environment_id}/app_installations/{app_definition_id}/access_tokens`
with `Authorization: Bearer <app JWT>`. The toolkit calls the last path parameter
`appInstallationId` and uses the same value as the JWT issuer. The management
SDK's integration test uses `appDefinition.sys.id` for both. The SDK models the
issued credential as `token` and its expiry as `sys.expiresAt`.
[SDK adapter][token-sdk], [SDK integration test][token-test], [toolkit][token-toolkit],
[response type][token-entity].

Current identity documentation names Asset, ContentType, EditorInterface, Entry, Locale,
Release, ReleaseAction, ScheduledAction, Snapshot (master only), Tag, Task, and the
app's own app installation. The token overview omits some release/scheduling entities, so
those permissions remain a documented discrepancy rather than a tested matrix. Backend
identities act independently of a particular user's permissions; frontend operations on
behalf of users remain subject to user access. [App Identity][identity], [Framework
overview][framework].

## Direct CMA observations for token issuance

On 2026-11-27 UTC, ten exchanges against one installation returned HTTP 201,
the status expected by the toolkit. The [issuance reference][issue] documents
HTTP 200. The installation's `sys.id` matched the app definition ID. Sending a
management token to the same endpoint returned 401 `AccessTokenInvalid`.

RS256 JWTs containing only `iss`, `iat` and `exp` succeeded both with a `kid`
identifying the registered key and with `kid` omitted. The experiment did not test
omitting any of the three claims.

Signed JWTs valid for 60, 120, 300 and 600 seconds all produced app access tokens valid
for 600 seconds. The response's `sys.expiresAt` matched the `exp` in the app access
token returned by Contentful. One token still read its installation three seconds
after the app's signed JWT expired. Five seconds after the issued token expired, the same read
returned 401. Adding a valid `nbf` or a unique `jti` to the signed JWT did
not copy that claim into the issued token.

The response linked the app definition, space and environment and omitted
`sys.id` and `sys.version`. Both `sys.createdBy` and `sys.updatedBy` linked
to the app definition. The [reference example][issue] shows User links.

## App access token permissions

On 2026-11-28 UTC, a disposable app created and published an Author Entry
in a test environment. An independent CMA read confirmed the Author's name
in the default locale and a published version. The Entry's `sys.createdBy`,
`sys.updatedBy` and `sys.publishedBy` all linked to the App Definition used
to obtain the token.

| Requests made as the installed app | CMA result |
| --- | --- |
| Create, read, update, activate, deactivate and delete a Content Type | All succeeded. |
| Read and update that Content Type's Editor Interface | Both returned 200. A management read confirmed the change. |
| Create, read, update and delete a private Tag | All succeeded. |
| List Locales | Returned 200. |
| Read the app's own installation | Returned 200, including the submitted ordinary and Secret parameter values. |
| Update or delete the app's own installation | Both returned 403 `AccessDenied`. The same update with a management credential succeeded. |
| Read the app's own App Definition | Returned 404 `NotFound` while the definition existed. |
| Read the installation's Space or Environment | Both returned 401 `AccessTokenInvalid` with a working app token. |

The [App Identity overview][identity] lists the app's own installation among
the entities it can manage. In these installations the app could read its
parameters, while changing them required a management credential. A subsequent
app-authenticated read returned the values from the management update.

Creating an optional Locale returned 403 with both the app token and the
management credential, including for a standard locale code absent from the
environment. This account could not distinguish an app-specific restriction
from an account limitation. Locale updates and deletions were not exercised.

## Key rotation

Contentful recommends registering the new key, deploying it, then removing
the old key. When the signed JWT omits `kid`, Contentful checks the app's
registered keys. How quickly a key change takes effect and whether deleting a key
affects previously issued app access tokens remain unverified. [App key FAQ][key-faq], [delete an app key][delete-key].

[token-sdk]: https://github.com/contentful/contentful-management.js/blob/722a8efde640ac74571a0730e3f56e339a2d7a69/lib/adapters/REST/endpoints/app-access-token.ts
[token-test]: https://github.com/contentful/contentful-management.js/blob/722a8efde640ac74571a0730e3f56e339a2d7a69/test/integration/app-access-token-integration.test.ts#L38-L85
[issue]: https://www.contentful.com/developers/docs/references/content-management-api/app-access-token/issue-a-token-for-an-app-installation-in-a-space-environment/
[keys]: https://www.contentful.com/developers/docs/references/content-management-api/app-keys/
[key-faq]: https://www.contentful.com/help/faq/extensibility/
[delete-key]: https://www.contentful.com/developers/docs/references/content-management-api/app-keys/delete-an-app-key/
[tokens]: https://www.contentful.com/developers/docs/references/content-management-api/app-access-token/
[token-toolkit]: https://github.com/contentful/node-apps-toolkit/blob/91ac4e0c4bb3b4d06c0806d9d611ee3bc2e629a4/src/keys/get-management-token.ts
[identity]: https://www.contentful.com/developers/docs/extensibility/app-framework/app-identity/
[framework]: https://www.contentful.com/developers/docs/extensibility/app-framework/overview/
[key-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-key.ts
[token-entity]: https://github.com/contentful/contentful-management.js/blob/722a8efde640ac74571a0730e3f56e339a2d7a69/lib/entities/app-access-token.ts
