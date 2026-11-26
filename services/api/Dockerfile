# syntax=docker/dockerfile:1
FROM node:22-alpine AS build
WORKDIR /repo
COPY tsconfig.base.json ./
COPY services/api/package.json services/api/tsconfig.json services/api/tsconfig.build.json services/api/
WORKDIR /repo/services/api
# Reason: the optional proxy_ca secret lets builds run behind a TLS-intercepting proxy
# (see docker-compose.proxy.yml).
RUN --mount=type=secret,id=proxy_ca,required=false \
    if [ -f /run/secrets/proxy_ca ]; then export NODE_EXTRA_CA_CERTS=/run/secrets/proxy_ca; fi; \
    npm install --no-audit --no-fund
COPY services/api/src ./src
RUN npm run build && npm prune --omit=dev

FROM node:22-alpine
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /repo/services/api/node_modules ./node_modules
COPY --from=build /repo/services/api/dist ./dist
COPY services/api/package.json ./
USER node
EXPOSE 8080
CMD ["node", "dist/server.js"]
