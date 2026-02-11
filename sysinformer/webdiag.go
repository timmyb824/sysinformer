package sysinformer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"
)

type WebDiagOptions struct {
	Target     string
	Ping       bool
	Latency    bool
	DNS        bool
	HTTP       bool
	SSL        bool
	Whois      bool
	Trace      bool
	Full       bool
	TimeoutSec int
	Count      int
}

func ValidateTarget(raw string) (normalizedURL string, domain string, err error) {
	if raw == "" {
		return "", "", errors.New("empty target")
	}
	candidate := raw
	if !strings.HasPrefix(candidate, "http://") && !strings.HasPrefix(candidate, "https://") {
		candidate = "http://" + candidate
	}

	u, err := url.Parse(candidate)
	if err != nil {
		return "", "", err
	}
	if u.Host == "" {
		return "", "", fmt.Errorf("invalid url: %q", raw)
	}
	// Extract hostname without port (handles IPv4, IPv6, and hostnames)
	host := u.Hostname()
	if host == "" {
		return "", "", fmt.Errorf("invalid host: %q", raw)
	}

	// Resolve to ensure it exists
	if _, err := net.DefaultResolver.LookupHost(context.Background(), host); err != nil {
		return "", "", err
	}
	return u.String(), host, nil
}

func RunWebDiagnostics(opts WebDiagOptions) error {
	if opts.TimeoutSec <= 0 {
		opts.TimeoutSec = 10
	}
	if opts.Count <= 0 {
		opts.Count = 4
	}

	nURL, domain, err := ValidateTarget(opts.Target)
	if err != nil {
		return err
	}

	subtitle := nURL
	PrintPanel("Website Diagnostic", subtitle)

	runAll := opts.Full || !(opts.Ping || opts.Latency || opts.DNS || opts.HTTP || opts.SSL || opts.Whois || opts.Trace)

	if opts.Ping || runAll {
		PingWebsite(domain, opts.Count, time.Duration(opts.TimeoutSec)*time.Second)
	}
	if opts.Latency || runAll {
		CheckLatency(nURL, 3, time.Duration(opts.TimeoutSec)*time.Second)
	}
	if opts.DNS || runAll {
		CheckDNS(domain, time.Duration(opts.TimeoutSec)*time.Second)
	}
	if opts.HTTP || runAll {
		CheckHTTPStatusAndHeaders(nURL, time.Duration(opts.TimeoutSec)*time.Second)
	}
	if opts.SSL || (runAll && strings.HasPrefix(nURL, "https://")) {
		CheckSSL(domain, time.Duration(opts.TimeoutSec)*time.Second)
	}
	if opts.Whois || runAll {
		CheckWhois(domain, time.Duration(opts.TimeoutSec)*time.Second)
	}
	if opts.Trace || runAll {
		TraceRoute(domain, time.Duration(opts.TimeoutSec)*time.Second)
	}

	PrintPanel("Diagnostic complete", "")
	return nil
}

func PingWebsite(domain string, count int, timeout time.Duration) {
	fmt.Println("")
	PrintSectionHeader("PING")

	pingParam := "-c"
	if runtime.GOOS == "windows" {
		pingParam = "-n"
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ping", pingParam, fmt.Sprintf("%d", count), domain)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Ping failed: %v\n", err)
		return
	}
	text := string(out)

	// Best-effort summaries across OSes
	summary := []string{}
	for _, line := range strings.Split(text, "\n") {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		if strings.Contains(l, "packets transmitted") || strings.Contains(l, "Packets:") || strings.Contains(l, "packet loss") || strings.Contains(l, "loss") {
			summary = append(summary, l)
		}
	}
	for _, l := range summary {
		fmt.Println(l)
	}
}

