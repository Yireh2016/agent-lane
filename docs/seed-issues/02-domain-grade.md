---
title: Add a letter grade per domain
labels: [agent-lane]
lang: typescript
purpose: happy path, API only
---

Customers want a single grade per domain, computed from its latest completed scan.

**Acceptance criteria**
- `GET /domains/:id/grade` returns `{ domainId, scanId, grade }`.
- Grade: `A` with no findings, `B` when the worst finding is `low`, `C` for `medium`, `D` for `high`, `F` for `critical`.
- `404` with `{ error }` when the domain doesn't exist or has no completed scan.
- Requires the API key like every other endpoint.
- In the e2e stack, `badsite` grades `D` and `goodsite` grades `A`.
