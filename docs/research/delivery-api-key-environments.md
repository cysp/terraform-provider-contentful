# Delivery API key `environments` behavior

Observed against the Contentful CMA on 2026-07-28 using temporary Delivery API
keys:

| Request value | Create | Update from `[test]` |
|---|---|---|
| property omitted | returns `[master]` | returns `[master]` |
| `[]` | returns `[master]` | returns `[master]` |
| explicit environment links | returns those links | returns those links |
| `null` | HTTP 400 | HTTP 400; existing value is unchanged |

Successful create, update, and subsequent get responses contained an
`environments` array. The linked Preview API key reflected the same environment
selection as its Delivery API key.

These are point-in-time observations rather than a published compatibility
guarantee. In particular, the API could later preserve an empty list or choose a
default environment with a different identifier.

## Design impact

The provider applies the ownership and known/null/unknown rules in
[Terraform value semantics](../design/terraform-value-semantics.md#request-conversion)
before converting `environments`. In particular, an unknown plan for a
configuration-owned list is an error, not omission.

For values accepted by that boundary, the generated client preserves the
request distinction:

- A null list or a response-owned unknown list becomes a nil Go slice, so the
  request omits `environments`.
- A known empty Terraform list becomes a non-nil empty Go slice, so the request
  contains `"environments":[]`.
- A known populated list becomes the corresponding Environment links.
- Response conversion preserves the response shape rather than reconstructing or
  validating a presumed default.

This lets Contentful define the effective environment selection and ensures that
future valid empty responses or changes to its default are represented faithfully.

## Supporting documentation

- Contentful documents CMA updates as full replacements rather than merges:
  [Updating content](https://www.contentful.com/developers/docs/references/content-management-api/overview/#updating-content).
- Contentful's official Ruby SDK defaults an omitted create argument to `[]` and
  notes that an empty value defaults to master:
  [`ApiKey`](https://github.com/contentful/contentful-management.rb/blob/master/lib/contentful/management/api_key.rb).
- The REST examples show `environments` represented as an array of Environment
  links:
  [create](https://www.contentful.com/developers/docs/references/content-management-api/api-keys/create-a-delivery-api-key/),
  [update](https://www.contentful.com/developers/docs/references/content-management-api/api-keys/update-a-delivery-api-key/), and
  [get](https://www.contentful.com/developers/docs/references/content-management-api/api-keys/get-a-delivery-api-key/).
