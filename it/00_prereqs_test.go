package tests

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestMain is the integration-test entry point. It probes the local environment
// before any tests run so developers see immediately what is available and what
// will be skipped — no more waiting 7 minutes only to find k3s was unreachable.
func TestMain(m *testing.M) {
	probeEnvironment()
	os.Exit(m.Run())
}

// ttyPrintf is like ttyPrintf( ...) but writes directly to /dev/tty
// to bypass go test's stderr capture pipe. Same trick as logProgress.
func ttyPrintf(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	if tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0); err == nil {
		tty.WriteString(line)
		tty.Close()
	}
}

// probeEnvironment checks every external dependency the IT suite may need and
// prints a summary table directly to the terminal via /dev/tty.
func probeEnvironment() {
	ts := time.Now().Format("15:04:05")
	ttyPrintf("\n")
	ttyPrintf("  ╔══════════════════════════════════════════════════════════════╗\n")
	ttyPrintf("  ║  IT Environment Probe  %s                              ║\n", ts)
	ttyPrintf("  ╚══════════════════════════════════════════════════════════════╝\n")

	type check struct {
		label   string
		path    string // binary to LookPath
		detail  string // extra info if found
		hard    bool   // hard prerequisite (exit if missing)
		enabled bool
	}

	checks := []check{
		{label: "kubectl", path: "kubectl"},
		{label: "helm", path: "helm"},
		{label: "docker", path: "/bin/docker"},
		{label: "k3s", path: "k3s"},
		{label: "sudo", path: "sudo"},
	}

	// Probe binaries
	for i := range checks {
		c := &checks[i]
		p, err := exec.LookPath(c.path)
		if err == nil {
			c.enabled = true
			// Use --client for kubectl to avoid cluster-config warnings
			verFlag := "--version"
			if c.path == "kubectl" {
				verFlag = "version" // kubectl version doesn't need --client flag
			}
			out, verr := exec.Command(p, verFlag).CombinedOutput()
			if verr == nil && c.path == "kubectl" {
				// kubectl version prints to stderr on k3s; skip detail for brevity
				c.detail = "found"
			} else if verr == nil {
				line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
				if len(line) > 60 {
					line = line[:57] + "..."
				}
				c.detail = line
			}
		}
	}

	// Probe k8s cluster connectivity (with sudo fallback for root-owned k3s kubeconfig)
	k8sOK := false
	var k8sDetail string
	for _, cmd := range [][]string{{"kubectl", "cluster-info"}, {"sudo", "kubectl", "cluster-info"}} {
		if _, err := exec.LookPath(cmd[0]); err != nil {
			continue
		}
		c := exec.Command(cmd[0], cmd[1:]...)
		_, err := c.CombinedOutput()
		if err == nil {
			k8sOK = true
			if cmd[0] == "sudo" {
				k8sDetail = "reachable via sudo kubectl"
				deployKubectl = []string{"sudo", "kubectl"}
			} else {
				k8sDetail = "reachable"
				deployKubectl = []string{"kubectl"}
			}
			break
		}
	}

	// Probe Keycloak / Docker availability
	dockerOK := false
	for _, c := range checks {
		if c.label == "docker" && c.enabled {
			dockerOK = true
			break
		}
	}

	// ── Print results ──────────────────────────────────────────────────────────
	ttyPrintf( "  ┌─ Tools ──────────────────────────────────────────────────────┐\n")
	for _, c := range checks {
		mark := "✗"
		if c.enabled {
			mark = "✓"
		}
		ttyPrintf( "  │ %s %-8s", mark, c.label)
		if c.detail != "" {
			ttyPrintf( "  %s", c.detail)
		}
		ttyPrintf( "\n")
	}
	ttyPrintf( "  ├──────────────────────────────────────────────────────────────┤\n")

	// k8s cluster
	mark := "✗"
	if k8sOK {
		mark = "✓"
	}
	ttyPrintf( "  │ %s %-8s  %s\n", mark, "cluster", k8sDetail)
	if !k8sOK {
		k3sCfg := "/etc/rancher/k3s/k3s.yaml"
		if _, err := os.Stat(k3sCfg); err == nil {
			ttyPrintf( "  │   ⚠  k3s kubeconfig is root-only.\n")
			ttyPrintf( "  │      Fix: sudo chmod 644 %s\n", k3sCfg)
		}
		ttyPrintf( "  │      → Deploy tests will be skipped.\n")
	}

	// Docker / Keycloak
	mark = "✗"
	if dockerOK {
		mark = "✓"
	}
	ttyPrintf( "  │ %s %-8s", mark, "docker")
	if dockerOK {
		ttyPrintf( "  OIDC tests can use real Keycloak\n")
	} else {
		ttyPrintf( "  OIDC tests that need Keycloak will skip\n")
	}

	ttyPrintf( "  └──────────────────────────────────────────────────────────────┘\n\n")
}
