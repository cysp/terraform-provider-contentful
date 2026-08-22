# Extension source values

Contentful's [UI Extensions CMA reference](https://www.contentful.com/developers/docs/references/content-management-api/ui-extensions/)
requires exactly one of `extension.src` or `extension.srcdoc`. The first-party
[`contentful-management.js` SDK](https://contentful.github.io/contentful-management.js/contentful-management/10.3.1/modules.html#createuiextensionprops)
models the same create contract with
`RequireExactlyOne<..., "src" | "srcdoc">`.

On 2026-08-21, isolated live CMA `PUT` requests tested the remaining empty-value
ambiguity. The requests used an empty `fieldTypes` list and `sidebar=false`:

| Request source members | Result | Response evidence |
| --- | --- | --- |
| `"src":""` | HTTP 422 | `extension.src` failed Contentful's URL regexp. |
| `"srcdoc":""` | HTTP 201 | The response retained `"srcdoc":""`; the probe resource was then deleted with HTTP 204. |
| Both empty | HTTP 422 | The response reported both the invalid `src` URL and the source-choice constraint. |

Provider impact: explicit empty `src` is rejected by configuration validation;
explicit empty `srcdoc` remains a meaningful known value and is serialized;
and both configured sources are rejected. HCL may omit both attributes for an
imported or otherwise response-owned source. Planning must preserve the sole
prior-state source so that the resulting CMA request still contains exactly one
source. These are point-in-time service observations supporting the first-party
contract, not a claim about unrelated Extension fields.
