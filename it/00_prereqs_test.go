package tests

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestMain is the integration-test entry point. It auto-installs missing
// binaries (kubectl, helm, k3s), starts a k3s cluster if none is reachable,
// then probes the environment.
func TestMain(m *testing.M) {
	autoSetup()
	probeEnvironment()
	os.Exit(m.Run())
}

// ─── multi-path binary lookup + auto-install ──────────────────────────────

// toolCacheDir is where downloaded binaries live (~/.local/mcpfather-it).
// Matches XDG-style conventions; CI caches it via actions/cache@v4.
func toolCacheDir() string {
	d, _ := os.UserHomeDir()
	if d == "" {
		d = "/tmp"
	}
	return filepath.Join(d, ".local", "mcpfather-it")
}

// lookPathMulti searches for a binary in PATH AND a list of well-known
// absolute locations. Returns the full path or "".
func lookPathMulti(name string, extra ...string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	for _, loc := range extra {
		if fi, err := os.Stat(loc); err == nil && !fi.IsDir() {
			return loc
		}
	}
	return ""
}

// k3sBinary returns the best available k3s binary path. It probes:
//
//	PATH → $CACHE/k3s → /usr/local/bin/k3s → /opt/homebrew/bin/k3s → ~/bin/k3s
func k3sBinary() string {
	home, _ := os.UserHomeDir()
	return lookPathMulti("k3s",
		filepath.Join(toolCacheDir(), "k3s"),
		"/usr/local/bin/k3s",
		"/opt/homebrew/bin/k3s",
		filepath.Join(home, "bin", "k3s"),
	)
}

// helmBinary returns the best available helm binary path.
//
//	PATH → $CACHE/helm → /usr/local/bin/helm → /opt/homebrew/bin/helm
func helmBinary() string {
	return lookPathMulti("helm",
		filepath.Join(toolCacheDir(), "helm"),
		"/usr/local/bin/helm",
		"/opt/homebrew/bin/helm",
	)
}

// ensureBinary downloads a tool to $CACHE/name if nothing was found.
func ensureBinary(name, linuxURL, darwinURL string, extraDirs ...string) {
	if p := lookPathMulti(name, extraDirs...); p != "" {
		return // already available
	}
	cd := toolCacheDir()
	os.MkdirAll(cd, 0755)
	os.Setenv("PATH", cd+":"+os.Getenv("PATH"))

	url := linuxURL
	if runtime.GOOS == "darwin" {
		url = darwinURL
	}
	dst := filepath.Join(cd, name)
	ttyPrintf("\n  … auto-setup: installing %s (to %s) …\n", name, cd)
	if err := installFromURL(dst, url); err != nil {
		ttyPrintf("\n  … auto-setup: download %s failed: %v — will skip k8s tests\n", name, err)
		return
	}
	ttyPrintf("\n  … auto-setup: %s installed at %s\n", name, dst)
}

