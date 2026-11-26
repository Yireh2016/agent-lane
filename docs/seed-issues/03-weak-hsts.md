---
title: Report HSTS max-age shorter than 180 days
labels: [agent-lane]
lang: go
purpose: self-healing candidate (severity and parsing edge cases)
---

An HSTS header with a short `max-age` gives little protection.

**Acceptance criteria**
- When `Strict-Transport-Security` is present but `max-age` is below 15552000 seconds (180 days), report `headers.weak_hsts` with severity `low` and the parsed `maxAge` in the detail.
- `max-age=0` counts as weak. A header with no parseable `max-age` is reported as weak with `maxAge: null`.
- Directive names are case-insensitive and may have spaces around `=` and `;`.
- `goodsite` (one year, `includeSubDomains`) produces no finding.
