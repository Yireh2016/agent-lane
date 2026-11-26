package checks

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func startVersionedSite(t *testing.T, minVersion, maxVersion uint16) (string, int) {
	t.Helper()
	now := time.Now()
	_, _, issue := newCA(t)
	cert, err := issue(now.Add(-time.Hour), now.AddDate(0, 6, 0)).TLSCertificate()
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	srv.TLS = &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: minVersion, MaxVersion: maxVersion}
	srv.StartTLS()
	t.Cleanup(srv.Close)
	_, portStr, _ := net.SplitHostPort(srv.Listener.Addr().String())
	port, _ := strconv.Atoi(portStr)
	return "localhost", port
}

func TestCheckLegacyProtocols(t *testing.T) {
	t.Run("should report TLS 1.0 and 1.1 as high when the server accepts both", func(t *testing.T) {
		host, port := startVersionedSite(t, tls.VersionTLS10, tls.VersionTLS13)
		fs := CheckLegacyProtocols(context.Background(), host, port, LegacyOptions{})
		if len(fs) != 1 || fs[0].CheckID != "tls.legacy_protocol" || fs[0].Severity != SeverityHigh {
			t.Fatalf("got %v", fs)
		}
		got := fs[0].Detail["acceptedVersions"]
		if want := []string{"TLS 1.0", "TLS 1.1"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("acceptedVersions: got %v want %v", got, want)
		}
	})

	t.Run("should list only TLS 1.1 when the server refuses TLS 1.0", func(t *testing.T) {
		host, port := startVersionedSite(t, tls.VersionTLS11, tls.VersionTLS12)
		fs := CheckLegacyProtocols(context.Background(), host, port, LegacyOptions{})
		if len(fs) != 1 || fs[0].Severity != SeverityHigh {
			t.Fatalf("got %v", fs)
		}
		if got, want := fs[0].Detail["acceptedVersions"], []string{"TLS 1.1"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("acceptedVersions: got %v want %v", got, want)
		}
	})

	t.Run("should list only TLS 1.0 when the server accepts nothing newer", func(t *testing.T) {
		host, port := startVersionedSite(t, tls.VersionTLS10, tls.VersionTLS10)
		fs := CheckLegacyProtocols(context.Background(), host, port, LegacyOptions{})
		if len(fs) != 1 {
			t.Fatalf("got %v", fs)
		}
		if got, want := fs[0].Detail["acceptedVersions"], []string{"TLS 1.0"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("acceptedVersions: got %v want %v", got, want)
		}
	})

	t.Run("should report nothing for a server that accepts only TLS 1.2 and 1.3", func(t *testing.T) {
		host, port := startVersionedSite(t, tls.VersionTLS12, tls.VersionTLS13)
		if fs := CheckLegacyProtocols(context.Background(), host, port, LegacyOptions{}); len(fs) != 0 {
			t.Fatalf("expected no findings, got %v", fs)
		}
	})

	t.Run("should report nothing when the probe cannot connect", func(t *testing.T) {
		fs := CheckLegacyProtocols(context.Background(), "127.0.0.1", 1, LegacyOptions{Timeout: time.Second})
		if len(fs) != 0 {
			t.Fatalf("expected no findings, got %v", fs)
		}
	})
}