// autoSetup ensures kubectl, helm, and k3s are present.
func autoSetup() {
	cd := toolCacheDir()
	os.MkdirAll(cd, 0755)
	os.Setenv("PATH", cd+":"+os.Getenv("PATH"))

	ensureBinary("kubectl",
		"https://dl.k8s.io/release/v1.32.0/bin/linux/amd64/kubectl",
		"https://dl.k8s.io/release/v1.32.0/bin/darwin/arm64/kubectl",
		"/usr/local/bin/kubectl",
		"/opt/homebrew/bin/kubectl",
	)

	// helm — tar.gz, need to extract
	if helmBinary() == "" {
		ttyPrintf("\n  … auto-setup: installing helm (to %s) …\n", cd)
		helmTar := filepath.Join(cd, "helm.tar.gz")
		helmDir := filepath.Join(cd, "helm-tmp")
		os.MkdirAll(helmDir, 0755)
		helmURL := "https://get.helm.sh/helm-v3.17.0-linux-amd64.tar.gz"
		if runtime.GOOS == "darwin" {
			helmURL = "https://get.helm.sh/helm-v3.17.0-darwin-arm64.tar.gz"
		}
		if err := installFromURL(helmTar, helmURL); err == nil {
			exec.Command("tar", "-xzf", helmTar, "-C", helmDir).Run()
			filepath.Walk(helmDir, func(p string, _ os.FileInfo, _ error) error {
				if filepath.Base(p) == "helm" {
					os.Rename(p, filepath.Join(cd, "helm"))
					os.Chmod(filepath.Join(cd, "helm"), 0755)
				}
				return nil
			})
			os.RemoveAll(helmDir)
			os.Remove(helmTar)
			if helmBinary() != "" {
				ttyPrintf("\n  … auto-setup: helm installed at %s\n", helmBinary())
			}
		}
	}

	// k3s
	ensureBinary("k3s",
		"https://github.com/k3s-io/k3s/releases/download/v1.32.3%2Bk3s1/k3s",
		"https://github.com/k3s-io/k3s/releases/download/v1.32.3%2Bk3s1/k3s", // k3s binary is multi-arch
		"/usr/local/bin/k3s",
		"/opt/homebrew/bin/k3s",
	)

	// ── kubeconfig ──
	if _, err := os.Stat("/etc/rancher/k3s/k3s.yaml"); err == nil {
		home, _ := os.UserHomeDir()
		cfgFile := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(cfgFile), 0700)
			data, _ := os.ReadFile("/etc/rancher/k3s/k3s.yaml")
			if len(data) > 0 {
				ttyPrintf("\n  … auto-setup: copying k3s kubeconfig → %s …\n", cfgFile)
				os.WriteFile(cfgFile, data, 0600)
			}
		}
	}

	// ── k3s server ──
	kubectlOK, _ := probeLive("kubectl get ns", []string{"kubectl", "get", "ns"})
	if !kubectlOK && k3sBinary() != "" {
		ttyPrintf("\n  … auto-setup: starting k3s server …\n")
		k3sPath := k3sBinary()
		cmd := exec.Command("sudo", k3sPath, "server", "--write-kubeconfig-mode=644")
		// Redirect stdout/stderr to /dev/null so go test does not
		// hang at exit waiting for the long-lived k3s process to
		// close inherited I/O pipes (WaitDelay expired).
		dn, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		if dn != nil {
			cmd.Stderr = dn
			cmd.Stdout = dn
		}
		cmd.Start()

		for i := 0; i < 30; i++ {
			if _, err := os.Stat("/etc/rancher/k3s/k3s.yaml"); err == nil {
				home, _ := os.UserHomeDir()
				cfgFile := filepath.Join(home, ".kube", "config")
				os.MkdirAll(filepath.Dir(cfgFile), 0700)
				data, _ := os.ReadFile("/etc/rancher/k3s/k3s.yaml")
				os.WriteFile(cfgFile, data, 0600)
				ttyPrintf("\n  … auto-setup: k3s kubeconfig ready\n")
				break
			}
			time.Sleep(2 * time.Second)
		}

		for i := 0; i < 60; i++ {
			out, _ := exec.Command("kubectl", "get", "nodes", "-o", "name").CombinedOutput()
			if strings.TrimSpace(string(out)) != "" {
				exec.Command("kubectl", "wait", "--for=condition=Ready", "node", "--all", "--timeout=60s").Run()
				// Print namespaces so CI logs confirm cluster readiness.
				if nsOut, err := exec.Command("kubectl", "get", "ns").CombinedOutput(); err == nil {
					ttyPrintf("\n  … auto-setup: k3s cluster ready, namespaces:\n%s", string(nsOut))
				} else {
					ttyPrintf("\n  … auto-setup: k3s cluster ready\n")
				}
				return
			}
			time.Sleep(2 * time.Second)
		}
		ttyPrintf("\n  … auto-setup: k3s may still be starting\n")
	}
}

// installFromURL downloads a file to dest.
func installFromURL(dst, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	f, err := os.Create(dst + ".tmp")
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		os.Remove(dst + ".tmp")
		return err
	}
	os.Chmod(dst+".tmp", 0755)
	return os.Rename(dst+".tmp", dst)
}

// ─── helpers ──────────────────────────────────────────────────────────────

