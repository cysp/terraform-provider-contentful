# Repository guidance

## Terminology and design

- Preserve established Contentful, Terraform, API, protocol, and codebase terminology; do not substitute terminology merely to make naming seem more descriptive. Keep Contentful `sys.version` scoped as `version`; do not invent `currentVersion`. Add a term only when established vocabulary cannot express a concrete need, and confirm non-obvious terminology with the maintainer before introducing it.
- Prefer the simplest cohesive implementation that preserves required behavior, public contracts, and explicit lifecycle boundaries. Evaluate the affected concern as a whole, and retain complexity only for a concrete current requirement. Treat requested examples as evidence of the desired outcome rather than an exhaustive checklist unless the user explicitly limits the scope. Keep each fact and policy in one authoritative place.
- Before adding an abstraction, mode flag, shallow wrapper, provider-private marker, provider-private status field, or production seam, identify the concrete current production behavior or lifecycle boundary that requires it. Provider-private markers and status fields require maintainer agreement and must address that behavior or boundary rather than merely simplifying local control flow. Do not add production seams solely for tests.
- When changing Terraform schemas, planning, validation, request conversion, response projection, or state publication, read and follow [Terraform value semantics](docs/design/terraform-value-semantics.md).

## Evidence and tests

- Support implementation claims with repository code or direct experiments; support external behavior and compatibility claims with primary sources or direct experiments. State what was and was not verified.
- Choose tests for independent behavioral evidence, not assertion count. For request and lifecycle behavior, prefer exact request and version checks plus end-to-end coverage of the affected lifecycle transitions; do not derive expected results from the production logic under test.

## Documentation and workflow

- Keep durable documentation current: record user-visible contracts, invariants, evidence, and limitations; do not retain dated audit inventories, cleanup chronology, or completed plans.
- Leave unrelated concerns out of each change, preserve unrelated worktree changes, and use a separate worktree and pull request when concurrently pursuing an independent concern; keep history reviewable.
- Use Conventional Commit messages. Choose a scope for the affected codebase area, following recent repository history when a matching scope exists.
- After changing a schema or another input to generated code or documentation, run `go generate ./...` and inspect the generated diff.
- Before every `golangci-lint run`, run `golangci-lint cache clean`.
- Run focused tests for changed behavior plus proportionate broader verification.
