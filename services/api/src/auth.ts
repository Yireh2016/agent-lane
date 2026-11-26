import { timingSafeEqual } from 'node:crypto';
import type { FastifyReply, FastifyRequest } from 'fastify';

export function apiKeyGuard(expected: string) {
  const expectedBuf = Buffer.from(expected);
  return async (req: FastifyRequest, reply: FastifyReply) => {
    const given = req.headers['x-api-key'];
    const givenBuf = Buffer.from(typeof given === 'string' ? given : '');
    if (givenBuf.length !== expectedBuf.length || !timingSafeEqual(givenBuf, expectedBuf)) {
      return reply.code(401).send({ error: 'unauthorized' });
    }
  };
}
