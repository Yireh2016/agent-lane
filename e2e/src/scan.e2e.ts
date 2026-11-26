import { describe, expect, it } from 'vitest';
import { call, scanAndWait } from './client.js';

const bySeverity = (fs: { checkId: string; severity: string }[]) =>
  Object.fromEntries(fs.map((f) => [f.checkId, f.severity]));

describe('scan e2e (real API, worker, Postgres and test sites)', () => {
  it('should report the expired, untrusted cert and every missing header on badsite', async () => {
    const { findings } = await scanAndWait('badsite');
    expect(bySeverity(findings)).toEqual({
      'tls.certificate_expired': 'high',
      'tls.certificate_untrusted': 'high',
      'headers.missing_hsts': 'medium',
      'headers.missing_csp': 'medium',
      'headers.missing_x_content_type_options': 'low',
      'headers.missing_x_frame_options': 'low',
    });
  });

  it('should report nothing on goodsite', async () => {
    const { findings } = await scanAndWait('goodsite');
    expect(findings).toEqual([]);
  });

  it('should mark a scan failed when the host does not answer', async () => {
    await expect(scanAndWait('nothing-here')).rejects.toThrow(/scan failed/);
  });

  it('should reject calls without the API key', async () => {
    const res = await fetch(`http://127.0.0.1:${process.env.LANE_API_PORT ?? 8080}/domains/1`);
    expect(res.status).toBe(401);
    expect((await call('GET', '/health')).status).toBe(200);
  });
});

describe('network isolation', () => {
  it('should not let the worker reach the public internet', async () => {
    await expect(scanAndWait('example.com')).rejects.toThrow(/scan failed/);
  });
});
