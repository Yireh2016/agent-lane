package checks

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type headerRule struct {
	header    string
	checkID   string
	severity  Severity
	title     string
	satisfied func(h http.Header) bool
}

var headerRules = []headerRule{
	{
		header:   "Strict-Transport-Security",
		checkID:  "headers.missing_hsts",
		severity: SeverityMedium,
		title:    "Strict-Transport-Security header is missing",
	},
	{
		header:   "Content-Security-Policy",
		checkID:  "headers.missing_csp",
		severity: SeverityMedium,
		title:    "Content-Security-Policy header is missing",
	},
	{
		header:   "X-Content-Type-Options",
		checkID:  "headers.missing_x_content_type_options",
		severity: SeverityLow,
		title:    "X-Content-Type-Options header is missing",
	},
	{
		header:   "X-Frame-Options",
		checkID:  "headers.missing_x_frame_options",
		severity: SeverityLow,
		title:    "X-Frame-Options header is missing",
		satisfied: func(h http.Header) bool {
			return strings.Contains(strings.ToLower(h.Get("Content-Security-Policy")), "frame-ancestors")
		},
	},
}

type HeaderOptions struct {
	Roots   *x509.CertPool
	Timeout time.Duration
}

func CheckHeaders(ctx context.Context, host string, port int, opts HeaderOptions) ([]Finding, error) {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			// Reason: header checks must still run on sites whose certificate fails;
			// trust problems are reported by CheckTLS.
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true, RootCAs: opts.Roots},
		},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	url := "https://" + net.JoinHostPort(host, strconv.Itoa(port)) + "/"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("headers %s: %w", url, err)
	}
	defer resp.Body.Close()

	var findings []Finding
	for _, rule := range headerRules {
		if resp.Header.Get(rule.header) != "" {
			continue
		}
		if rule.satisfied != nil && rule.satisfied(resp.Header) {
			continue
		}
		findings = append(findings, Finding{
			CheckID:  rule.checkID,
			Severity: rule.severity,
			Title:    rule.title,
			Detail:   map[string]any{"header": rule.header, "url": url},
		})
	}
	return findings, nil
}
