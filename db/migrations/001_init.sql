CREATE TABLE domains (
  id          BIGSERIAL PRIMARY KEY,
  hostname    TEXT NOT NULL,
  port        INTEGER NOT NULL DEFAULT 443 CHECK (port BETWEEN 1 AND 65535),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (hostname, port)
);

CREATE TABLE scans (
  id           BIGSERIAL PRIMARY KEY,
  domain_id    BIGINT NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
  status       TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'running', 'done', 'failed')),
  error        TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  started_at   TIMESTAMPTZ,
  finished_at  TIMESTAMPTZ
);

CREATE INDEX scans_queued_idx ON scans (id) WHERE status = 'queued';

CREATE TABLE findings (
  id          BIGSERIAL PRIMARY KEY,
  scan_id     BIGINT NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
  domain_id   BIGINT NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
  check_id    TEXT NOT NULL,
  severity    TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
  title       TEXT NOT NULL,
  detail      JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX findings_scan_idx ON findings (scan_id);
