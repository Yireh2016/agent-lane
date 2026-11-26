# Agent lane memory

Lives on branch `lane/memory`. Never merged. The orchestrator reads it at the start of every run.

## Notes from humans

<!-- Priorities, corrections to the dependency graph, standing decisions. The lane obeys these. -->

## Active items

### #1 Flag servers that still accept TLS 1.0 or 1.1
- State: pr_needs_review
- Waiting on: human:review (needs-human: docker-compose.yml)
- Branch / PR: lane/1-legacy-tls-protocols · #6 (4f3c013)
- Requirements:
  - R1: a server accepting TLS 1.0 or 1.1 yields `tls.legacy_protocol`, severity `high`, detail lists the accepted legacy versions (met)
  - R2: a server accepting only TLS 1.2/1.3 yields no `tls.legacy_protocol` (met)
  - R3: `goodsite` still yields no findings at all (met)
  - R4: a local test site that accepts TLS 1.0 is added to the stack and an e2e test proves R1 through it (met)
- Assumptions (posted as I1–I3 on the issue, non-blocking): `detail.acceptedVersions` with "TLS 1.0"/"TLS 1.1"; new `legacysite` in docker-compose.yml (protected path → needs-human); a probe failing for non-protocol reasons yields no finding and doesn't fail the scan
- Questions: none open
- Last checked: 2026-11-26T04:35Z
- Validation: pass (2026-11-26, sha 4f3c013) 4/4
- Log: 2026-11-26 picked; gap review posted; implemented, audited, validated first pass; PR #6 with needs-human

### #2 Add a letter grade per domain
- State: pr_needs_review
- Waiting on: human:review
- Branch / PR: lane/2-domain-grade · #7 (ea8ac02)
- Requirements:
  - R1: `GET /domains/:id/grade` returns `{ domainId, scanId, grade }` (met)
  - R2: grade A (no findings), B (worst low), C (worst medium), D (worst high), F (worst critical), from the latest completed scan (met)
  - R3: 404 with `{ error }` for an unknown domain or a domain with no completed scan (met)
  - R4: the endpoint requires the API key (met)
  - R5: in the e2e stack, badsite grades D and goodsite grades A (met)
- Assumptions: none; issue was clear
- Last checked: 2026-11-26T04:35Z
- Validation: pass (2026-11-26, sha ea8ac02) 5/5
- Log: 2026-11-26 picked; implemented (own e2e flake from shared DB fixed by the implementer), audited, validated first pass; PR #7

### #3 Report HSTS max-age shorter than 180 days
- State: pr_needs_review
- Waiting on: human:review
- Branch / PR: lane/3-weak-hsts · #8 (a4bdbe0)
- Requirements:
  - R1: HSTS present with max-age < 15552000 → `headers.weak_hsts`, severity `low`, `detail.maxAge` = parsed value (met)
  - R2: `max-age=0` is weak (met)
  - R3: HSTS without a parseable max-age → weak with `maxAge: null` (met)
  - R4: directive names case-insensitive; spaces around `=` and `;` tolerated (met)
  - R5: goodsite yields no finding (met)
- Assumptions (posted as I1–I3, non-blocking): exactly 15552000 is not weak; first max-age wins; missing HSTS gives only `headers.missing_hsts`
- Last checked: 2026-11-26T04:35Z
- Validation: pass (2026-11-26, sha a4bdbe0) 4/4
- Log: 2026-11-26 picked; gap review posted; implemented, audited, validated first pass; PR #8

### #4 Treat Content-Security-Policy-Report-Only as a partial CSP
- State: waiting
- Waiting on: issue:#3 (overlap: services/worker/internal/checks/headers.go)
- Log: 2026-11-26 not started (overlap)

### #5 Let the status page read findings without an API key
- State: waiting
- Waiting on: issue:#2 (overlap: services/api/src/app.ts)
- Log: 2026-11-26 not started (overlap)

<!--
### #N Title
- State: in_progress | pr_needs_review | pr_needs_answers | pr_needs_validation | pr_changes_requested | issue_needs_answers | blocked | failed | merged | closed
- Waiting on: orchestrator | human:answers | human:validation | human:review | issue:#M
- Branch / PR:
- Requirements: R1 … (met | partial | missing | n/a)
- Questions: Q1 … (open | answered)
- Decisions / Assumptions:
- Validation: pass | fail | partial (date, sha)
- Last checked: ISO timestamp
- Log: one line per run
-->

## Dependency graph

| Issue | Depends on | Kind | Reason | Status |
|---|---|---|---|---|
| #1 | none | | | ready |
| #2 | none | | | ready |
| #3 | none | | | ready |
| #4 | #3 | overlap (inferred) | both change headers.go | waiting |
| #5 | #2 | overlap (inferred) | both change app.ts routes | waiting |

## Run log

<!-- ### YYYY-MM-DD: Plan, Results, Problems, Tomorrow -->

### 2026-11-26

**Plan.** First run. New items #1 (Go, legacy TLS, new test site), #2 (TS, grade endpoint), #3 (Go, weak HSTS). They don't share files when each adds its own test and e2e files. #4 waits on #3 and #5 waits on #2 (inferred overlap).

**Results.** 3 PRs opened, all requirements met, all validation steps passed on the first run with no healing rounds: #6 (issue #1, 4/4, needs-human for docker-compose.yml), #7 (issue #2, 5/5), #8 (issue #3, 4/4). CI green on #6 and #7 at publish time.

**Problems.**
- The `needs-human-review` CI job went green on #6 immediately: the `protected-change` environment has no required reviewers yet. A person must add them in repo settings, or the gate is advisory only.
- e2e files run in parallel against one database and share hostnames (badsite, goodsite). Issue #2's implementer hit this (scanId 4 vs 3) and loosened its own assertion. Worth an issue: run e2e files sequentially or isolate state per file.
- PR bodies start with the project attribution lines, so the `lane:pr` marker is on line 3, not line 1. `lane-qa-executor` local mode checks "first line"; it should check "contains".
- README/CLAUDE.md still say two test sites; #6 adds a third.

**Tomorrow.** Sync #6–#8. When #8 merges, #4 (overlaps headers.go) becomes ready; when #7 merges, #5 (overlaps app.ts) becomes ready. Then pick them up.
