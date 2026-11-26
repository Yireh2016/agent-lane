import { beforeEach, describe, expect, it } from 'vitest';
import { buildApp } from './app.js';
import { MemoryRepository } from './memory-repository.js';
import type { Severity } from './repository.js';

const KEY = 'test-key';
const auth = { 'x-api-key': KEY };

describe('api', () => {
  let repo: MemoryRepository;
  let app: ReturnType<typeof buildApp>;

  beforeEach(() => {
    repo = new MemoryRepository();
    app = buildApp({ repo, apiKey: KEY });
  });

  it('should serve health without an API key', async () => {
    const res = await app.inject({ method: 'GET', url: '/health' });
    expect(res.statusCode).toBe(200);
  });

  it('should reject requests with a missing or wrong API key', async () => {
    const missing = await app.inject({ method: 'GET', url: '/domains/1' });
    const wrong = await app.inject({ method: 'GET', url: '/domains/1', headers: { 'x-api-key': 'nope' } });
    expect(missing.statusCode).toBe(401);
    expect(wrong.statusCode).toBe(401);
  });

  it('should create a domain once and return the existing one after', async () => {
    const first = await app.inject({
      method: 'POST',
      url: '/domains',
      headers: auth,
      payload: { hostname: 'Example.com' },
    });
    const again = await app.inject({
      method: 'POST',
      url: '/domains',
      headers: auth,
      payload: { hostname: 'example.com' },
    });
    expect(first.statusCode).toBe(201);
    expect(again.statusCode).toBe(200);
    expect(again.json().id).toBe(first.json().id);
    expect(first.json()).toMatchObject({ hostname: 'example.com', port: 443 });
  });

  it('should reject an invalid hostname or port', async () => {
    const host = await app.inject({
      method: 'POST',
      url: '/domains',
      headers: auth,
      payload: { hostname: 'http://x' },
    });
    const port = await app.inject({
      method: 'POST',
      url: '/domains',
      headers: auth,
      payload: { hostname: 'x.com', port: 70000 },
    });
    expect(host.statusCode).toBe(400);
    expect(port.statusCode).toBe(400);
  });

  it('should queue a scan for an existing domain and 404 otherwise', async () => {
    await repo.createDomain('example.com', 443);
    const ok = await app.inject({ method: 'POST', url: '/domains/1/scans', headers: auth });
    const missing = await app.inject({ method: 'POST', url: '/domains/99/scans', headers: auth });
    expect(ok.statusCode).toBe(202);
    expect(ok.json()).toMatchObject({ domainId: 1, status: 'queued' });
    expect(missing.statusCode).toBe(404);
  });

  it('should return findings from the latest completed scan only', async () => {
    await repo.createDomain('example.com', 443);
    const old = await repo.createScan(1);
    old.status = 'done';
    const latest = await repo.createScan(1);
    latest.status = 'done';
    await repo.createScan(1);
    repo.findings.push(
      {
        id: 1,
        scanId: old.id,
        checkId: 'tls.certificate_expired',
        severity: 'high',
        title: 'old',
        detail: {},
      },
      {
        id: 2,
        scanId: latest.id,
        checkId: 'headers.missing_hsts',
        severity: 'medium',
        title: 'new',
        detail: {},
      },
    );
    const res = await app.inject({ method: 'GET', url: '/domains/1/findings', headers: auth });
    expect(res.statusCode).toBe(200);
    expect(res.json().scanId).toBe(latest.id);
    expect(res.json().findings.map((f: { checkId: string }) => f.checkId)).toEqual(['headers.missing_hsts']);
  });

  describe('GET /domains/:id/grade', () => {
    const addFinding = (scanId: number, severity: Severity) =>
      repo.findings.push({
        id: repo.findings.length + 1,
        scanId,
        checkId: `test.${severity}`,
        severity,
        title: severity,
        detail: {},
      });

    it('should grade the latest completed scan and ignore older or unfinished ones', async () => {
      await repo.createDomain('example.com', 443);
      const old = await repo.createScan(1);
      old.status = 'done';
      addFinding(old.id, 'critical');
      const latest = await repo.createScan(1);
      latest.status = 'done';
      addFinding(latest.id, 'low');
      addFinding(latest.id, 'medium');
      const running = await repo.createScan(1);
      running.status = 'running';
      addFinding(running.id, 'high');
      const res = await app.inject({ method: 'GET', url: '/domains/1/grade', headers: auth });
      expect(res.statusCode).toBe(200);
      expect(res.json()).toEqual({ domainId: 1, scanId: latest.id, grade: 'C' });
    });

    it('should grade A for a completed scan with no findings', async () => {
      await repo.createDomain('example.com', 443);
      const scan = await repo.createScan(1);
      scan.status = 'done';
      const res = await app.inject({ method: 'GET', url: '/domains/1/grade', headers: auth });
      expect(res.statusCode).toBe(200);
      expect(res.json()).toEqual({ domainId: 1, scanId: scan.id, grade: 'A' });
    });

    it('should 404 with an error for an unknown domain or one without a completed scan', async () => {
      await repo.createDomain('example.com', 443);
      const failed = await repo.createScan(1);
      failed.status = 'failed';
      await repo.createScan(1);
      const unknown = await app.inject({ method: 'GET', url: '/domains/99/grade', headers: auth });
      const invalid = await app.inject({ method: 'GET', url: '/domains/abc/grade', headers: auth });
      const noScan = await app.inject({ method: 'GET', url: '/domains/1/grade', headers: auth });
      for (const res of [unknown, invalid, noScan]) {
        expect(res.statusCode).toBe(404);
        expect(typeof res.json().error).toBe('string');
      }
    });

    it('should require the API key', async () => {
      await repo.createDomain('example.com', 443);
      const scan = await repo.createScan(1);
      scan.status = 'done';
      const missing = await app.inject({ method: 'GET', url: '/domains/1/grade' });
      const wrong = await app.inject({
        method: 'GET',
        url: '/domains/1/grade',
        headers: { 'x-api-key': 'nope' },
      });
      expect(missing.statusCode).toBe(401);
      expect(wrong.statusCode).toBe(401);
    });
  });
});
