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

The [CMA reference][cma-aliases] explains when `master` becomes an alias. A Space
admin opts in through **Settings → Environments**, which creates a permanent
`master` alias. Its example of `sys.aliasedEnvironment` describes the Environment
returned by `GET /environments/{alias}`. The example contains no child resource.

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

### Entry and Content Type detail reads

Contentful's Ruby Management SDK recorded three consecutive CMA GETs in the same
Space on January 7, 2021. All three returned HTTP 200.

| GET relative to `/spaces/{space}/environments` | Recorded identity |
| --- | --- |
| `/master` | The [Environment response][rb-env] has `sys.id = master` and `sys.aliasedEnvironment.sys.id = backup-1`. |
| `/master/entries/{entry}` | The [Entry response][rb-entry] has `sys.environment.sys.id = master`. Its [Content Type link][rb-link] identifies `topicArticle`. |
| [`/master/content_types/topicArticle`][rb-type-get] | The [Content Type response][rb-type] is stored as base64-encoded JSON. Decoding it gives `sys.id = topicArticle` and `sys.environment.sys.id = master`. |

Both children returned a `Link` with `linkType = Environment`. A second
[Asset recording][rb-asset] from the same Space and date also returned `master`
in that link. These recordings show the alias in individual GET responses. They
contain no child GET through `backup-1`, collection comparison, or alias-routed
write. The published [Entry][entry-get] and [Content Type][type-get] examples
do not identify a distinct alias and target.

In a second Space, a [2022 Environment collection][rb-main] lists `master` as
an alias on the concrete environment `main`. Contentful's Python SDK recorded
an [Entry GET][py-entry] and a [Content Type GET][py-type] through `master` in
September 2023. Both returned `master` in `sys.environment.sys.id`. A
[September 2025 Content Type GET][py-type-2025] shows the same link. Neither
Python recording checks which concrete environment the alias targeted that day.

The Python SDK also recorded an [Entry creation][py-entry-create] and
[update][py-entry-update], a [Content Type update][py-type-update] and a
[Tag collection][py-tags] through `master` in that Space. The mutation responses
and returned Tags carry `master` in their own `sys.environment.sys.id`. The
recordings identify the request address and returned links. The contemporary
alias target and parity with a direct-target response remain unverified.

On 2026-11-28 (UTC), the example account's management credential could read one
Space (`GET /spaces?limit=1`, HTTP 200, `total: 1`). The following GETs used that
Space. `{entry}` and `{type}` stand for the example Entry and the Content Type
referenced by its `sys.contentType` link.

| GET relative to `/spaces/{space}` | HTTP | Identity observed |
| --- | --- | --- |
| `/environment_aliases` | 200 | `total: 0`, `items: []` |
| `/environment_aliases/master` | 404 | No alias document for the configured address. |
| `/environments/master` | 200 | `sys.type = Environment`, `sys.id = master`, `sys.aliasedEnvironment` absent. |
| `/environments/master/entries/{entry}` | 200 | `sys.type = Entry`, `sys.id = {entry}`. The Space, Environment and ContentType links identified `{space}`, `master` and `{type}`, respectively. |
| `/environments/master/content_types/{type}` | 200 | `sys.type = ContentType`, `sys.id = {type}`. The Space and Environment links identified `{space}` and `master`. |

Both child responses used `sys.type = Link` and the corresponding `sys.linkType`
inside each link. Neither child had `sys.aliasedEnvironment`. The two 200s
establish the concrete-environment baseline. The empty alias collection prevented
a fresh comparison with the SDK recording.

To repeat the comparison today, use a readable Space with an existing alias pointing to a
different ID and a target containing an Entry and its Content Type. Read the alias
document, GET the same two objects through each address, then reread the alias
document and check its target and version. For each GET, record HTTP status,
`sys.type`, `sys.id`, the Space and Environment links, the Entry's ContentType
link and any alias metadata. Keeping the resource IDs fixed makes it possible to
compare the two environment links without retaining customer content.

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
[cma-aliases]: https://www.contentful.com/developers/docs/references/content-management-api/environment-aliases/
[entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/environment-alias.ts
[rb-env]: https://github.com/contentful/contentful-management.rb/blob/bc118aa6150e814570749c9d939a8d75c782df79/spec/fixtures/vcr_cassettes/entry/issue_219.yml#L233-L245
[rb-entry]: https://github.com/contentful/contentful-management.rb/blob/bc118aa6150e814570749c9d939a8d75c782df79/spec/fixtures/vcr_cassettes/entry/issue_219.yml#L388-L398
[rb-link]: https://github.com/contentful/contentful-management.rb/blob/bc118aa6150e814570749c9d939a8d75c782df79/spec/fixtures/vcr_cassettes/entry/issue_219.yml#L416-L422
[rb-type-get]: https://github.com/contentful/contentful-management.rb/blob/bc118aa6150e814570749c9d939a8d75c782df79/spec/fixtures/vcr_cassettes/entry/issue_219.yml#L444-L446
[rb-type]: https://github.com/contentful/contentful-management.rb/blob/bc118aa6150e814570749c9d939a8d75c782df79/spec/fixtures/vcr_cassettes/entry/issue_219.yml#L529-L534
[rb-asset]: https://github.com/contentful/contentful-management.rb/blob/bc118aa6150e814570749c9d939a8d75c782df79/spec/fixtures/vcr_cassettes/asset/issue_219.yml#L388-L398
[rb-main]: https://github.com/contentful/contentful-management.rb/blob/bc118aa6150e814570749c9d939a8d75c782df79/spec/fixtures/vcr_cassettes/environment/all/limit_100.yml#L115-L158
[py-entry]: https://github.com/contentful/contentful-management.py/blob/6384b17bb6d376b49a1ae79362de8ce58e0dfd31/fixtures/tags/add_tag_to_an_entry.yaml#L19-L23
[py-type]: https://github.com/contentful/contentful-management.py/blob/6384b17bb6d376b49a1ae79362de8ce58e0dfd31/fixtures/tags/add_tag_to_an_entry.yaml#L102-L106
[py-type-2025]: https://github.com/contentful/contentful-management.py/blob/6384b17bb6d376b49a1ae79362de8ce58e0dfd31/fixtures/content_type/add_taxonomies_to_content_type.yaml#L19-L23
[py-entry-create]: https://github.com/contentful/contentful-management.py/blob/6384b17bb6d376b49a1ae79362de8ce58e0dfd31/fixtures/tags/create_entry_with_tag.yaml#L24-L28
[py-entry-update]: https://github.com/contentful/contentful-management.py/blob/6384b17bb6d376b49a1ae79362de8ce58e0dfd31/fixtures/tags/add_tag_to_an_entry.yaml#L189-L193
[py-type-update]: https://github.com/contentful/contentful-management.py/blob/6384b17bb6d376b49a1ae79362de8ce58e0dfd31/fixtures/content_type/add_taxonomies_to_content_type.yaml#L110-L114
[py-tags]: https://github.com/contentful/contentful-management.py/blob/6384b17bb6d376b49a1ae79362de8ce58e0dfd31/fixtures/tags/all.yaml#L19-L35
[entry-get]: https://www.contentful.com/developers/docs/references/content-management-api/entries/get-an-entry/
[type-get]: https://www.contentful.com/developers/docs/references/content-management-api/content-types/get-a-content-type/
