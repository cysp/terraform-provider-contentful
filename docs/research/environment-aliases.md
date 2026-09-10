# Environment aliases: routing and response identity

An environment alias routes requests to its selected environment. A child resource's
returned environment link can retain the alias used in the request, so that link alone
does not identify the concrete environment serving the data.

## Published routing contract

Contentful's [environment alias guide][guide] describes substituting an alias ID for
the environment ID when requesting environment resources. The alias itself is addressed
at `/spaces/{space_id}/environment_aliases/{environment_alias_id}`. Its top-level
`environment` link identifies the target, while `sys.id` identifies the alias. The
pinned [SDK entity][entity] preserves these separate properties.

## Direct read observations

Observed: 2026-09-09 (UTC), paired passive GETs through an existing alias and its
concrete target. The alias document was read before and after the comparison and was
identical.

| Resource | Alias and concrete-target comparison |
| --- | --- |
| [UI Config](ui-config.md) | Alias request returned the alias ID at `sys.environment.sys.id`; concrete request returned the target ID. The responses were otherwise equal after removing only this environment link. |
| [Locales](locales.md) | Returned locale environment links likewise echoed the addressed alias or concrete ID. The compared responses were otherwise equal after removing those links. |

Separate [AppInstallation observations](app-framework/configuration.md#observed-installation-reads)
also found metadata echoing the addressed alias. [Live preview variables](live-preview-variables.md#errors-and-alias-reads)
have an earlier, more limited alias-read observation; one family's result does not
establish the projection of every environment-scoped resource.

## Interpretation and limits

The alias document supplies the target relationship; the sampled child responses
preserve the addressing context. Resolving an alias and then reading its target does
not by itself create an atomic snapshot. Matching alias reads before and after the
comparison establish equality at those two points, not continuous stability between
them.

These reads did not retarget an alias, clone an environment, or write through an alias.
They establish neither propagation time nor behavior during concurrent retargeting.
Environment creation has a separate [readiness contract](environment-readiness.md).

[guide]: https://www.contentful.com/developers/docs/concepts/environment-aliases/
[entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/environment-alias.ts
