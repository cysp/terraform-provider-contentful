# Repository guidance

## Execution and scope

- Complete requested implementation work with relevant verification and documentation. For reviews and plans, deliver analysis without implementing changes unless requested.
- When reviewing a pull request, commit range, or working-tree diff, follow [Reviewing changes](docs/reviewing.md).
- Resolve routine choices from repository evidence. Ask when an unresolved question materially affects the public contract, scope, or authorization; continue independent work while awaiting the answer. Carry forward agreement already given in the conversation.

## Terminology and design

- Preserve established Contentful, Terraform, API, and protocol terminology. Keep Contentful `sys.version` scoped as `version`. Rename internal symbols when a concrete correctness or comprehension benefit justifies the change cost. Use established vocabulary where it expresses the concept; confirm non-obvious domain terminology with the maintainer before introducing it.
- Choose the simplest cohesive design for the whole affected concern, including its callers and lifecycle transitions. Preserve required behavior, public contracts, and lifecycle guarantees; distinguish these obligations from implementation choices recorded in code, tests, or documentation. Assess existing and proposed implementations against the same current requirements. Restructure when concrete correctness or maintainability benefits outweigh migration, regression, review, and maintenance costs. Treat change size as a cost, not the objective.
- Treat requested examples as evidence of the desired outcome unless the user explicitly limits the scope. Keep each fact and policy in one authoritative place.
- When adding defensive mechanisms or removing safeguards, identify the failure scenario, required guarantees, evidence, and uncertainty. Compare alternatives that preserve those guarantees, including their complexity, operational costs, and introduced failure modes. Reserve explicit tradeoff explanations for consequential or uncertain decisions.
- Extract abstractions and helpers when they expose a coherent operation or lifecycle boundary, remove meaningful duplication, or materially simplify understanding the caller. Judge extraction by these benefits, independently of naming or call count. Name helpers for the operations they expose, and keep simple one-off expressions local.
- Before adding a mode flag, shallow wrapper, provider-private marker, provider-private status field, or production seam, identify the concrete current production behavior or lifecycle boundary that requires it. Provider-private markers and status fields also require maintainer agreement. Production seams must serve a purpose beyond testing.
- When changing or reviewing Terraform schemas, planning, validation, request conversion, response projection, or state publication, read and follow [Terraform value semantics](docs/design/terraform-value-semantics.md).
- When changing or reviewing HTTP retries, deadlines, or mutation recovery, read and follow the [Contentful HTTP retry policy](docs/design/contentful-http-retry-policy.md).
- For resource-specific behavior or CMA test-server behavior and defaults, use the [design documentation map](docs/design/README.md#design-contracts) to find and follow the applicable contract.

## Evidence and tests

- Support implementation claims with repository code or direct experiments; support external behavior and compatibility claims with primary sources or direct experiments. State what was and was not verified.
- Choose tests for independent behavioral evidence, not assertion count. For request and lifecycle behavior, prefer exact request and version checks plus end-to-end coverage of the affected lifecycle transitions; do not derive expected results from the production logic under test.
- When changing or reviewing tests, read [Testing](docs/testing.md).
- Before choosing checks, read the [validation scope](DEVELOPMENT.md#validation-scope) and the command sections relevant to the change: [generation](DEVELOPMENT.md#code-generation), [tests](DEVELOPMENT.md#tests), or [linting](DEVELOPMENT.md#linting). Once appropriate checks pass, broaden or repeat them only for new changes, failures, or unresolved concerns.

## Documentation and workflow

- When writing or reviewing documentation, read the [documentation practices](docs/design/provider-documentation.md).
- Keep durable documentation current: record user-visible contracts, invariants, evidence, and limitations; do not retain dated audit inventories, cleanup chronology, or completed plans.
- Leave unrelated concerns out of each change, preserve unrelated worktree changes, and use a separate worktree and pull request when concurrently pursuing an independent concern; keep history reviewable.
- When changing generated code or generated documentation, use [Code generation](DEVELOPMENT.md#code-generation) and, for documentation, [Documentation authoring](DEVELOPMENT.md#documentation-authoring) to locate and edit the authoritative inputs. After changing a schema or another generation input, run `go generate ./...` and inspect both tracked changes and untracked output.
- Review the final diff against the request and these rules. Report completed work, verification, and any remaining blocker, with relevant evidence.

## Commit messages

- Use Conventional Commit messages. The type describes the kind of change, the scope identifies the affected component or maintenance concern, and the subject describes the specific change.
- For components under `internal/`, use the immediate directory's exact name as the scope, such as `contentful-management-go` or `provider`. Name individual resources in the subject.
- Tests, examples, generated files, and documentation about a component use that component's scope, regardless of their location.
- For repository maintenance, use the established concern name, such as `deps`, `lint`, `generate`, `release`, or `agents`. Scope workflow changes by their purpose.
- When a cohesive change spans components, use its primary concern as the scope if one clearly owns the change; otherwise omit the scope. Split independently useful changes into separate commits where appropriate.
- For a new concern, prefer an existing directory, tool, or workflow name. Use recent history to resolve choices left open by these rules.
