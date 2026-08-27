# Contentful HTTP retry policy

Status: current provider design. This note defines the provider HTTP layer's
automatic retry and deadline boundaries for Contentful Management API requests.

## Deadline budget

Every CMA request has a finite context deadline. Resource and data-source
operations establish their configured Terraform timeout, normally defaulting
to two minutes, before calling the CMA client. Those operation deadlines remain
authoritative. If another call path, such as a list resource or internal CMA
operation, reaches the provider HTTP layer without a deadline, that layer adds
the same two-minute provider default to the request.

The context deadline or cancellation is the effective retry budget. The HTTP
layer does not expose or enforce a practitioner-configurable retry count.
Backoff waits select on the request context, so a long Contentful reset value
cannot extend an operation past its deadline.

Before entering a retry wait for a retryable HTTP response, the provider
calculates the normal backoff. If the context is still active but that
response-retry wait cannot complete before its deadline, the provider declines
the retry immediately and returns the final HTTP response without draining it.
This differs from a context that has already expired or been cancelled, or that
expires or is cancelled during a retry wait: in those cases the context error
remains authoritative.

## Retry classification

The provider retries:

- explicit HTTP 429 responses for every method, following Contentful's
  documented rate-limit handling and first-party client practice; and
- transport failures and retryable server responses for safe GET, HEAD, and
  OPTIONS requests.

The provider does not transparently replay POST, PUT, PATCH, or DELETE after a
transport failure or an ordinary 5xx response. Those outcomes do not establish
whether Contentful committed the mutation, so replay could repeat an already
applied write.

### 429 evidence boundary

| Evidence | Establishes | Does not establish |
| --- | --- | --- |
| Contentful's [CMA rate-limit documentation](https://www.contentful.com/developers/docs/references/content-management-api/overview/#api-rate-limits) | 429 represents rate limiting and the reset interval tells clients when to retry | Whether a mutation returning 429 committed |
| Contentful's [first-party management SDK](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/README.md#L387-L389) | The SDK retries 429 and 500 responses by default | A server commitment guarantee |

The provider retains all-method 429 retries for Contentful interoperability,
accepting residual ambiguity for mutation commitment. Transport failures and
ordinary 5xx responses remain non-retried for POST, PUT, PATCH, and DELETE.

For a valid `X-Contentful-RateLimit-Reset` response, the reset is the earliest
retry time. The provider waits for the reset plus 100ms and full jitter from a
contention window that starts at 500ms, doubles for each retryablehttp backoff
attempt, and caps at four seconds. The reset value itself is never multiplied.
Missing, invalid, or unrepresentable reset values retain retryablehttp's
`Retry-After` and linear-jitter fallback behavior.

If retryablehttp reaches its terminal error-handler path, the final HTTP
response or underlying transport error is passed through rather than replaced
with a generic retry-exhaustion error. Cancellation or deadline expiry during a
backoff wait instead returns the context error and no prior 429 response.

## Safety boundary

For POST, PUT, PATCH, and DELETE, this policy prevents transparent replay by the
provider HTTP layer after transport failures and ordinary 5xx responses. The
accepted all-method 429 policy retains the mutation-commitment ambiguity
described above. It does not provide at-most-once semantics across separate
Terraform operations or applies. In particular, repeating a create whose first
result was ambiguous can create another remote object when Contentful generates
the identifier.
