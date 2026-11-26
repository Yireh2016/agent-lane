# Lane formats

Every GitHub post ends with:

```
---
_Posted by the agent lane · a person reviews and merges_
```

## Markers (first line of a post, HTML comment)

| Marker | Where |
|---|---|
| `<!-- lane:pr issue=N item=lane/N-slug -->` | First line of every lane PR body |
| `<!-- lane:issue-review issue=N -->` | Gap review on an issue |
| `<!-- lane:question id=Q3 blocking=true -->` | One question, on the PR or the issue |
| `<!-- lane:validation run=scheduled\|local sha=abc1234 -->` | Validation report |

## PR body

```
<!-- lane:pr issue=N item=lane/N-slug -->
## Summary
One paragraph. `Closes #N` only when every requirement is met; otherwise `Part of #N`.

## Decisions
- D1: choice made where the issue allowed several, and why.

## Assumptions
- A1: what was taken as true without confirmation.

## Open questions
- Q1 (blocking): question · link to its comment

## Verification
`pnpm verify`: pass/fail with the failing command.

## Gates
| Gate | Result | Detail |
|---|---|---|

## Requirements coverage
| Req | Status | Where (path:line) | Proven by |
|---|---|---|---|

## Validation plan
(steps, format below)

## Validation results
Link to the latest `lane:validation` comment.
```

## Validation step

```
### V1 · R2 · Needs: shell | stack | browser | human
**Do:** exact commands, one plain command per line, or a Script block.
**Expect:** the exact observable result. Never edited to match what the code does.
```

`stack` steps need `scripts/stack.sh start` (Postgres, API, worker, badsite, goodsite).

## Implementer report

```
STATUS: done | partial | blocked
BRANCH: lane/N-slug (and stacked branches, bottom first)
REQUIREMENTS: R1 met (path:line) · R2 partial (what is missing) …
REMAINING: what is left, if partial
DECISIONS / ASSUMPTIONS / QUESTIONS: lists
VERIFY: commands run and results
VALIDATION_HINTS: how a reviewer could observe each requirement
```

## QA executor report

```
SHA: abc1234
V1 pass · evidence (command + up to 10 lines)
V2 fail · expected X, saw Y · evidence
V3 deferred · cloud fallback: reason
```

## Ledger line (`ledger.jsonl` on `lane/memory`, one JSON object per item per run)

```json
{"date":"2026-12-01","issue":12,"pr":34,"mode":"new","lang":["go"],
 "implementer_agents":2,"heal_rounds":{"self_check":1,"audit":0,"validation":1},
 "first_pass_validation":"fail","final_validation":"pass",
 "gates":{"protected-paths":"pass","diff-size":"pass","dependencies":"pass","secrets":"pass"},
 "requirements":{"met":3,"partial":0,"missing":0},
 "injection_suspected":false,"deferred_steps":0,
 "tokens_in":0,"tokens_out":0,"cost_usd":0,"wall_minutes":0,
 "failure_categories":["wrong-severity"]}
```

`failure_categories` uses a short fixed vocabulary where possible: `misread-requirement`,
`edited-tests-to-pass`, `wrong-file`, `missing-edge-case`, `flaky-env`, `scope-creep`,
`dependency-added`, `protected-path`, `injection`, `wrong-severity`, `other:<word>`.
