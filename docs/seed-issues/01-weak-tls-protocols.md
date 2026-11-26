---
title: Flag servers that still accept TLS 1.0 or 1.1
labels: [agent-lane]
lang: go
purpose: happy path, needs a new test site (stack e2e)
---

Servers that still negotiate TLS 1.0 or 1.1 should be reported.

**Acceptance criteria**
- A scan of a server that accepts TLS 1.0 or 1.1 produces a `tls.legacy_protocol` finding with severity `high`, and the finding's detail lists which legacy versions were accepted.
- A server that only accepts TLS 1.2 and 1.3 produces no `tls.legacy_protocol` finding.
- `goodsite` still produces no findings at all.
- Add a local test site to the stack that accepts TLS 1.0 so the e2e suite proves the finding end to end.
