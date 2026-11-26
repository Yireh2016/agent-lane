package checks

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/agent-lane/worker/internal/certgen"
)

type site struct {
	host  string
	port  int
	roots *x509.CertPool
}

func startSite(t *testing.T, pair certgen.Pair, roots *x509.CertPool, headers map[string]string) site {
	t.Helper()
	cert, err := pair.TLSCertificate()
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(http.StatusOK)
	}))
	srv.TLS = &tls.Config{Certificates: []tls.Certificate{cert}}
	srv.StartTLS()
	t.Cleanup(srv.Close)
	_, portStr, _ := net.SplitHostPort(srv.Listener.Addr().String())
	port, _ := strconv.Atoi(portStr)
	return site{host: "localhost", port: port, roots: roots}
}

func newCA(t *testing.T) (certgen.Pair, *x509.CertPool, func(notBefore, notAfter time.Time) certgen.Pair) {
	t.Helper()
	now := time.Now()
	ca, key, err := certgen.NewCA("test CA", now.Add(-48*time.Hour), now.AddDate(2, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(ca.Cert)
	issue := func(nb, na time.Time) certgen.Pair {
		p, err := certgen.Leaf([]string{"localhost"}, nb, na, ca.Cert, key)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	return ca, pool, issue
}

func ids(fs []Finding) map[string]Severity {
	m := map[string]Severity{}
	for _, f := range fs {
		m[f.CheckID] = f.Severity
	}
	return m
}

func TestCheckTLS(t *testing.T) {
	now := time.Now()
	_, roots, issue := newCA(t)

	t.Run("should report nothing for a trusted cert valid for months", func(t *testing.T) {
		s := startSite(t, issue(now.Add(-time.Hour), now.AddDate(0, 6, 0)), roots, nil)
		fs, err := CheckTLS(context.Background(), s.host, s.port, TLSOptions{Roots: roots})
		if err != nil {
			t.Fatal(err)
		}
		if len(fs) != 0 {
			t.Fatalf("expected no findings, got %v", ids(fs))
		}
	})

	t.Run("should report expiring soon as medium", func(t *testing.T) {
		s := startSite(t, issue(now.Add(-time.Hour), now.Add(10*24*time.Hour)), roots, nil)
		fs, _ := CheckTLS(context.Background(), s.host, s.port, TLSOptions{Roots: roots})
		got := ids(fs)
		if got["tls.certificate_expiring_soon"] != SeverityMedium || len(got) != 1 {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("should report an expired cert as high but not as untrusted", func(t *testing.T) {
		s := startSite(t, issue(now.AddDate(-1, 0, 0), now.Add(-time.Hour)), roots, nil)
		fs, _ := CheckTLS(context.Background(), s.host, s.port, TLSOptions{Roots: roots})
		got := ids(fs)
		if got["tls.certificate_expired"] != SeverityHigh || len(got) != 1 {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("should report a self-signed cert as untrusted", func(t *testing.T) {
		self, err := certgen.Leaf([]string{"localhost"}, now.Add(-time.Hour), now.AddDate(0, 6, 0), nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		s := startSite(t, self, roots, nil)
		fs, _ := CheckTLS(context.Background(), s.host, s.port, TLSOptions{Roots: roots})
		got := ids(fs)
		if got["tls.certificate_untrusted"] != SeverityHigh || len(got) != 1 {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("should return an error when nothing listens", func(t *testing.T) {
		_, err := CheckTLS(context.Background(), "127.0.0.1", 1, TLSOptions{Timeout: time.Second})
		if err == nil {
			t.Fatal("expected an error")
		}
	})
}

func TestCheckHeaders(t *testing.T) {
	now := time.Now()
	_, roots, issue := newCA(t)
	pair := issue(now.Add(-time.Hour), now.AddDate(0, 6, 0))

	t.Run("should report every missing header with its severity", func(t *testing.T) {
		s := startSite(t, pair, roots, nil)
		fs, err := CheckHeaders(context.Background(), s.host, s.port, HeaderOptions{Roots: roots})
		if err != nil {
			t.Fatal(err)
		}
		got := ids(fs)
		want := map[string]Severity{
			"headers.missing_hsts":                   SeverityMedium,
			"headers.missing_csp":                    SeverityMedium,
			"headers.missing_x_content_type_options": SeverityLow,
			"headers.missing_x_frame_options":        SeverityLow,
		}
		if len(got) != len(want) {
			t.Fatalf("got %v", got)
		}
		for k, v := range want {
			if got[k] != v {
				t.Fatalf("%s: got %q want %q", k, got[k], v)
			}
		}
	})

	t.Run("should accept CSP frame-ancestors instead of X-Frame-Options", func(t *testing.T) {
		s := startSite(t, pair, roots, map[string]string{
			"Strict-Transport-Security": "max-age=31536000",
			"Content-Security-Policy":   "default-src 'self'; frame-ancestors 'none'",
			"X-Content-Type-Options":    "nosniff",
		})
		fs, _ := CheckHeaders(context.Background(), s.host, s.port, HeaderOptions{Roots: roots})
		if len(fs) != 0 {
			t.Fatalf("expected no findings, got %v", ids(fs))
		}
	})
}
