import type { Domain, Finding, Repository, Scan } from './repository.js';

// In-memory Repository for unit tests. Behaves like PgRepository, including the
// (hostname, port) uniqueness rule, so tests exercise real request handling.
export class MemoryRepository implements Repository {
  domains: Domain[] = [];
  scans: Scan[] = [];
  findings: Finding[] = [];

  async createDomain(hostname: string, port: number) {
    const existing = this.domains.find((d) => d.hostname === hostname && d.port === port);
    if (existing) return { domain: existing, created: false };
    const domain = { id: this.domains.length + 1, hostname, port, createdAt: new Date().toISOString() };
    this.domains.push(domain);
    return { domain, created: true };
  }

  async getDomain(id: number) {
    return this.domains.find((d) => d.id === id) ?? null;
  }

  async createScan(domainId: number) {
    const scan: Scan = {
      id: this.scans.length + 1,
      domainId,
      status: 'queued',
      error: null,
      createdAt: new Date().toISOString(),
      finishedAt: null,
    };
    this.scans.push(scan);
    return scan;
  }

  async getScan(id: number) {
    return this.scans.find((s) => s.id === id) ?? null;
  }

  async latestDoneScan(domainId: number) {
    return this.scans.filter((s) => s.domainId === domainId && s.status === 'done').at(-1) ?? null;
  }

  async findingsForScan(scanId: number) {
    return this.findings.filter((f) => f.scanId === scanId);
  }
}
