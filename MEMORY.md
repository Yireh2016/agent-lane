# Agent lane memory

Lives on branch `lane/memory`. Never merged. The orchestrator reads it at the start of every run.

## Notes from humans

<!-- Priorities, corrections to the dependency graph, standing decisions. The lane obeys these. -->

## Active items

### #1 Flag servers that still accept TLS 1.0 or 1.1
- State: in_progress
- Waiting on: orchestrator
- Branch / PR: lane/1-legacy-tls-protocols
- Requirements:
  - R1: a server accepting TLS 1.0 or 1.1 yields `tls.legacy_protocol`, severity `high`, detail lists the accepted legacy versions (missing)
  - R2: a server accepting only TLS 1.2/1.3 yields no `tls.legacy_protocol` (missing)
  - R3: `goodsite` still yields no findings at all (missing)
  - R4: a local test site that accepts TLS 1.0 is added to the stack and an e2e test proves R1 through it (missing)
- Assumptions (posted as I1–I3 on the issue, non-blocking): `detail.acceptedVersions` with "TLS 1.0"/"TLS 1.1"; new `legacysite` in docker-compose.yml (protected path → needs-human); a probe failing for non-protocol reasons yields no finding and doesn't fail the scan
- Questions: none open
- Last checked: 2026-11-26T04:30Z
- Log: 2026-11-26 picked; gap review posted

### #2 Add a letter grade per domain
- State: in_progress
- Waiting on: orchestrator
- Branch / PR: lane/2-domain-grade
- Requirements:
  - R1: `GET /domains/:id/grade` returns `{ domainId, scanId, grade }` (missing)
  - R2: grade A (no findings), B (worst low), C (worst medium), D (worst high), F (worst critical), from the latest completed scan (missing)
  - R3: 404 with `{ error }` for an unknown domain or a domain with no completed scan (missing)
  - R4: the endpoint requires the API key (missing)
  - R5: in the e2e stack, badsite grades D and goodsite grades A (missing)
- Assumptions: none; issue was clear
- Last checked: 2026-11-26T04:30Z
- Log: 2026-11-26 picked

### #3 Report HSTS max-age shorter than 180 days
- State: in_progress
- Waiting on: orchestrator
- Branch / PR: lane/3-weak-hsts
- Requirements:
  - R1: HSTS present with max-age < 15552000 → `headers.weak_hsts`, severity `low`, `detail.maxAge` = parsed value (missing)
  - R2: `max-age=0` is weak (missing)
  - R3: HSTS without a parseable max-age → weak with `maxAge: null` (missing)
  - R4: directive names case-insensitive; spaces around `=` and `;` tolerated (missing)
  - R5: goodsite yields no finding (missing)
- Assumptions (posted as I1–I3, non-blocking): exactly 15552000 is not weak; first max-age wins; missing HSTS gives only `headers.missing_hsts`
- Last checked: 2026-11-26T04:30Z
- Log: 2026-11-26 picked; gap review posted

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
