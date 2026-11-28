---
name: review-implementation-pr
description: Review an implementation PR, branch, or local diff against its intended outcome, including re-review after fixes or a rebase.
---

# Review an implementation PR

## Select the change

Use the target the developer supplied. For a PR, resolve its actual base, head, and merge base. For a local branch, infer the intended target from its PR or repository context. For a working-tree review, include staged and unstaged changes and relevant untracked files. State the comparison and whether it includes local work. Ask only when plausible targets would give materially different reviews.

Establish the intended outcome from the request, PR description, any linked issue, and existing contracts. Distinguish a stated requirement from an inference. If intent remains uncertain, review the parts you can verify and explain which decision depends on clarification.

## Review the implementation

Follow the [repository review guidance](../../../docs/reviewing.md), including its upgrade assessment when the diff affects existing users. Trace the changed behavior from a caller through the operation that delivers the requested outcome. Look for a missing outcome even when the code is internally consistent and CI is green. Evaluate design concerns by their effect on a caller, lifecycle guarantee, or maintenance task.

Before relying on an existing check, identify the revision it ran on and the behavior it exercised. Give the author one review ordered by consequence, including any upgrade findings. State what needs to change before merge, label optional improvements, and say when no material findings remain.

## Review after a fix or rebase

Review the current revision against its base and intended outcome, even in a fresh conversation. If an earlier finding is available, check its original trigger against the current code and report what happens now. Inspect the fix for effects on other callers.

If you know both the old and new commit ranges, `git range-diff OLD_BASE..OLD_HEAD NEW_BASE..NEW_HEAD` can focus attention after a rebase or rewrite. Inspect relevant base changes as well. Claim a particular defect was fixed only when you can establish the original failure and its current outcome.
