# CLAUDE.md

Agent instructions only. Human docs: `README.md`.

## Layout

- `services/api`: TypeScript (Fastify, pg). Domains, scans, findings. API key on every route but `/health`.
- `services/worker`: Go. Claims queued scans, runs TLS and header checks, writes findings.
- `db/migrations`: SQL applied by Postgres on first start of the stack.
- `testsites/`: nginx sites the stack scans (`badsite`: expired self-signed cert, no headers; `goodsite`: valid cert from the local CA, all headers).
- `e2e/`: tests against the running stack.
- `tools/gates/` + `.lane/policy.json`: policy gates for agent changes.
- `.claude/skills/lane-*`: the agent lane (orchestrator, implement, qa-writer, qa-executor).

## Commands

```bash
pnpm verify                 # lint, typecheck, unit tests (TS, Go), gate tests: run before calling work done
scripts/stack.sh start      # Postgres, API :8080, worker, test sites; returns when healthy
pnpm e2e                    # e2e against the stack
scripts/stack.sh stop
node tools/gates/check.mjs  # gates vs origin/main: exit 0 pass, 1 fail, 3 needs a human
```

## Conventions

- Tests start with "should" and test behaviour, not mocks. Go tests use real TLS servers (`internal/certgen`).
- A runtime behaviour change gets an e2e test in `e2e/src/`.
- New finding ids are `<area>.<snake_case>`; severities are `low | medium | high | critical`.
- Never weaken a test, check or gate to make it pass. Never edit `.lane/policy.json` or `tools/gates/` in an issue's PR.
- The stack's worker network has no internet access. Don't add workarounds for that.
