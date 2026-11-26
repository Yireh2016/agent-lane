package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/agent-lane/worker/internal/checks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Job struct {
	ScanID   int64
	DomainID int64
	Hostname string
	Port     int
}

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// ClaimNext marks the oldest queued scan as running and returns it, or nil when none is queued.
func (s *Store) ClaimNext(ctx context.Context) (*Job, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE scans SET status = 'running', started_at = now()
		WHERE id = (
			SELECT id FROM scans WHERE status = 'queued'
			ORDER BY id FOR UPDATE SKIP LOCKED LIMIT 1
		)
		RETURNING id, domain_id,
			(SELECT hostname FROM domains WHERE domains.id = scans.domain_id),
			(SELECT port FROM domains WHERE domains.id = scans.domain_id)`)
	var j Job
	if err := row.Scan(&j.ScanID, &j.DomainID, &j.Hostname, &j.Port); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &j, nil
}

func (s *Store) Complete(ctx context.Context, job Job, findings []checks.Finding) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, f := range findings {
		detail, err := json.Marshal(f.Detail)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO findings (scan_id, domain_id, check_id, severity, title, detail)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			job.ScanID, job.DomainID, f.CheckID, string(f.Severity), f.Title, detail); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE scans SET status = 'done', finished_at = now() WHERE id = $1`, job.ScanID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) Fail(ctx context.Context, job Job, cause error) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE scans SET status = 'failed', error = $2, finished_at = now() WHERE id = $1`,
		job.ScanID, cause.Error())
	return err
}
