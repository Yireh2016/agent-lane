import { describe, expect, it } from 'vitest';
import { call, scanAndWait } from './client.js';

type Finding = { checkId: string; severity: string; detail: { acceptedVersions?: string[] } };

describe('legacy TLS protocols e2e (real worker against nginx test sites)', () => {
  it('should report legacysite as accepting TLS 1.0 and 1.1 with severity high', async () => {
    const { domainId } = await scanAndWait('legacysite');
    const { body } = await call<{ findings: Finding[] }>('GET', `/domains/${domainId}/findings`);
    expect(body.findings.map((f) => f.checkId)).toEqual(['tls.legacy_protocol']);
    expect(body.findings[0]?.severity).toBe('high');
    expect(body.findings[0]?.detail.acceptedVersions).toEqual(['TLS 1.0', 'TLS 1.1']);
  });

  it('should report no findings on goodsite, which only accepts TLS 1.2 and 1.3', async () => {
    const { findings } = await scanAndWait('goodsite');
    expect(findings).toEqual([]);
  });
});
