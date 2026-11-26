import { describe, expect, it } from 'vitest';
import { API, call, scanAndWait } from './client.js';

type GradeBody = { domainId: number; scanId: number; grade: string };

// Reason: other e2e files scan the same hostnames in parallel, so the latest completed scan
// can be newer than ours; it must still be a completed scan of this domain.
async function expectGrade(hostname: string, grade: string) {
  const { domainId, scanId } = await scanAndWait(hostname);
  const res = await call<GradeBody>('GET', `/domains/${domainId}/grade`);
  expect(res.status).toBe(200);
  expect(Object.keys(res.body).sort()).toEqual(['domainId', 'grade', 'scanId']);
  expect(res.body).toMatchObject({ domainId, grade });
  expect(res.body.scanId).toBeGreaterThanOrEqual(scanId);
  const scan = await call<{ domainId: number; status: string }>('GET', `/scans/${res.body.scanId}`);
  expect(scan.body).toMatchObject({ domainId, status: 'done' });
}

describe('grade e2e (real API, worker, Postgres and test sites)', () => {
  it('should grade badsite D from its latest completed scan', async () => {
    await expectGrade('badsite', 'D');
  });

  it('should grade goodsite A', async () => {
    await expectGrade('goodsite', 'A');
  });

  it('should 404 with an error for an unknown domain or one without a completed scan', async () => {
    const unknown = await call<{ error: string }>('GET', '/domains/999999/grade');
    expect(unknown.status).toBe(404);
    expect(typeof unknown.body.error).toBe('string');
    const domain = await call<{ id: number }>('POST', '/domains', { hostname: 'never-scanned.test' });
    const noScan = await call<{ error: string }>('GET', `/domains/${domain.body.id}/grade`);
    expect(noScan.status).toBe(404);
    expect(typeof noScan.body.error).toBe('string');
  });

  it('should reject a grade request without the API key', async () => {
    const res = await fetch(`${API}/domains/1/grade`);
    expect(res.status).toBe(401);
  });
});
