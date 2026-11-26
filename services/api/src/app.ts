import Fastify from 'fastify';
import { apiKeyGuard } from './auth.js';
import { gradeFor } from './grade.js';
import { normalizeHostname } from './hostname.js';
import type { Repository } from './repository.js';

export interface AppOptions {
  repo: Repository;
  apiKey: string;
  logger?: boolean;
}

const parseId = (raw: unknown): number | null => {
  const n = Number(raw);
  return Number.isSafeInteger(n) && n > 0 ? n : null;
};

export function buildApp({ repo, apiKey, logger = false }: AppOptions) {
  const app = Fastify({ logger });

  app.get('/health', async () => ({ status: 'ok' }));

  app.register(async (api) => {
    api.addHook('onRequest', apiKeyGuard(apiKey));

    api.post('/domains', async (req, reply) => {
      const body = (req.body ?? {}) as { hostname?: unknown; port?: unknown };
      const host = normalizeHostname(body.hostname);
      if (!host.ok) return reply.code(400).send({ error: host.reason });
      const port = body.port === undefined ? 443 : Number(body.port);
      if (!Number.isInteger(port) || port < 1 || port > 65535) {
        return reply.code(400).send({ error: 'port must be an integer from 1 to 65535' });
      }
      const { domain, created } = await repo.createDomain(host.hostname, port);
      return reply.code(created ? 201 : 200).send(domain);
    });

    api.get('/domains/:id', async (req, reply) => {
      const id = parseId((req.params as { id: string }).id);
      const domain = id ? await repo.getDomain(id) : null;
      if (!domain) return reply.code(404).send({ error: 'domain not found' });
      return domain;
    });

    api.post('/domains/:id/scans', async (req, reply) => {
      const id = parseId((req.params as { id: string }).id);
      const domain = id ? await repo.getDomain(id) : null;
      if (!domain) return reply.code(404).send({ error: 'domain not found' });
      const scan = await repo.createScan(domain.id);
      return reply.code(202).send(scan);
    });

    api.get('/scans/:id', async (req, reply) => {
      const id = parseId((req.params as { id: string }).id);
      const scan = id ? await repo.getScan(id) : null;
      if (!scan) return reply.code(404).send({ error: 'scan not found' });
      return scan;
    });

    api.get('/domains/:id/findings', async (req, reply) => {
      const id = parseId((req.params as { id: string }).id);
      const domain = id ? await repo.getDomain(id) : null;
      if (!domain) return reply.code(404).send({ error: 'domain not found' });
      const scan = await repo.latestDoneScan(domain.id);
      if (!scan) return reply.code(404).send({ error: 'no completed scan for this domain' });
      const findings = await repo.findingsForScan(scan.id);
      return { domainId: domain.id, scanId: scan.id, findings };
    });

    api.get('/domains/:id/grade', async (req, reply) => {
      const id = parseId((req.params as { id: string }).id);
      const domain = id ? await repo.getDomain(id) : null;
      if (!domain) return reply.code(404).send({ error: 'domain not found' });
      const scan = await repo.latestDoneScan(domain.id);
      if (!scan) return reply.code(404).send({ error: 'no completed scan for this domain' });
      const findings = await repo.findingsForScan(scan.id);
      return { domainId: domain.id, scanId: scan.id, grade: gradeFor(findings) };
    });
  });

  return app;
}