func ttyPrintf(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	if tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0); err == nil {
		tty.WriteString(line)
		tty.Close()
	} else {
		os.Stderr.WriteString(line)
	}
}

func probeLive(label string, cmdArgs []string) (bool, string) {
	_, err := exec.LookPath(cmdArgs[0])
	if err != nil {
		return false, "not found in PATH"
	}
	out, err := exec.Command(cmdArgs[0], cmdArgs[1:]...).CombinedOutput()
	s := strings.TrimSpace(string(out))
	if err != nil {
		if s == "" {
			return false, "failed"
		}
		line := strings.SplitN(s, "\n", 2)[0]
		if len(line) > 55 {
			line = line[:52] + "..."
		}
		return false, line
	}
	line := strings.SplitN(s, "\n", 2)[0]
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
	if ok {
		return "✓"
	}
	return "✗"
}

// ─── GFW detection ─────────────────────────────────────────────────────────

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
		// Country identified and is NOT CN (US, SG, JP, etc.) — no GFW.
		os.Setenv("IN_CN_GFW", "false")
		return "false", "auto-detected non-CN (" + country + ")"
	}
	conn, err := net.DialTimeout("tcp", "goproxy.cn:443", 3*time.Second)
	if err == nil {
		conn.Close()
		setCNEnv()
		return "true", "auto-detected CN (goproxy.cn reachable, ipinfo.io blocked)"
	}
	os.Setenv("IN_CN_GFW", "false")
	return "false", "auto-detected non-CN (neither reachable)"
}

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
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://ipinfo.io/country")
	if err != nil {
		return "??"
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 128))
	return strings.TrimSpace(string(body))
}

// ─── environment probe table ──────────────────────────────────────────────

func probeEnvironment() {
	ts := time.Now().Format("15:04:05")
	gfwVal, gfwDetail := gfwMode()

	ttyPrintf("\n")
	ttyPrintf("  ╔══════════════════════════════════════════════════════════════╗\n")
	ttyPrintf("  ║  IT Environment Probe  %s                            ║\n", ts)
	ttyPrintf("  ╚══════════════════════════════════════════════════════════════╝\n")
	ttyPrintf("  ┌─ Network ────────────────────────────────────────────────────┐\n")
	ttyPrintf("  │ %s %-12s %s\n", mark(gfwVal == "true"), "IN_CN_GFW", gfwDetail)

	internetTarget := "github.com:443"
	if gfwVal == "true" {
		internetTarget = "goproxy.cn:443"
	}
	netOK, _ := probeTCP(internetTarget)
	ttyPrintf("  │ %s %-12s %s reachable\n", mark(netOK), "internet", internetTarget)
	ttyPrintf("  ├── Tools ─────────────────────────────────────────────────────┤\n")

	dockerOK, dockerDetail := probeLive("docker ps", []string{"/bin/docker", "ps"})
	ttyPrintf("  │ %s %-12s %s\n", mark(dockerOK), "docker", dockerDetail)

	kubectlOK, kubectlDetail := probeLive("kubectl get ns", []string{"kubectl", "get", "ns"})
	ttyPrintf("  │ %s %-12s %s\n", mark(kubectlOK), "kubectl", kubectlDetail)
	if !kubectlOK {
		if _, err := os.Stat("/etc/rancher/k3s/k3s.yaml"); err == nil {
			// k3s kubeconfig exists but is root-only (0600) — autoSetup
			// already tried to copy it to ~/.kube/config. If it still
			// fails, the copy didn't work (permission denied without sudo).
			home, _ := os.UserHomeDir()
			ttyPrintf("  │   ⚠  /etc/rancher/k3s/k3s.yaml found but not accessible.\n")
			ttyPrintf("  │      Fix:  sudo cp %s %s/.kube/config && chmod 600 %s/.kube/config\n",
				"/etc/rancher/k3s/k3s.yaml", home, home)
		}
	}

	helmOK, helmDetail := probeLive("helm version", []string{"helm", "version"})
	ttyPrintf("  │ %s %-12s %s\n", mark(helmOK), "helm", helmDetail)

	ttyPrintf("  └──────────────────────────────────────────────────────────────┘\n\n")
}
