package main

import (
	"context"
	"crypto/x509"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agent-lane/worker/internal/checks"
	"github.com/agent-lane/worker/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Error("connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	roots, err := loadRoots(os.Getenv("EXTRA_CA_FILE"))
	if err != nil {
		log.Error("load CA", "err", err)
		os.Exit(1)
	}

	st := store.New(pool)
	log.Info("worker started")
	for ctx.Err() == nil {
		job, err := st.ClaimNext(ctx)
		if err != nil {
			log.Error("claim", "err", err)
			sleep(ctx, 2*time.Second)
			continue
		}
		if job == nil {
			sleep(ctx, 500*time.Millisecond)
			continue
		}
		runJob(ctx, log, st, *job, roots)
	}
}

func runJob(ctx context.Context, log *slog.Logger, st *store.Store, job store.Job, roots *x509.CertPool) {
	jobCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	log = log.With("scan", job.ScanID, "host", job.Hostname, "port", job.Port)

	tlsFindings, tlsErr := checks.CheckTLS(jobCtx, job.Hostname, job.Port, checks.TLSOptions{Roots: roots})
	headerFindings, headerErr := checks.CheckHeaders(jobCtx, job.Hostname, job.Port, checks.HeaderOptions{Roots: roots})
	if err := errors.Join(tlsErr, headerErr); err != nil {
		log.Warn("scan failed", "err", err)
		if ferr := st.Fail(ctx, job, err); ferr != nil {
			log.Error("record failure", "err", ferr)
		}
		return
	}
	findings := append(tlsFindings, headerFindings...)
	findings = append(findings, checks.CheckLegacyProtocols(jobCtx, job.Hostname, job.Port, checks.LegacyOptions{})...)
	if err := st.Complete(ctx, job, findings); err != nil {
		log.Error("record findings", "err", err)
		return
	}
	log.Info("scan done", "findings", len(findings))
}

func loadRoots(extraCAFile string) (*x509.CertPool, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}
	if extraCAFile == "" {
		return pool, nil
	}
	pem, err := os.ReadFile(extraCAFile)
	if err != nil {
		return nil, err
	}
	if !pool.AppendCertsFromPEM(pem) {
		return nil, errors.New("no certificates found in " + extraCAFile)
	}
	return pool, nil
}

func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
