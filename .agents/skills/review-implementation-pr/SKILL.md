---
name: review-implementation-pr
description: Review an implementation PR, branch, or local diff against its intended outcome, including re-review after fixes or a rebase, or prepare a prompt for an independent reviewer.
---

# Review an implementation PR

- For a direct review, give the maintainer one recommendation that accounts for any upgrade findings. Say when no material findings remain.
- For a requested prompt or a delegated review, [prepare the handoff](../../../docs/reviewing.md#preparing-a-review-prompt). Give the reviewer the decision to make, the settled requirements and the evidence to examine. Leave the recommendation to that reviewer.

## Select the change

Use the target the developer supplied. For a PR, resolve its actual base, head, and merge base. For a local branch, infer the intended target from its PR or repository context. For a working-tree review, include staged and unstaged changes and relevant untracked files. State the comparison and whether it includes local work. Ask only when plausible targets would give materially different reviews.

Establish the intended outcome from the request, PR description, any linked issue, and existing contracts. Distinguish a stated requirement from an inference. If intent remains uncertain, review the parts you can verify and explain which decision depends on clarification.

## Review the implementation

Use the [repository review guidance](../../../docs/reviewing.md) to test the intended outcome across the changed components, including upgrade impact where applicable. Start from the documented entry point. Compare the prerequisites a reader can infer there with those the implementation and tests actually use. When a simpler design could reduce a concrete burden on callers or maintainers, compare it against the same lifecycle guarantees.

## Review after a fix or rebase

Review the current revision against its base and intended outcome, even in a fresh conversation. If an earlier finding is available, check its original trigger against the current code and report what happens now. Inspect the fix for effects on other callers.

If you know both the old and new commit ranges, `git range-diff OLD_BASE..OLD_HEAD NEW_BASE..NEW_HEAD` can focus attention after a rebase or rewrite. Inspect relevant base changes as well. Claim a particular defect was fixed only when you can establish the original failure and its current outcome.