func CheckLatency(targetURL string, count int, timeout time.Duration) {
	fmt.Println("")
	PrintSectionHeader("LATENCY")

	client := &http.Client{Timeout: timeout}

	headers := []string{"Request #", "Response Time (ms)", "Status"}
	rows := make([][]string, 0, count)

	var total float64
	var ok int
	for i := 1; i <= count; i++ {
		start := time.Now()
		resp, err := client.Get(targetURL)
		if err != nil {
			rows = append(rows, []string{fmt.Sprintf("%d", i), "N/A", "Failed"})
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		ms := float64(time.Since(start).Milliseconds())
		status := fmt.Sprintf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		rows = append(rows, []string{fmt.Sprintf("%d", i), fmt.Sprintf("%.0f", ms), status})
		total += ms
		ok++
	}

	RenderTable(headers, rows)
	if ok > 0 {
		fmt.Printf("Average response time: %.2f ms\n", total/float64(ok))
	} else {
		fmt.Println("Could not measure latency - all requests failed")
	}
}

func CheckDNS(domain string, timeout time.Duration) {
	fmt.Println("")
	PrintSectionHeader("DNS")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// A/AAAA via net
	if ips, err := net.DefaultResolver.LookupIPAddr(ctx, domain); err == nil {
		sort.Slice(ips, func(i, j int) bool { return ips[i].IP.String() < ips[j].IP.String() })
		rows := [][]string{}
		for _, ip := range ips {
			kind := "A"
			if ip.IP.To4() == nil {
				kind = "AAAA"
			}
			rows = append(rows, []string{kind, ip.IP.String()})
		}
		RenderTable([]string{"Record", "Value"}, rows)
	} else {
		fmt.Printf("DNS lookup failed: %v\n", err)
	}

	// MX / NS via net
	if mx, err := net.DefaultResolver.LookupMX(ctx, domain); err == nil {
		rows := [][]string{}
		sort.Slice(mx, func(i, j int) bool { return mx[i].Pref < mx[j].Pref })
		for _, r := range mx {
			rows = append(rows, []string{"MX", fmt.Sprintf("%d %s", r.Pref, strings.TrimSuffix(r.Host, "."))})
		}
		if len(rows) > 0 {
			RenderTable([]string{"Record", "Value"}, rows)
		}
	}
	if ns, err := net.DefaultResolver.LookupNS(ctx, domain); err == nil {
		rows := [][]string{}
		for _, r := range ns {
			rows = append(rows, []string{"NS", strings.TrimSuffix(r.Host, ".")})
		}
		if len(rows) > 0 {
			RenderTable([]string{"Record", "Value"}, rows)
		}
	}
	if txt, err := net.DefaultResolver.LookupTXT(ctx, domain); err == nil {
		rows := [][]string{}
		for _, r := range txt {
			rows = append(rows, []string{"TXT", r})
		}
		if len(rows) > 0 {
			RenderTable([]string{"Record", "Value"}, rows)
		}
	}
}

func CheckHTTPStatusAndHeaders(targetURL string, timeout time.Duration) {
	fmt.Println("")
	PrintSectionHeader("HTTP STATUS & HEADERS")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(targetURL)
	if err != nil {
		fmt.Printf("HTTP request failed: %v\n", err)
		return
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	fmt.Printf("Status: %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))

	rows := [][]string{}
	keys := make([]string, 0, len(resp.Header))
	for k := range resp.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		val := strings.Join(resp.Header.Values(k), ", ")
		val = normalizeWhitespace(val)
		// Keep this small because some headers (e.g. Report-To) can be enormous and will otherwise
		// force awkward terminal soft-wrapping.
		val = truncateForDisplay(val, 120)
		rows = append(rows, []string{k, val})
	}
	RenderKeyValueTable("Header", "Value", rows)
}

func CheckSSL(domain string, timeout time.Duration) {
	fmt.Println("")
	PrintSectionHeader("SSL/TLS CERTIFICATE")

	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(domain, "443"), &tls.Config{ServerName: domain})
	if err != nil {
		fmt.Printf("SSL check failed: %v\n", err)
		return
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		fmt.Println("No peer certificate presented")
		return
	}
	leaf := state.PeerCertificates[0]

	isValid := time.Now().After(leaf.NotBefore) && time.Now().Before(leaf.NotAfter)
	status := "Valid"
	if !isValid {
		status = "Invalid"
	}

	issuer := leaf.Issuer.Organization
	issuerStr := ""
	if len(issuer) > 0 {
		issuerStr = issuer[0]
	} else {
		issuerStr = leaf.Issuer.CommonName
	}

	sans := make([]string, 0, len(leaf.DNSNames))
	for _, n := range leaf.DNSNames {
		sans = append(sans, n)
	}
	sort.Strings(sans)
	sanStr := strings.Join(limitStrings(sans, 5), "\n")
	if len(sans) > 5 {
		sanStr = sanStr + fmt.Sprintf("\n... and %d more", len(sans)-5)
	}

	rows := [][]string{
		{"Common Name", leaf.Subject.CommonName},
		{"Issuer", issuerStr},
		{"Valid From", leaf.NotBefore.Format("2006-01-02 15:04:05")},
		{"Valid Until", leaf.NotAfter.Format("2006-01-02 15:04:05")},
		{"Status", status},
	}
	if sanStr != "" {
		rows = append(rows, []string{"Subject Alternative Names", sanStr})
	}
	RenderTable([]string{"Field", "Value"}, rows)
}

func limitStrings(in []string, n int) []string {
	if n <= 0 {
		return nil
	}
	if len(in) <= n {
		return in
	}
	return in[:n]
}

func TraceRoute(domain string, timeout time.Duration) {
	fmt.Println("")
	PrintSectionHeader("TRACEROUTE")

	cmdName := "traceroute"
	args := []string{domain}
	if runtime.GOOS == "windows" {
		cmdName = "tracert"
		args = []string{domain}
	}

	if _, err := exec.LookPath(cmdName); err != nil {
		fmt.Printf("%s not found on PATH. Install it (e.g. 'brew install traceroute' on macOS) or omit --trace.\n", cmdName)
		return
	}

	// Traceroute can take a while; use a more forgiving timeout.
	effectiveTimeout := timeout
	if effectiveTimeout < 30*time.Second {
		effectiveTimeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), effectiveTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdName, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			fmt.Printf("Traceroute timed out after %s\n", effectiveTimeout)
			return
		}
		fmt.Printf("Traceroute failed: %v\n", err)
		return
	}
	text := strings.TrimSpace(string(out))
	if text == "" {
		fmt.Println("No traceroute output")
		return
	}
	fmt.Println(text)
}

