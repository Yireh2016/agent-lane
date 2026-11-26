import pg from 'pg';
import { buildApp } from './app.js';
import { PgRepository } from './pg-repository.js';

const required = (name: string) => {
  const value = process.env[name];
  if (!value) throw new Error(`${name} is required`);
  return value;
};

const pool = new pg.Pool({ connectionString: required('DATABASE_URL') });
const app = buildApp({ repo: new PgRepository(pool), apiKey: required('API_KEY'), logger: true });

const port = Number(process.env.PORT ?? 8080);
await app.listen({ host: '0.0.0.0', port });

for (const signal of ['SIGINT', 'SIGTERM'] as const) {
  process.on(signal, async () => {
    await app.close();
    await pool.end();
    process.exit(0);
  });
}
