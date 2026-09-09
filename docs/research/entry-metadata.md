# Entry metadata: tag and concept ordering

Entry `metadata.tags` and `metadata.concepts` are arrays of links. Their order and
duplicate occurrences were echoed by immediate mutation responses but were not preserved
by subsequent reads and publication. This reference distinguishes that observed
normalization from the documented list representation.

## Scope and sources

Published sources reviewed: 2026-09-03. First-party source: `contentful-management.js`
v12.15.0 commit
[`cc096a3`](https://github.com/contentful/contentful-management.js/tree/cc096a337f0e1db6114e8da645d69bb6eb90f11c).

The
[Entry](https://www.contentful.com/developers/docs/references/content-management-api/entries/),
[Tags](https://www.contentful.com/developers/docs/references/content-management-api/tags/#tags-on-entries-and-assets),
and
[Taxonomy](https://www.contentful.com/developers/docs/references/content-management-api/taxonomy/#concepts-on-entries)
references describe both properties as lists of links, but do not define order or
repeated-link semantics. The first-party client models both as arrays in
[`MetadataProps`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/common-types.ts#L486-L489),
[sends supplied metadata
unchanged](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L155),
and [wraps returned
data](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/entities/entry.ts#L64-L68)
without normalizing it. That is evidence of client behavior, not a CMA preservation
guarantee.

## Direct observations

Experiment dates unrecorded.

The probe used separate disposable Entries and activated Content Types for three private
tags and three concepts in a disposable concept scheme. Symbols identify resources by
creation order. For each property, the first unique submission created the Entry; the
remaining three submissions used full-body PUT at the exact current version. Each
mutation was followed by GET, whole-Entry Publish, and another GET.

| Property | Unique submissions | Duplicate submissions | Immediate responses | Later GET and Publish responses |
| --- | --- | --- | --- | --- |
| Tags | `T1,T2,T3`; `T3,T1,T2` | `T1,T2,T1`; `T2,T1,T2` | Echoed each submission | `T1,T2,T3` after unique; `T1,T2` after duplicate |
| Concepts | `C1,C2,C3`; `C3,C1,C2` | `C1,C2,C1`; `C2,C1,C2` | Echoed each submission | `C1,C2,C3` after unique; `C1,C2` after duplicate |

All eight mutations succeeded. For both properties, mutation responses echoed submitted
order and duplicates; GET and Publish restored the first assignment order and reduced
repeated links to one occurrence. Neither characteristic was therefore durable. The
matching tag and concept results were measured independently. They do not establish a
documented set contract or a general canonical-order rule.

The probe covered private tags and concepts in a concept scheme, using separate
disposable Entries and activated Content Types. It did not establish regional parity or
cover public tags, Assets, PATCH, locale-based publishing, concurrent mutations, or
other collection sizes. The symbols above are synthetic aliases; they preserve
assignment order and multiplicity without identifying resources.
