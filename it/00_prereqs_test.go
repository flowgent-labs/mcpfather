package tests

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestMain is the integration-test entry point. It probes the local environment
// before any tests run so developers see immediately what is available, what
// will be skipped, and what MUST be fixed before IT tests can pass.
func TestMain(m *testing.M) {
	probeEnvironment()
	os.Exit(m.Run())
}

// ttyPrintf writes directly to /dev/tty to bypass go test's stderr capture pipe.
// Falls back to os.Stderr when /dev/tty is unavailable (CI environments).
func ttyPrintf(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	if tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0); err == nil {
		tty.WriteString(line)
		tty.Close()
	} else {
		os.Stderr.WriteString(line)
	}
}

// gfwMode returns the effective IN_CN_GFW value and detail string.
//
// Priority:
//  1. Explicit IN_CN_GFW env → use as-is
//  2. ipinfo.io reachable → use country code (CN or not)
//  3. ipinfo.io unreachable but goproxy.cn reachable → assume CN
//     (ipinfo.io is blocked by GFW; goproxy.cn is inside GFW)
//  4. Neither reachable → assume non-CN (CI runner with no internet)
func gfwMode() (string, string) {
	if v, ok := os.LookupEnv("IN_CN_GFW"); ok && v != "" {
		return v, fmt.Sprintf("IN_CN_GFW=%s (explicit)", v)
	}

	country := detectCountry()
	if country == "CN" {
		setCNEnv()
		return "true", "auto-detected CN → GOPROXY=goproxy.cn"
	}
	if country != "??" {
		os.Setenv("IN_CN_GFW", "false")
		return "false", "auto-detected non-CN (" + country + ")"
	}

	// ipinfo.io failed — probably behind GFW without proxy.
	// Try goproxy.cn directly to confirm.
	conn, err := net.DialTimeout("tcp", "goproxy.cn:443", 3*time.Second)
	if err == nil {
		conn.Close()
		setCNEnv()
		return "true", "auto-detected CN (goproxy.cn reachable, ipinfo.io blocked)"
	}

	os.Setenv("IN_CN_GFW", "false")
	return "false", "auto-detected non-CN (neither ipinfo.io nor goproxy.cn reachable)"
}

// setCNEnv sets GOPROXY + sumdb bypass when in GFW mode.
// Does NOT override env vars that are already explicitly set.
func setCNEnv() {
	os.Setenv("IN_CN_GFW", "true")
	if v, ok := os.LookupEnv("GOPROXY"); !ok || v == "" {
		os.Setenv("GOPROXY", "https://goproxy.cn,direct")
	}
	if v, ok := os.LookupEnv("GONOSUMDB"); !ok || v == "" {
		os.Setenv("GONOSUMDB", "*")
	}
	if v, ok := os.LookupEnv("GONOSUMCHECK"); !ok || v == "" {
		os.Setenv("GONOSUMCHECK", "*")
	}
}

func detectCountry() string {
	// Use plain HTTP (not HTTPS) — most GFW proxies (squid, tinyproxy, etc.)
	// do forward-proxy GET but can't CONNECT-tunnel TLS through the firewall.
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://ipinfo.io/country")
	if err != nil {
		return "??"
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 128))
	// /country returns plain text (e.g. "CN\n"), not JSON
	return strings.TrimSpace(string(body))
}

// probeEnvironment checks every external dependency the IT suite may need and
// prints a summary table directly to the terminal via /dev/tty.
func probeEnvironment() {
	ts := time.Now().Format("15:04:05")

	// Signal immediately so the developer knows compilation is done and
	// probing has started (gfwMode may take 0–3s querying ipinfo.io).
	ttyPrintf("\n  … IT Probe  %s  detecting network …\n", ts)

	// Resolve IN_CN_GFW (may take 0–3s querying ipinfo.io).
	gfwVal, gfwDetail := gfwMode()

	// ── Print results table ─────────────────────────────────────────────────
	ttyPrintf("  ╔══════════════════════════════════════════════════════════════╗\n")
	ttyPrintf("  ║  IT Environment Probe  %s                            ║\n", ts)
	ttyPrintf("  ╚══════════════════════════════════════════════════════════════╝\n")
	ttyPrintf("  ┌─ Network ────────────────────────────────────────────────────┐\n")
	ttyPrintf("  │ %s %-12s %s\n", mark(gfwVal == "true"), "IN_CN_GFW", gfwDetail)

	// Quick connectivity check: can we reach the internet at all?
	// goproxy.cn for CN, github.com for everywhere else.
	internetTarget := "github.com:443"
	if gfwVal == "true" {
		internetTarget = "goproxy.cn:443"
	}
	netOK, _ := probeTCP(internetTarget)
	ttyPrintf("  │ %s %-12s %s reachable\n", mark(netOK), "internet", internetTarget)
	ttyPrintf("  ├── Tools ─────────────────────────────────────────────────────┤\n")

	// docker ps — can we talk to the daemon?
	dockerOK, dockerDetail := probeLive("docker ps", []string{"/bin/docker", "ps"})
	ttyPrintf("  │ %s %-12s %s\n", mark(dockerOK), "docker", dockerDetail)
	if !dockerOK {
		ttyPrintf("  │   ⚠  docker ps failed — deploy tests will skip.\n")
	}

	// kubectl get ns — can we reach a k8s cluster via ~/.kube/config?
	kubectlOK, kubectlDetail := probeLive("kubectl get ns", []string{"kubectl", "get", "ns"})
	ttyPrintf("  │ %s %-12s %s\n", mark(kubectlOK), "kubectl", kubectlDetail)
	if !kubectlOK {
		if _, err := os.Stat("/etc/rancher/k3s/k3s.yaml"); err == nil {
			home, _ := os.UserHomeDir()
			target := home + "/.kube/config"
			ttyPrintf("  │   ⚠  /etc/rancher/k3s/k3s.yaml found but not accessible.\n")
			ttyPrintf("  │      Fix:  mkdir -p %s/.kube && sudo cp /etc/rancher/k3s/k3s.yaml %s && chmod 600 %s\n", home, target, target)
		}
	}

	// helm version — is helm present?
	helmOK, helmDetail := probeLive("helm version", []string{"helm", "version"})
	ttyPrintf("  │ %s %-12s %s\n", mark(helmOK), "helm", helmDetail)

	ttyPrintf("  └──────────────────────────────────────────────────────────────┘\n\n")
}

// probeLive runs a command and returns whether it succeeded plus a one-line
// detail string (first line of stdout, or first line of stderr on failure).
func probeLive(label string, cmdArgs []string) (bool, string) {
	_, err := exec.LookPath(cmdArgs[0])
	if err != nil {
		return false, "not found in PATH"
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	out, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(out))
	if err != nil {
		if outStr == "" {
			return false, "failed"
		}
		line := strings.SplitN(outStr, "\n", 2)[0]
		if len(line) > 55 {
			line = line[:52] + "..."
		}
		return false, line
	}
	line := strings.SplitN(outStr, "\n", 2)[0]
	if len(line) > 55 {
		line = line[:52] + "..."
	}
	return true, line
}

func probeTCP(addr string) (bool, string) {
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return false, "not reachable"
	}
	conn.Close()
	return true, "ok"
}

func mark(ok bool) string {
	if ok { return "✓" }
	return "✗"
}
