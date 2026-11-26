package checks

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strconv"
	"time"
)

const expiringSoonWindow = 30 * 24 * time.Hour

type TLSOptions struct {
	Roots   *x509.CertPool
	Now     func() time.Time
	Timeout time.Duration
}

func CheckTLS(ctx context.Context, host string, port int, opts TLSOptions) ([]Finding, error) {
	now := time.Now
	if opts.Now != nil {
		now = opts.Now
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: timeout},
		// Reason: we must inspect certificates that fail verification, so verification
		// is done by hand below against opts.Roots.
		Config: &tls.Config{ServerName: host, InsecureSkipVerify: true},
	}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return nil, fmt.Errorf("tls dial %s:%d: %w", host, port, err)
	}
	defer conn.Close()

	state := conn.(*tls.Conn).ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("tls %s:%d: no peer certificates", host, port)
	}
	leaf := state.PeerCertificates[0]
	intermediates := x509.NewCertPool()
	for _, c := range state.PeerCertificates[1:] {
		intermediates.AddCert(c)
	}

	var findings []Finding
	at := now()

	if at.After(leaf.NotAfter) {
		findings = append(findings, Finding{
			CheckID:  "tls.certificate_expired",
			Severity: SeverityHigh,
			Title:    "TLS certificate has expired",
			Detail:   map[string]any{"notAfter": leaf.NotAfter.UTC().Format(time.RFC3339)},
		})
	} else if leaf.NotAfter.Sub(at) < expiringSoonWindow {
		findings = append(findings, Finding{
			CheckID:  "tls.certificate_expiring_soon",
			Severity: SeverityMedium,
			Title:    "TLS certificate expires within 30 days",
			Detail:   map[string]any{"notAfter": leaf.NotAfter.UTC().Format(time.RFC3339)},
		})
	}

	_, verifyErr := leaf.Verify(x509.VerifyOptions{
		DNSName:       host,
		Roots:         opts.Roots,
		Intermediates: intermediates,
		// Reason: expiry is reported by its own finding; verify trust at a time the cert is valid.
		CurrentTime: clampToValidity(at, leaf),
	})
	if verifyErr != nil {
		findings = append(findings, Finding{
			CheckID:  "tls.certificate_untrusted",
			Severity: SeverityHigh,
			Title:    "TLS certificate is not trusted for this hostname",
			Detail:   map[string]any{"reason": verifyErr.Error()},
		})
	}

	return findings, nil
}

func clampToValidity(t time.Time, cert *x509.Certificate) time.Time {
	if t.After(cert.NotAfter) {
		return cert.NotAfter.Add(-time.Second)
	}
	if t.Before(cert.NotBefore) {
		return cert.NotBefore.Add(time.Second)
	}
	return t
}
