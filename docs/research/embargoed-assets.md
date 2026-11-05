# Embargoed assets: protection modes and asset keys

Protection for embargoed assets is configured for a space. Asset keys are
short-lived credentials used to sign protected asset URLs.

## Protection modes

Published sources reviewed: 2026-11-07.

Contentful's [protection-mode guide][modes] describes three modes:

| Mode | Documented effect |
| --- | --- |
| Preparation | Signed URLs can be tested while CMA, CDA, and CPA continue returning public asset URLs. |
| Unpublished assets protected | Unpublished assets require signed URLs. CDA returns public URLs; CMA and CPA return unsigned embargoed asset URLs for signing. |
| All assets protected | Both published and unpublished assets require signed URLs. CMA, CDA, and CPA return unsigned embargoed asset URLs for signing. |

For this feature, an asset counts as published if its ID is published in any
environment in the space. Publication in the addressed environment alone therefore
does not determine protection in Unpublished assets protected mode. The guide says
existing public URLs for assets covered by either protection mode stop working
within 48 hours of enabling it.

## Asset keys

The [CMA create endpoint][create-key] uses
`POST /spaces/{space_id}/environments/{environment_id}/asset_keys`. The request
supplies `expiresAt` as Unix time in seconds, at most 48 hours ahead. The response
supplies `policy` and `secret`, which are used to sign asset URLs. A signed URL
cannot outlive its asset key, though its JWT can set an earlier expiry. See the
[AssetKey reference][keys] for signing details.

Key creation and space protection are separate operations. The key endpoint's
environment path does not narrow the space-wide protection rule described above.

## Protection-mode API observations

On 2026-07-14, `GET /spaces/{space_id}/embargoed_assets` with a personal management
token returned `403 Forbidden`. The cause of that denial was not determined. No
protection mode, asset publication, or signing credential was changed.

The study did not establish an HTTP contract for changing protection modes.
Accepted wire values, transition responses, and concurrency behavior remain
unverified.

[modes]: https://www.contentful.com/help/media/embargoed-assets/embargoed-assets-modes/
[keys]: https://www.contentful.com/developers/docs/references/content-management-api/asset-keys/
[create-key]: https://www.contentful.com/developers/docs/references/content-management-api/asset-keys/create-an-asset-key/
