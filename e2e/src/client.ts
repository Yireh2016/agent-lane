export const API = `http://127.0.0.1:${process.env.LANE_API_PORT ?? 8080}`;
const KEY = process.env.LANE_API_KEY ?? 'local-dev-key';

export async function call<T = unknown>(method: string, path: string, body?: unknown) {
  const res = await fetch(API + path, {
    method,
    headers: { 'x-api-key': KEY, ...(body ? { 'content-type': 'application/json' } : {}) },
    body: body ? JSON.stringify(body) : undefined,
  });
  const text = await res.text();
  return { status: res.status, body: (text ? JSON.parse(text) : null) as T };
}

export async function scanAndWait(hostname: string) {
  const domain = await call<{ id: number }>('POST', '/domains', { hostname });
  const scan = await call<{ id: number }>('POST', `/domains/${domain.body.id}/scans`);
  const deadline = Date.now() + 30_000;
  while (Date.now() < deadline) {
    const s = await call<{ status: string; error: string | null }>('GET', `/scans/${scan.body.id}`);
    if (s.body.status === 'done') {
      const f = await call<{ findings: { checkId: string; severity: string }[] }>(
        'GET',
        `/domains/${domain.body.id}/findings`,
      );
      return { domainId: domain.body.id, scanId: scan.body.id, findings: f.body.findings };
    }
    if (s.body.status === 'failed') throw new Error(`scan failed: ${s.body.error}`);
    await new Promise((r) => setTimeout(r, 300));
  }
  throw new Error(`scan ${scan.body.id} did not finish in 30s`);
}
