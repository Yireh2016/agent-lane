import { describe, expect, it } from 'vitest';
import { scanAndWait } from './client.js';

const checkIds = (fs: { checkId: string }[]) => fs.map((f) => f.checkId);

describe('HSTS max-age e2e (real API, worker, Postgres and test sites)', () => {
  it('should not report weak HSTS on goodsite, whose header has a one-year max-age', async () => {
    const { findings } = await scanAndWait('goodsite');
    expect(checkIds(findings)).not.toContain('headers.weak_hsts');
  });

  it('should report only missing HSTS, not weak HSTS, on badsite, which sends no header', async () => {
    const { findings } = await scanAndWait('badsite');
    const ids = checkIds(findings);
    expect(ids).toContain('headers.missing_hsts');
    expect(ids).not.toContain('headers.weak_hsts');
  });
});
