# App Identity: keys and access tokens

AppKey establishes an app's asymmetric identity. A signed app JWT can be exchanged for
an AppAccessToken scoped to one installation. [Request signing](request-signing.md)
and [app sharing](configuration.md#appaccessgrant-and-sharing) have separate purposes.

See the [shared study scope and sources](README.md#scope-and-evidence) and [API
inventory](README.md#api-inventory). The permission discrepancies below remain
unverified.

## AppKey and AppAccessToken

AppKey is asymmetric identity, not the symmetric event signing secret. The CMA uses
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

An AppAccessToken is valid for 10 minutes and scoped to one installation. The app JWT
used to obtain it identifies the AppDefinition through `iss` and has standard `iat`
and `exp` claims. The token reference's duration wording is ambiguous. The official toolkit
sets the app JWT's ten-minute lifetime with `expiresIn: '10m'`. JWT `exp` instead records
a NumericDate expiration timestamp; a literal `exp: 600` does not express a ten-minute
lifetime. [Token reference][tokens], [token implementation][token-toolkit],
[JWT expiration claim](https://www.rfc-editor.org/rfc/rfc7519.html#section-4.1.4).

The SDK token helper takes `{jwt}`, but converts it to `Authorization: Bearer <app JWT>`
on a POST with **no request body**. The response contains `token` and `sys.expiresAt`,
plus definition/space/environment links; the SDK omits id/version. This exchange is
distinct from sending a management access token to configure organization-owned
resources. [Token adapter][token-sdk], [token entity][token-entity].

Current identity documentation names Asset, ContentType, EditorInterface, Entry, Locale,
Release, ReleaseAction, ScheduledAction, Snapshot (master only), Tag, Task, and the
app's own AppInstallation. The token overview omits some release/scheduling entities, so
those permissions remain a documented discrepancy rather than a tested matrix. Backend
identities act independently of a particular user's permissions; frontend operations on
behalf of users remain subject to user access. [App Identity][identity], [Framework
overview][framework].

## Unresolved behavior

The evidence does not establish how key revocation affects already issued tokens or how
quickly a rotation takes effect.

[token-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/app-access-token.ts
[keys]: https://www.contentful.com/developers/docs/references/content-management-api/app-keys/
[tokens]: https://www.contentful.com/developers/docs/references/content-management-api/app-access-token/
[token-toolkit]: https://github.com/contentful/node-apps-toolkit/blob/64fa31b6b2223cd1c8b1798fa540e8aad5e2d319/src/keys/get-management-token.ts
[identity]: https://www.contentful.com/developers/docs/extensibility/app-framework/app-identity/
[framework]: https://www.contentful.com/developers/docs/extensibility/app-framework/overview/
[key-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-key.ts
[token-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-access-token.ts
