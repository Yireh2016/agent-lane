package checks

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"time"
)

type LegacyOptions struct {
	Timeout time.Duration
}

var legacyVersions = []struct {
	version uint16
	name    string
}{
	{tls.VersionTLS10, "TLS 1.0"},
	{tls.VersionTLS11, "TLS 1.1"},
}

// CheckLegacyProtocols reports a server that completes a handshake at TLS 1.0 or 1.1.
// A probe that fails for any reason counts as "not accepted": it never fails the scan.
func CheckLegacyProtocols(ctx context.Context, host string, port int, opts LegacyOptions) []Finding {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	addr := net.JoinHostPort(host, strconv.Itoa(port))

	var accepted []string
	for _, v := range legacyVersions {
		if acceptsVersion(ctx, addr, host, v.version, timeout) {
			accepted = append(accepted, v.name)
		}
	}
	if len(accepted) == 0 {
		return nil
	}
	return []Finding{{
		CheckID:  "tls.legacy_protocol",
		Severity: SeverityHigh,
		Title:    "Server accepts legacy TLS versions (1.0 or 1.1)",
		Detail:   map[string]any{"acceptedVersions": accepted},
	}}
}

func acceptsVersion(ctx context.Context, addr, host string, version uint16, timeout time.Duration) bool {
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: timeout},
		Config: &tls.Config{
			ServerName: host,
			// Reason: only the protocol version matters here; certificate problems are reported by CheckTLS.
			InsecureSkipVerify: true,
			MinVersion:         version,
			MaxVersion:         version,
			// Reason: Go's default client suites leave out ones many legacy-only servers offer (RSA key exchange, 3DES).
			CipherSuites: allCipherSuites(),
		},
	}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func allCipherSuites() []uint16 {
	var ids []uint16
	for _, s := range append(tls.CipherSuites(), tls.InsecureCipherSuites()...) {
		ids = append(ids, s.ID)
	}
	return ids
}