func truncateForDisplay(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 3 {
		return "..."
	}
	return string(runes[:max-3]) + "..."
}

func normalizeWhitespace(s string) string {
	// Some headers (e.g. Report-To / NEL) can contain huge JSON blobs.
	// Normalize whitespace so wrapping is predictable.
	return strings.Join(strings.Fields(s), " ")
}

func CheckWhois(domain string, timeout time.Duration) {
	fmt.Println("")
	PrintSectionHeader("WHOIS")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Use system whois to avoid pulling in a heavy dependency and to keep portability.
	cmd := exec.CommandContext(ctx, "whois", domain)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("WHOIS failed: %v\n", err)
		return
	}
	text := string(out)
	// Print a curated subset of lines (best-effort across registries)
	wantPrefixes := []string{
		"Domain Name:",
		"Registrar:",
		"Registry Expiry Date:",
		"Expiration Date:",
		"Creation Date:",
		"Updated Date:",
		"Name Server:",
		"DNSSEC:",
		"Registrar WHOIS Server:",
	}

	rows := [][]string{}
	for _, line := range strings.Split(text, "\n") {
		l := strings.TrimSpace(line)
		for _, p := range wantPrefixes {
			if strings.HasPrefix(l, p) {
				val := strings.TrimSpace(strings.TrimPrefix(l, p))
				rows = append(rows, []string{strings.TrimSuffix(p, ":"), val})
				break
			}
		}
		if len(rows) >= 25 {
			break
		}
	}

	if len(rows) == 0 {
		fmt.Println("WHOIS output (raw):")
		fmt.Println(text)
		return
	}
	RenderTable([]string{"Field", "Value"}, rows)
}
