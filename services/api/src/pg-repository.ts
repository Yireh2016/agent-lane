import type pg from 'pg';
import type { Domain, Finding, Repository, Scan } from './repository.js';

type Row = Record<string, unknown>;

const toDomain = (r: Row): Domain => ({
  id: Number(r.id),
  hostname: String(r.hostname),
  port: Number(r.port),
  createdAt: (r.created_at as Date).toISOString(),
});

const toScan = (r: Row): Scan => ({
  id: Number(r.id),
  domainId: Number(r.domain_id),
  status: r.status as Scan['status'],
  error: (r.error as string | null) ?? null,
  createdAt: (r.created_at as Date).toISOString(),
  finishedAt: r.finished_at ? (r.finished_at as Date).toISOString() : null,
});

const toFinding = (r: Row): Finding => ({
  id: Number(r.id),
  scanId: Number(r.scan_id),
  checkId: String(r.check_id),
  severity: r.severity as Finding['severity'],
  title: String(r.title),
  detail: r.detail as Record<string, unknown>,
});

export class PgRepository implements Repository {
  constructor(private readonly pool: pg.Pool) {}

  async createDomain(hostname: string, port: number) {
    const inserted = await this.pool.query(
      `INSERT INTO domains (hostname, port) VALUES ($1, $2)
       ON CONFLICT (hostname, port) DO NOTHING RETURNING *`,
      [hostname, port],
    );
    if (inserted.rows[0]) return { domain: toDomain(inserted.rows[0]), created: true };
    const existing = await this.pool.query('SELECT * FROM domains WHERE hostname = $1 AND port = $2', [
      hostname,
      port,
    ]);
    return { domain: toDomain(existing.rows[0]), created: false };
  }

  async getDomain(id: number) {
    const res = await this.pool.query('SELECT * FROM domains WHERE id = $1', [id]);
    return res.rows[0] ? toDomain(res.rows[0]) : null;
  }

  async createScan(domainId: number) {
    const res = await this.pool.query('INSERT INTO scans (domain_id) VALUES ($1) RETURNING *', [domainId]);
    return toScan(res.rows[0]);
  }

  async getScan(id: number) {
    const res = await this.pool.query('SELECT * FROM scans WHERE id = $1', [id]);
    return res.rows[0] ? toScan(res.rows[0]) : null;
  }

  async latestDoneScan(domainId: number) {
    const res = await this.pool.query(
      `SELECT * FROM scans WHERE domain_id = $1 AND status = 'done' ORDER BY id DESC LIMIT 1`,
      [domainId],
    );
    return res.rows[0] ? toScan(res.rows[0]) : null;
  }

  async findingsForScan(scanId: number) {
    const res = await this.pool.query('SELECT * FROM findings WHERE scan_id = $1 ORDER BY id', [scanId]);
    return res.rows.map(toFinding);
  }
}
