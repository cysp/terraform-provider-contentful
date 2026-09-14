# Repository guidance

## Execution and scope

- Complete requested implementation work with relevant verification and documentation. For reviews and plans, deliver analysis without implementing changes unless requested.
- Resolve routine choices from repository evidence. Ask when an unresolved question materially affects the public contract, scope, or authorization; continue independent work while awaiting the answer. Carry forward agreement already given in the conversation.

## Terminology and design

- Preserve established Contentful, Terraform, API, protocol, and codebase terminology; do not substitute terminology merely to make naming seem more descriptive. Keep Contentful `sys.version` scoped as `version`; do not invent `currentVersion`. Add a term only when established vocabulary cannot express a concrete need, and confirm non-obvious terminology with the maintainer before introducing it.
- Prefer the simplest cohesive implementation that preserves required behavior, public contracts, and explicit lifecycle boundaries. Evaluate the affected concern as a whole, and retain complexity only for a concrete current requirement. Treat requested examples as evidence of the desired outcome rather than an exhaustive checklist unless the user explicitly limits the scope. Keep each fact and policy in one authoritative place.
- Before adding an abstraction, mode flag, shallow wrapper, provider-private marker, provider-private status field, or production seam, identify the concrete current production behavior or lifecycle boundary that requires it. Provider-private markers and status fields require maintainer agreement and must address that behavior or boundary rather than merely simplifying local control flow. Do not add production seams solely for tests.
- When changing Terraform schemas, planning, validation, request conversion, response projection, or state publication, read and follow [Terraform value semantics](docs/design/terraform-value-semantics.md).
- When changing HTTP retries, deadlines, or mutation recovery, read and follow the [Contentful HTTP retry policy](docs/design/contentful-http-retry-policy.md).

## Evidence and tests

- Support implementation claims with repository code or direct experiments; support external behavior and compatibility claims with primary sources or direct experiments. State what was and was not verified.
- Choose tests for independent behavioral evidence, not assertion count. For request and lifecycle behavior, prefer exact request and version checks plus end-to-end coverage of the affected lifecycle transitions; do not derive expected results from the production logic under test.
- Before choosing checks, read the [validation scope](DEVELOPMENT.md#validation-scope) and the command sections relevant to the change: [generation](DEVELOPMENT.md#code-generation), [tests](DEVELOPMENT.md#tests), or [linting](DEVELOPMENT.md#linting). Once appropriate checks pass, broaden or repeat them only for new changes, failures, or unresolved concerns.

## Documentation and workflow

- Keep durable documentation current: record user-visible contracts, invariants, evidence, and limitations; do not retain dated audit inventories, cleanup chronology, or completed plans.
- Leave unrelated concerns out of each change, preserve unrelated worktree changes, and use a separate worktree and pull request when concurrently pursuing an independent concern; keep history reviewable.
- After changing a schema or another input to generated code or documentation, run `go generate ./...` and inspect both tracked changes and untracked output.
- Review the final diff against the request and these rules. Report completed work, verification, and any remaining blocker, with relevant evidence.

## Commit messages

- Use Conventional Commit messages. The type describes the kind of change, the scope identifies the affected component or maintenance concern, and the subject describes the specific change.
- For components under `internal/`, use the immediate directory's exact name as the scope, such as `contentful-management-go` or `provider`. Name individual resources in the subject.
- Tests, examples, generated files, and documentation about a component use that component's scope, regardless of their location.
- For repository maintenance, use the established concern name, such as `deps`, `lint`, `generate`, `release`, or `agents`. Scope workflow changes by their purpose.
- When a cohesive change spans components, use its primary concern as the scope if one clearly owns the change; otherwise omit the scope. Split independently useful changes into separate commits where appropriate.
- For a new concern, prefer an existing directory, tool, or workflow name. Use recent history to resolve choices left open by these rules.
