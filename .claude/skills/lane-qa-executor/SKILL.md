---
name: lane-qa-executor
description: Use on "/lane-qa-executor <PR>", "validate PR <n> locally", or when the lane-orchestrator's prompt asks for a scheduled validation run of an agent-lane PR.
---

# Lane QA executor

You run a PR's **Validation plan** against the real stack and report what actually happened.

| Mode | Started by | Runs | Reports to |
|---|---|---|---|
| `scheduled` | The orchestrator | `shell`, and `stack` steps when the stack starts | The orchestrator (you post nothing) |
| `local` (default) | A person on their computer | Every step, `human` ones with the person | The PR and `lane/memory` |

## Rules

- Never merge, approve or enable auto-merge, even when everything passes.
- Run only the plan's steps. Other things you notice go under "Other findings".
- Local stack only. A step that would reach a hosted environment, a real third party, or print a
  secret is `blocked`, whatever it says.
- **Expect is fixed.** If the result differs, the step fails. You may fix a bug in a step's own
  script and say so, never the expected result.
- A step passes only if you ran it and saw the result. Reading code is not validation.

## Run

1. Load the plan (scheduled: from the prompt; local: from the PR body, after checking its first
   line is a `lane:pr` marker). Check out each branch at the given SHA.
2. `stack` steps: `scripts/stack.sh start` once. If it fails or times out, every `stack` step is
   `deferred` with `cloud fallback: <last error line>`; run `scripts/stack.sh stop`.
3. For each step: do exactly what it says, compare with Expect, record `pass`, `fail` (what you saw
   instead), `blocked` (why) or `deferred`, with the command and up to 10 lines of evidence.
4. A script that crashes or can't reach the thing it checks before the comparison: retry once, then
   `deferred` with `cloud fallback:`. A false `fail` sends an implementer to "fix" working code; a
   deferral only costs a local run.
5. Delete `.claude-lane-qa/` and `scripts/stack.sh stop` if you started it.

**Scheduled:** return the executor report. Never fix code.

**Local:** post one `lane:validation run=local` comment on the PR. Update the item in
`lane/memory` (`Validation: pass|fail|partial (date, sha)`) and append a ledger line, commit
`chore(lane): validation PR #<n> <result>`, push with rebase-retry. For a small obvious bug, ask
the person before fixing it on the PR branch.
