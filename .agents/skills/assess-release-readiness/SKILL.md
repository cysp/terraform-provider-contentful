---
name: assess-release-readiness
description: Assess the next release action for terraform-provider-contentful from origin/main or an explicit release PR, or prepare a release-specific assessment prompt.
---

# Assess release readiness

Default to a readiness decision. Use an isolated checkout when a local check needs one. Changes to pull requests, workflows, tags, and the live Contentful account require separate authorization.

## Identify the release decision

If the maintainer supplies a release PR, assess that release even if it has merged. Otherwise fetch `origin/main` and record its commit. Use the [release process](../../../docs/releasing.md) at the revision under review to determine the next action.

| State of the release | Decision |
| --- | --- |
| A merged release is awaiting publication or Registry verification | Assess that release first. Identify any later commits on main separately. Check recovery and abandonment records for an older unpublished tag. |
| A release PR is open against main | Assess whether to merge it. |
| Main has unreleased changes and no prepared release | Assess whether to run **Prepare release PR**. Describe version and migration implications. The generated notes, version, and PR checks become verifiable after preparation. |
| Main matches the published release and Registry verification is complete | Report **nothing to release** with the version and commit. Stop here. |

Identify the preceding published provider version on the target release line and verify its tag commit and ancestry. Confirm publication on GitHub and in the Terraform Registry for that version. A draft, abandoned tag, or GitHub Latest designation can point at the wrong baseline.

For an open PR, pin its head, target tip, and candidate merge tree. Compare release metadata in the PR and shipped behavior across the full published-to-candidate range. If main advanced after preparation, name the intervening changes and recommend refreshing the PR through **Prepare release PR**. Reuse evidence for unaffected contracts when reviewing the refreshed candidate. For a merged PR, use its actual merged tree. Check the version, notes, manifest, and any override when they exist.

## Decide what this release needs to prove

Compare the published version with the candidate and inspect the resulting code. Group changes by the behavior they affect. When two changes meet in one operation, look for an assumption in the tests for one that the other could invalidate. Vary that assumption and trace each affected input through the operation to its observable result. Use the notes, maintainer input, and any supplied findings as leads. Which shipped changes need disclosure? Which claims in the notes, examples, and rendered Registry docs does the implementation support?

Frame each consequential question around the resource, attribute, dependency, or publishing step, its starting condition, the observable outcome, and evidence that could settle it. A dependency-only release may need build and compatibility evidence. For a breaking change, trace the first refreshed plan and apply with and without the recommended pre-upgrade action, including the Contentful Management API (CMA) request and resulting state where relevant.

Trace migration claims from the candidate’s practitioner documentation to the implementation, upgrade tests, and fixtures. Identify which released provider created the test state, the transition exercised, and what the recorded run demonstrates. If the tested baseline differs from the release users would upgrade from, inspect intervening changes and verify gaps that could alter the decision. Include import, delegation, replacement, interrupted writes, feature-specific Terraform requirements, or downgrade where they affect the answer. For credential changes check rotation, retention, and exposure in plans, public and private state, and diagnostics. Use redacted evidence.

## Report the decision or prepare a prompt

**Prompt request.** Write a prompt another reviewer can use in a fresh conversation. Include the release PR or pinned main revision, published baseline, questions, evidence to inspect, and decision to make. Include a pinned link to this procedure or the relevant instructions. Ask the reviewer to inspect the rest of the diff for other risks and refresh the revisions and publication status before recommending an action. If the maintainer requests a separate thread, use the same prompt there when the agent supports it.

**Assessment request.** Follow the [review guidance](../../../docs/reviewing.md) and [validation scope](../../../DEVELOPMENT.md#validation-scope). For each CI result you rely on, check the revision, triggering event, jobs, and relevant subtests that ran. PR and main runs can execute different tests on identical trees. Distinguish prior-provider upgrade evidence, mock CMA behavior, and live CMA observations. Run focused checks for material gaps.

Refresh refs, checks, and review or publication status before reporting. If the candidate moved, assess the intervening delta or qualify the verdict against its pinned revision. For an open PR, lead with **merge**, **merge after named conditions**, or **hold** and distinguish provider readiness from GitHub merge requirements. For a merged release, trace merged-commit CI, tag and draft creation, the publisher run, GitHub publication, and the Registry version and rendered docs. Link the first unfinished step or its recovery procedure. Recommend announcing only once the publication evidence supports it.

For each blocker, give its trigger, consequence, evidence, and smallest corrective action. Summarize supported upgrade paths and remaining uncertainty that could change the decision.
