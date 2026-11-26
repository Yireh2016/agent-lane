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
	if hsts := resp.Header.Get("Strict-Transport-Security"); hsts != "" {
		if f, weak := weakHSTS(hsts, url); weak {
			findings = append(findings, f)
		}
	}
	return findings, nil
}

const minHSTSMaxAge = 15552000

func weakHSTS(value, url string) (Finding, bool) {
	maxAge, ok := hstsMaxAge(value)
	if ok && maxAge >= minHSTSMaxAge {
		return Finding{}, false
	}
	var detailAge any
	if ok {
		detailAge = maxAge
	}
	return Finding{
		CheckID:  "headers.weak_hsts",
		Severity: SeverityLow,
		Title:    "Strict-Transport-Security max-age is shorter than 180 days",
		Detail: map[string]any{
			"header":    "Strict-Transport-Security",
			"url":       url,
			"value":     value,
			"maxAge":    detailAge,
			"minMaxAge": int64(minHSTSMaxAge),
		},
	}, true
}

// Reason: only the first max-age directive counts; if it does not hold a
// non-negative integer the header has no usable max-age.
func hstsMaxAge(value string) (int64, bool) {
	for _, directive := range strings.Split(value, ";") {
		name, raw, _ := strings.Cut(directive, "=")
		if !strings.EqualFold(strings.TrimSpace(name), "max-age") {
			continue
		}
		raw = strings.TrimSpace(raw)
		if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
			raw = raw[1 : len(raw)-1]
		}
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n < 0 {
			return 0, false
		}
		return n, true
	}
	return 0, false
}
