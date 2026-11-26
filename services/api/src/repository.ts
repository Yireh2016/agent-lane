export type Severity = 'low' | 'medium' | 'high' | 'critical';
export type ScanStatus = 'queued' | 'running' | 'done' | 'failed';

export interface Domain {
  id: number;
  hostname: string;
  port: number;
  createdAt: string;
}

export interface Scan {
  id: number;
  domainId: number;
  status: ScanStatus;
  error: string | null;
  createdAt: string;
  finishedAt: string | null;
}

export interface Finding {
  id: number;
  scanId: number;
  checkId: string;
  severity: Severity;
  title: string;
  detail: Record<string, unknown>;
}

export interface Repository {
  createDomain(hostname: string, port: number): Promise<{ domain: Domain; created: boolean }>;
  getDomain(id: number): Promise<Domain | null>;
  createScan(domainId: number): Promise<Scan>;
  getScan(id: number): Promise<Scan | null>;
  latestDoneScan(domainId: number): Promise<Scan | null>;
  findingsForScan(scanId: number): Promise<Finding[]>;
}
