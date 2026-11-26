package checks

import (
	"context"
	"testing"
	"time"
)

func TestWeakHSTS(t *testing.T) {
	now := time.Now()
	_, roots, issue := newCA(t)
	pair := issue(now.Add(-time.Hour), now.AddDate(0, 6, 0))

	scanHSTS := func(t *testing.T, hsts string) []Finding {
		t.Helper()
		headers := map[string]string{
			"Content-Security-Policy": "default-src 'self'; frame-ancestors 'none'",
			"X-Content-Type-Options":  "nosniff",
		}
		if hsts != "" {
			headers["Strict-Transport-Security"] = hsts
		}
		s := startSite(t, pair, roots, headers)
		fs, err := CheckHeaders(context.Background(), s.host, s.port, HeaderOptions{Roots: roots})
		if err != nil {
			t.Fatal(err)
		}
		return fs
	}

	weak := func(fs []Finding) *Finding {
		for i := range fs {
			if fs[i].CheckID == "headers.weak_hsts" {
				return &fs[i]
			}
		}
		return nil
	}

	weakCases := []struct {
		name   string
		header string
		maxAge any
	}{
		{"should report a max-age of one day as weak with its value", "max-age=86400", int64(86400)},
		{"should report a max-age one second under 180 days as weak", "max-age=15551999; includeSubDomains", int64(15551999)},
		{"should report max-age=0 as weak", "max-age=0", int64(0)},
		{"should report a header without max-age as weak with a null maxAge", "includeSubDomains; preload", nil},
		{"should report an unparseable max-age as weak with a null maxAge", "max-age=forever", nil},
		{"should read the directive name case-insensitively", "Max-Age=3600", int64(3600)},
		{"should tolerate spaces around = and ;", "includeSubDomains ;  MAX-AGE = 600 ; preload", int64(600)},
		{"should accept a quoted max-age value", `max-age="1200"`, int64(1200)},
		{"should use the first max-age directive", "max-age=60; max-age=31536000", int64(60)},
	}
	for _, c := range weakCases {
		t.Run(c.name, func(t *testing.T) {
			fs := scanHSTS(t, c.header)
			f := weak(fs)
			if f == nil {
				t.Fatalf("expected headers.weak_hsts, got %v", ids(fs))
			}
			if f.Severity != SeverityLow {
				t.Fatalf("severity: got %q want low", f.Severity)
			}
			got, ok := f.Detail["maxAge"]
			if !ok {
				t.Fatalf("detail has no maxAge: %v", f.Detail)
			}
			if got != c.maxAge {
				t.Fatalf("maxAge: got %#v want %#v", got, c.maxAge)
			}
			if _, missing := ids(fs)["headers.missing_hsts"]; missing {
				t.Fatalf("a present header must not also be reported missing: %v", ids(fs))
			}
		})
	}

	strongCases := []struct {
		name   string
		header string
	}{
		{"should not report exactly 180 days", "max-age=15552000"},
		{"should not report goodsite's one-year header with includeSubDomains", "max-age=31536000; includeSubDomains"},
	}
	for _, c := range strongCases {
		t.Run(c.name, func(t *testing.T) {
			fs := scanHSTS(t, c.header)
			if len(fs) != 0 {
				t.Fatalf("expected no findings, got %v", ids(fs))
			}
		})
	}

	t.Run("should report only missing_hsts when the header is absent", func(t *testing.T) {
		got := ids(scanHSTS(t, ""))
		if len(got) != 1 || got["headers.missing_hsts"] != SeverityMedium {
			t.Fatalf("got %v", got)
		}
	})
}
