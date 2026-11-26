const LABEL = /^(?!-)[a-z0-9-]{1,63}(?<!-)$/;

export type HostnameResult = { ok: true; hostname: string } | { ok: false; reason: string };

export function normalizeHostname(input: unknown): HostnameResult {
  if (typeof input !== 'string') return { ok: false, reason: 'hostname must be a string' };
  const hostname = input.trim().toLowerCase().replace(/\.$/, '');
  if (hostname.length === 0 || hostname.length > 253) {
    return { ok: false, reason: 'hostname must be 1 to 253 characters' };
  }
  if (hostname.includes('://') || hostname.includes('/') || hostname.includes(':')) {
    return { ok: false, reason: 'hostname must not include a scheme, port or path' };
  }
  if (!hostname.split('.').every((label) => LABEL.test(label))) {
    return { ok: false, reason: 'hostname has an invalid label' };
  }
  return { ok: true, hostname };
}
