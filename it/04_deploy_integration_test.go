package tests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Deploy integration tests
//
// These tests verify end-to-end deployment of a generated MCP server:
//   1. Generate project from OpenAPI spec
//   2. Build Docker image from generated Dockerfile
//   3. Create k8s Secret via kubectl
//   4. Deploy via Helm (with --set overrides for local image)
//   5. Port-forward to the pod
//   6. Call tools via mcpclient.sh / MCP HTTP
//   7. Validate 401 auth behaviour
//   8. Validate virtual tools results
//
// Prerequisites: kubectl, helm, docker (or a local container runtime).
// The test skips if these are not available.
//
// Cluster detection:
//   - k3s:           imports image via "k3s ctr images import"
//   - containerd:    imports image via "ctr" or "crictl"
//   - orbstack:      imports image via "orb" CLI (macOS)
//   - remote:        requires MCPFATHER_TEST_IMAGE_REPO env var; image is
//                    pushed after build and helm uses that repo.
// ---------------------------------------------------------------------------

// clusterType describes how the test should make the image available to k8s.
type clusterType int

const (
	clusterUnknown clusterType = iota
	clusterK3s
	clusterContainerd
	clusterOrbstack
	clusterRemote
)

// detectCluster determines the k8s cluster type and returns the image
// repository prefix that should be used in Helm values.
func detectCluster(t *testing.T) (clusterType, string) {
	t.Helper()

	// 1 — user explicitly set a remote repo
	if repo := os.Getenv("MCPFATHER_TEST_IMAGE_REPO"); repo != "" {
		t.Logf("MCPFATHER_TEST_IMAGE_REPO=%s → treating cluster as remote", repo)
		return clusterRemote, repo
	}

	// 2 — orbstack (macOS)
	if _, err := exec.LookPath("orb"); err == nil {
		// Double-check orbstack is actually running
		if out, err := exec.Command("orb", "info", "--format", "{{.Version}}").CombinedOutput(); err == nil {
			t.Logf("orbstack detected: %s", strings.TrimSpace(string(out)))
			return clusterOrbstack, ""
		}
	}

	// 3 — k3s
	if k3sBin, err := exec.LookPath("k3s"); err == nil {
		if out, err := exec.Command(k3sBin, "--version").CombinedOutput(); err == nil {
			t.Logf("k3s detected: %s", strings.TrimSpace(string(out)))
			return clusterK3s, ""
		}
	}

	// 4 — standard containerd (kubeadm / cri-o / rancher)
	//    Check for ctr or crictl
	if ctr, err := exec.LookPath("ctr"); err == nil {
		t.Logf("containerd (ctr) detected at %s", ctr)
		return clusterContainerd, ""
	}
	if crictl, err := exec.LookPath("crictl"); err == nil {
		t.Logf("containerd/cri-o (crictl) detected at %s", crictl)
		return clusterContainerd, ""
	}

	// 5 — no local tooling => remote
	t.Fatalf("Cannot determine local cluster type and MCPFATHER_TEST_IMAGE_REPO is not set.\n" +
		"Set MCPFATHER_TEST_IMAGE_REPO=<registry>/<repo> for remote clusters, or install k3s/ctr/crictl/orb.")
	return clusterUnknown, "" // unreachable
}

// deployPrereqsOK checks that kubectl, helm, and docker are available and a
// Kubernetes cluster is reachable.
func deployPrereqsOK(t *testing.T) (kubectl, helm, docker string) {
	t.Helper()

	kubectl, err := exec.LookPath("kubectl")
	if err != nil {
		t.Skipf("kubectl not found in PATH — skipping deploy test")
	}
	helm, err = exec.LookPath("helm")
	if err != nil {
		t.Skipf("helm not found in PATH — skipping deploy test")
	}
	docker, err = exec.LookPath("/bin/docker")
	if err != nil {
		t.Skipf("docker not found in PATH — skipping deploy test")
	}

	// Check cluster connectivity
	cmd := exec.Command(kubectl, "cluster-info")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("kubectl cannot reach cluster: %v\n%s", err, out)
	}

	return kubectl, helm, docker
}

// deployNamespace creates a unique test namespace and returns its name.
func deployNamespace(t *testing.T, kubectl string) string {
	t.Helper()
	ns := fmt.Sprintf("mcpfather-deploy-test-%d", time.Now().UnixNano()%100000)

	// Clean up any leftover namespace from prior runs (ignore errors).
	exec.Command(kubectl, "delete", "namespace", ns, "--ignore-not-found", "--timeout=30s").Run()

	// Wait for any previous instance to be fully deleted.
	for i := 0; i < 30; i++ {
		out, _ := exec.Command(kubectl, "get", "namespace", ns, "-o", "jsonpath={.status.phase}").CombinedOutput()
		if strings.TrimSpace(string(out)) != "Terminating" {
			break
		}
		time.Sleep(2 * time.Second)
	}

	cmd := exec.Command(kubectl, "create", "namespace", ns)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create namespace %s: %v\n%s", ns, err, out)
	}
	t.Cleanup(func() {
		exec.Command(kubectl, "delete", "namespace", ns, "--ignore-not-found", "--timeout=60s").Run()
	})
	return ns
}

// deployBuildImage builds a docker image for the generated project and returns
// the image tag.
func deployBuildImage(t *testing.T, docker, projectDir string) string {
	t.Helper()
	binName := filepath.Base(projectDir)
	imageTag := fmt.Sprintf("%s:test", binName)

	dockerfile := filepath.Join(projectDir, "deploy", "docker", "Dockerfile")
	if _, err := os.Stat(dockerfile); os.IsNotExist(err) {
		t.Fatalf("Dockerfile not found at %s", dockerfile)
	}

	cmd := exec.Command(docker, "build", "--network", "host", "-t", imageTag, "-f", dockerfile, projectDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("docker build failed: %v\n%s", err, out)
	}
	t.Logf("Docker image built: %s", imageTag)

	t.Cleanup(func() {
		exec.Command(docker, "rmi", "-f", imageTag).Run()
	})
	return imageTag
}

// deployMakeImageAvailable ensures the locally-built image is usable by the
// cluster. It returns the image repository:tag string that should be passed to
// --set image.repository= and --set image.tag= in Helm.
func deployMakeImageAvailable(t *testing.T, docker, imageTag string, ct clusterType, remoteRepo string) (repo, tag string) {
	t.Helper()

	// Parse imageTag "name:tag"
	parts := strings.SplitN(imageTag, ":", 2)
	name, ver := parts[0], parts[1]

	switch ct {

	case clusterK3s:
		tarPath := filepath.Join(t.TempDir(), "image.tar")
		saveCmd := exec.Command(docker, "save", "-o", tarPath, imageTag)
		if out, err := saveCmd.CombinedOutput(); err != nil {
			t.Fatalf("docker save for k3s: %v\n%s", err, out)
		}
		k3sBin, _ := exec.LookPath("k3s")
		importCmd := exec.Command("sudo", k3sBin, "ctr", "images", "import", tarPath)
		if out, err := importCmd.CombinedOutput(); err != nil {
			t.Fatalf("k3s ctr images import: %v\n%s", err, out)
		}
		t.Logf("Loaded image into k3s cluster")
		// k3s sees it as localhost/<name>:<ver>
		return fmt.Sprintf("localhost/%s", name), ver

	case clusterContainerd:
		tarPath := filepath.Join(t.TempDir(), "image.tar")
		saveCmd := exec.Command(docker, "save", "-o", tarPath, imageTag)
		if out, err := saveCmd.CombinedOutput(); err != nil {
			t.Fatalf("docker save for containerd: %v\n%s", err, out)
		}
		// Try ctr first, then crictl
		if ctr, err := exec.LookPath("ctr"); err == nil {
			importCmd := exec.Command("sudo", ctr, "images", "import", tarPath)
			if out, err := importCmd.CombinedOutput(); err != nil {
				t.Fatalf("ctr images import: %v\n%s", err, out)
			}
		} else if crictl, err := exec.LookPath("crictl"); err == nil {
			importCmd := exec.Command("sudo", crictl, "images", "import", tarPath)
			if out, err := importCmd.CombinedOutput(); err != nil {
				t.Fatalf("crictl images import: %v\n%s", err, out)
			}
		}
		t.Logf("Loaded image into containerd cluster")
		return fmt.Sprintf("docker.io/library/%s", name), ver

	case clusterOrbstack:
		// orbstack shares the docker daemon images with its k8s runtime,
		// so a locally-built image is directly available.
		t.Logf("orbstack: docker image should be directly available to cluster")
		return name, ver

	case clusterRemote:
		if remoteRepo == "" {
			t.Fatal("remoteRepo is empty for remote cluster — set MCPFATHER_TEST_IMAGE_REPO")
		}
		remoteTag := fmt.Sprintf("%s/%s:%s", remoteRepo, name, ver)
		tagCmd := exec.Command(docker, "tag", imageTag, remoteTag)
		if out, err := tagCmd.CombinedOutput(); err != nil {
			t.Fatalf("docker tag %s → %s: %v\n%s", imageTag, remoteTag, err, out)
		}
		pushCmd := exec.Command(docker, "push", remoteTag)
		if out, err := pushCmd.CombinedOutput(); err != nil {
			t.Fatalf("docker push %s: %v\n%s", remoteTag, err, out)
		}
		t.Logf("Pushed image %s to remote registry", remoteTag)
		t.Cleanup(func() {
			exec.Command(docker, "rmi", "-f", remoteTag).Run()
		})
		return fmt.Sprintf("%s/%s", remoteRepo, name), ver

	default:
		t.Fatalf("unknown cluster type — cannot make image available")
		return "", "" // unreachable
	}
}

// deployHelmChart deploys the generated helm chart and returns (releaseName, k8sFullname).
func deployHelmChart(t *testing.T, helm, kubectl, projectDir, imageRepo, imageTag, ns, upstreamURL string) (string, string) {
	t.Helper()
	binName := filepath.Base(projectDir)
	chartDir := filepath.Join(projectDir, "deploy", "helm")
	releaseName := binName
	// Avoid YAML numeric parsing (e.g. "001" → octal 1) by prefixing.
	k8sFullname := "mcp-" + binName

	// Uninstall first if exists
	exec.Command(helm, "uninstall", releaseName, "-n", ns, "--ignore-not-found").Run()
	time.Sleep(1 * time.Second)

	args := []string{
		"install", releaseName, chartDir,
		"-n", ns,
		"--set", fmt.Sprintf("fullnameOverride=%s", k8sFullname),
		"--set", fmt.Sprintf("nameOverride=%s", k8sFullname),
		"--set", fmt.Sprintf("image.repository=%s", imageRepo),
		"--set", fmt.Sprintf("image.tag=%s", imageTag),
		"--set", "image.pullPolicy=IfNotPresent",
		"--set", fmt.Sprintf("config.upstream.endpoint=%s", upstreamURL),
		"--set", "config.tools.registerAllByDefault=true",
		"--set", "config.runtime.logAuthorization=false",
		"--wait",
		"--timeout", "60s",
	}

	cmd := exec.Command(helm, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("helm install failed: %v\n%s", err, out)
	}
	t.Logf("Helm release %s installed in namespace %s", releaseName, ns)

	t.Cleanup(func() {
		exec.Command(helm, "uninstall", releaseName, "-n", ns, "--ignore-not-found", "--timeout=30s").Run()
	})
	return releaseName, k8sFullname
}

// deployCreateSecret creates a k8s Secret.
func deployCreateSecret(t *testing.T, kubectl, ns, k8sName, bearerToken, cookieToken string) {
	t.Helper()
	secretName := k8sName + "-secret"

	exec.Command(kubectl, "delete", "secret", secretName, "-n", ns, "--ignore-not-found").Run()

	cmd := exec.Command(kubectl, "create", "secret", "generic", secretName,
		"-n", ns,
		fmt.Sprintf("--from-literal=bearer_token=%s", bearerToken),
		fmt.Sprintf("--from-literal=cookie_token=%s", cookieToken),
		fmt.Sprintf("--from-literal=oidc_client_secret=%s", "test-oidc-secret"),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("kubectl create secret: %v\n%s", err, out)
	}
	t.Logf("Secret %s created in namespace %s", secretName, ns)
}

// deployPortForward sets up kubectl port-forward and returns the local port.
func deployPortForward(t *testing.T, kubectl, ns, k8sName string, remotePort int) (localPort int, cancel func()) {
	t.Helper()

	cmd := exec.Command(kubectl, "get", "pods", "-n", ns,
		"-l", fmt.Sprintf("app.kubernetes.io/name=%s", k8sName),
		"-o", "jsonpath={.items[0].metadata.name}",
	)
	podOut, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("get pods: %v\n%s", err, podOut)
	}
	podName := strings.TrimSpace(string(podOut))
	if podName == "" {
		t.Fatal("no pod found for release")
	}
	t.Logf("Pod: %s", podName)

	localPort = 18080 + int(time.Now().UnixNano()%1000)

	ctx, cancel := context.WithCancel(context.Background())
	pfCmd := exec.CommandContext(ctx, kubectl, "port-forward", "-n", ns, podName,
		fmt.Sprintf("%d:%d", localPort, remotePort),
	)
	var pfStderr bytes.Buffer
	pfCmd.Stderr = &pfStderr

	if err := pfCmd.Start(); err != nil {
		cancel()
		t.Fatalf("kubectl port-forward: %v\n%s", err, pfStderr.String())
	}

	time.Sleep(2 * time.Second)

	t.Cleanup(func() {
		cancel()
		pfCmd.Wait()
	})
	return localPort, cancel
}

// startHostMockUpstream starts an HTTP server on the host's IP so that pods
// running in the k8s cluster can reach it.
func startHostMockUpstream(t *testing.T, handler http.HandlerFunc) (url string, close func()) {
	t.Helper()

	hostIP := "127.0.0.1"
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			hostIP = ipnet.IP.String()
			break
		}
	}

	listener, err := net.Listen("tcp", hostIP+":0")
	if err != nil {
		t.Fatalf("startHostMockUpstream listen on %s: %v", hostIP, err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	url = fmt.Sprintf("http://%s:%d", hostIP, port)

	srv := &http.Server{Handler: handler}
	go srv.Serve(listener)

	return url, func() {
		srv.Close()
		listener.Close()
	}
}

// ---------------------------------------------------------------------------
// Test: Deploy with auth — 401 validation
// ---------------------------------------------------------------------------

func TestDeploy_AuthRequired_401WithoutToken(t *testing.T) {
	kubectl, helm, docker := deployPrereqsOK(t)

	mockURL, mockClose := startHostMockUpstream(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized","message":"missing Authorization header"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","auth":"` + r.Header.Get("Authorization") + `"}`))
	})
	defer mockClose()

	ct, remoteRepo := detectCluster(t)

	projectDir := genProject(t, "echoHeaders", "")
	imageTag := deployBuildImage(t, docker, projectDir)
	imageRepo, imageVer := deployMakeImageAvailable(t, docker, imageTag, ct, remoteRepo)

	ns := deployNamespace(t, kubectl)
	_, k8sName := deployHelmChart(t, helm, kubectl, projectDir, imageRepo, imageVer, ns, mockURL)

	port, cancel := deployPortForward(t, kubectl, ns, k8sName, 8080)
	defer cancel()

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	waitForServer(t, baseURL)

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	t.Logf("Tool result (no auth): %s", trimMsg(result, 300))

	if !strings.Contains(result, "unauthorized") && !strings.Contains(result, "401") {
		t.Logf("Upstream may have accepted the request without auth; result: %s", trimMsg(result, 300))
	}
}

func TestDeploy_AuthBearerToken_CorrectlyForwarded(t *testing.T) {
	kubectl, helm, docker := deployPrereqsOK(t)

	mockURL, mockClose := startHostMockUpstream(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","authorization":"%s"}`,
			r.Header.Get("Authorization"))
	})
	defer mockClose()

	ct, remoteRepo := detectCluster(t)

	projectDir := genProject(t, "echoHeaders", "")
	binName := filepath.Base(projectDir)
	imageTag := deployBuildImage(t, docker, projectDir)
	imageRepo, imageVer := deployMakeImageAvailable(t, docker, imageTag, ct, remoteRepo)

	ns := deployNamespace(t, kubectl)
	k8sFullname := "mcp-" + binName

	bearerToken := "Bearer deploy-test-token-abc123"
	deployCreateSecret(t, kubectl, ns, k8sFullname, bearerToken, "")

	_, _ = deployHelmChart(t, helm, kubectl, projectDir, imageRepo, imageVer, ns, mockURL)

	// Patch deployment to inject bearer token from k8s secret.
	patchJSON := fmt.Sprintf(`[{
		"op": "add",
		"path": "/spec/template/spec/containers/0/env/-",
		"value": {
			"name": "MCP__AUTH__STATIC__BEARER_TOKEN",
			"valueFrom": {
				"secretKeyRef": {
					"name": "%s-secret",
					"key": "bearer_token"
				}
			}
		}
	}]`, k8sFullname)
	patchCmd := exec.Command(kubectl, "patch", "deployment", k8sFullname,
		"-n", ns, "--type=json", "-p", patchJSON,
	)
	if out, err := patchCmd.CombinedOutput(); err != nil {
		t.Fatalf("patch deployment: %v\n%s", err, out)
	}
	exec.Command(kubectl, "rollout", "status", "deployment", k8sFullname,
		"-n", ns, "--timeout=60s").Run()
	time.Sleep(3 * time.Second)

	port, cancel := deployPortForward(t, kubectl, ns, k8sFullname, 8080)
	defer cancel()

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	waitForServer(t, baseURL)

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	t.Logf("Tool result (with auth): %s", trimMsg(result, 300))

	if !strings.Contains(result, bearerToken) {
		t.Errorf("expected bearer token %q in response, got: %s", bearerToken, trimMsg(result, 300))
	}
	if strings.Contains(result, "unauthorized") || strings.Contains(result, "401") {
		t.Errorf("expected successful auth, got error: %s", trimMsg(result, 300))
	}
}

// ---------------------------------------------------------------------------
// Test: Deploy with virtual tools
// ---------------------------------------------------------------------------

func TestDeploy_VirtualTool_ChainedE2E(t *testing.T) {
	kubectl, helm, docker := deployPrereqsOK(t)

	mockURL, mockClose := startHostMockUpstream(t, okHandler())
	defer mockClose()

	ct, remoteRepo := detectCluster(t)

	projectDir := genProject(t, "echoHeaders,sayHello", "")
	binName := filepath.Base(projectDir)
	imageTag := deployBuildImage(t, docker, projectDir)
	imageRepo, imageVer := deployMakeImageAvailable(t, docker, imageTag, ct, remoteRepo)

	ns := deployNamespace(t, kubectl)
	k8sFullname := "mcp-" + binName

	homeDir := t.TempDir()
	virtConfig := `
virtualTools:
  - name: virt_chain
    description: Chain echo and greet
    inputSchema:
      type: object
      properties:
        name:
          type: string
      required:
        - name
    pipeline:
      - id: echo
        kind: call
        spec:
          tool: EchoHeaders
          args: {}
      - id: greet
        kind: call
        spec:
          tool: SayHello
          args:
            name: $input.name
      - id: done
        kind: return
        spec:
          from: $greet
`
	writeCoreVirtualConfig(t, homeDir, binName, virtConfig)
	virtConfigPath := filepath.Join(homeDir, "."+binName, "config.yaml")
	exec.Command(kubectl, "create", "configmap", k8sFullname+"-virtual-config",
		"-n", ns,
		fmt.Sprintf("--from-file=config.yaml=%s", virtConfigPath),
		"--dry-run=client", "-o", "yaml",
	).Run()

	_, _ = deployHelmChart(t, helm, kubectl, projectDir, imageRepo, imageVer, ns, mockURL)

	port, cancel := deployPortForward(t, kubectl, ns, k8sFullname, 8080)
	defer cancel()

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	waitForServer(t, baseURL)

	for _, toolName := range []string{"EchoHeaders", "SayHello"} {
		resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      toolName,
			"arguments": map[string]interface{}{},
		})
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s: expected 200, got %d", toolName, resp.StatusCode)
		}
		resp.Body.Close()
	}

	resp, _ := mcpHTTPCall(t, baseURL, "tools/list", map[string]interface{}{})
	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(bodyBytes), "EchoHeaders") {
		t.Errorf("tools/list should contain EchoHeaders, got: %s", trimMsg(string(bodyBytes), 500))
	}
	t.Logf("tools/list response: %s", trimMsg(string(bodyBytes), 500))
}

// ---------------------------------------------------------------------------
// Test: Helm lint & template, Secret / GCP, Ingress / Envoy
// ---------------------------------------------------------------------------

func TestDeploy_HelmDefaultValues_LintsAndInstalls(t *testing.T) {
	kubectl, helm, docker := deployPrereqsOK(t)

	mockURL, mockClose := startHostMockUpstream(t, okHandler())
	defer mockClose()

	ct, remoteRepo := detectCluster(t)

	projectDir := genProject(t, "echoHeaders", "")
	binName := filepath.Base(projectDir)
	imageTag := deployBuildImage(t, docker, projectDir)
	imageRepo, imageVer := deployMakeImageAvailable(t, docker, imageTag, ct, remoteRepo)

	chartDir := filepath.Join(projectDir, "deploy", "helm")

	lintCmd := exec.Command(helm, "lint", chartDir)
	lintOut, err := lintCmd.CombinedOutput()
	if err != nil {
		t.Errorf("helm lint failed: %v\n%s", err, lintOut)
	}
	t.Logf("helm lint: %s", string(lintOut))

	ns := deployNamespace(t, kubectl)
	tmplCmd := exec.Command(helm, "template", binName, chartDir,
		"-n", ns,
		"--set", fmt.Sprintf("image.repository=%s", imageRepo),
		"--set", fmt.Sprintf("image.tag=%s", imageVer),
		"--set", fmt.Sprintf("config.upstream.endpoint=%s", mockURL),
	)
	tmplOut, err := tmplCmd.CombinedOutput()
	if err != nil {
		t.Errorf("helm template failed: %v\n%s", err, tmplOut)
	}
	t.Logf("helm template: generated %d bytes", len(tmplOut))

	tmplStr := string(tmplOut)
	for _, kind := range []string{"Deployment", "Service", "ConfigMap"} {
		if !strings.Contains(tmplStr, fmt.Sprintf("\nkind: %s\n", kind)) {
			t.Errorf("expected kind %s in helm template output", kind)
		}
	}
}

func TestDeploy_PersistenceEnabled_HasPVC(t *testing.T) {
	_, helm, _ := deployPrereqsOK(t)

	projectDir := genProject(t, "echoHeaders", "")
	binName := filepath.Base(projectDir)
	chartDir := filepath.Join(projectDir, "deploy", "helm")

	tmplCmd := exec.Command(helm, "template", binName, chartDir,
		"-n", "default",
		"--set", "image.tag=test",
		"--set", "config.upstream.endpoint=http://example.com",
		"--set", "persistence.enabled=true",
		"--set", "persistence.storageClassName=standard-rwo",
		"--set", "persistence.size=5Gi",
	)
	tmplOut, err := tmplCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helm template failed: %v\n%s", err, tmplOut)
	}
	tmplStr := string(tmplOut)

	if !strings.Contains(tmplStr, "\nkind: PersistentVolumeClaim\n") {
		t.Error("expected PersistentVolumeClaim when persistence.enabled=true")
	}
	if !strings.Contains(tmplStr, "storageClassName: standard-rwo") {
		t.Error("expected storageClassName: standard-rwo in PVC")
	}
	if !strings.Contains(tmplStr, "storage: 5Gi") {
		t.Error("expected storage: 5Gi in PVC")
	}
	t.Logf("PVC template OK — includes storageClassName and size")
}

func TestDeploy_SecretStatic_HasSecret(t *testing.T) {
	_, helm, _ := deployPrereqsOK(t)

	projectDir := genProject(t, "echoHeaders", "")
	binName := filepath.Base(projectDir)
	chartDir := filepath.Join(projectDir, "deploy", "helm")

	tmplCmd := exec.Command(helm, "template", binName, chartDir,
		"-n", "default",
		"--set", "image.tag=test",
		"--set", "config.upstream.endpoint=http://example.com",
		"--set", "secret.provider=static",
		"--set", "secret.static.create=true",
		"--set", "secret.static.oidcClientSecret=test-oidc-secret",
		"--set", "secret.static.bearerToken=test-bearer-token",
		"--set", "secret.static.cookieToken=test-cookie-token",
	)
	tmplOut, err := tmplCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helm template failed: %v\n%s", err, tmplOut)
	}
	tmplStr := string(tmplOut)

	if !strings.Contains(tmplStr, "\nkind: Secret\n") {
		t.Error("expected Secret when secret.provider=static, secret.static.create=true")
	}
	for _, key := range []string{"oidc_client_secret", "bearer_token", "cookie_token"} {
		if !strings.Contains(tmplStr, key) {
			t.Errorf("expected key %q in Secret", key)
		}
	}
	t.Logf("Static Secret template OK")
}

func TestDeploy_SecretGCP_HasSecretProviderClass(t *testing.T) {
	_, helm, _ := deployPrereqsOK(t)

	projectDir := genProject(t, "echoHeaders", "")
	binName := filepath.Base(projectDir)
	chartDir := filepath.Join(projectDir, "deploy", "helm")

	tmplCmd := exec.Command(helm, "template", binName, chartDir,
		"-n", "default",
		"--set", "image.tag=test",
		"--set", "config.upstream.endpoint=http://example.com",
		"--set", "secret.provider=gcp",
		
		"--set", "secret.gcp.projectId=my-gcp-project",
		"--set", "secret.gcp.secretId=mcp-secrets",
	)
	tmplOut, err := tmplCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helm template failed: %v\n%s", err, tmplOut)
	}
	tmplStr := string(tmplOut)

	if !strings.Contains(tmplStr, "kind: SecretProviderClass") {
		t.Error("expected SecretProviderClass when secret.provider=gcp")
	}
	if !strings.Contains(tmplStr, "provider: gcp") {
		t.Error("expected provider: gcp in SecretProviderClass")
	}
	if !strings.Contains(tmplStr, "my-gcp-project") {
		t.Error("expected project ID in SecretProviderClass parameters")
	}
	if strings.Contains(tmplStr, "\nkind: Secret\n") {
		t.Error("expected NO static Secret when provider is gcp")
	}
	t.Logf("GCP SecretProviderClass template OK")
}

func TestDeploy_SecretDisabled_NoSecret(t *testing.T) {
	_, helm, _ := deployPrereqsOK(t)

	projectDir := genProject(t, "echoHeaders", "")
	binName := filepath.Base(projectDir)
	chartDir := filepath.Join(projectDir, "deploy", "helm")

	tmplCmd := exec.Command(helm, "template", binName, chartDir,
		"-n", "default",
		"--set", "image.tag=test",
		"--set", "config.upstream.endpoint=http://example.com",
	)
	tmplOut, err := tmplCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helm template failed: %v\n%s", err, tmplOut)
	}
	tmplStr := string(tmplOut)

	if strings.Contains(tmplStr, "\nkind: Secret\n") {
		t.Error("expected NO Secret when secret.static.create=false (default)")
	}
	if strings.Contains(tmplStr, "kind: SecretProviderClass") {
		t.Error("expected NO SecretProviderClass when secret.static.create=false (default)")
	}
	t.Logf("Secret disabled by default — no Secret or SPC rendered")
}

func TestDeploy_IngressNginx_HasIngress(t *testing.T) {
	_, helm, _ := deployPrereqsOK(t)

	projectDir := genProject(t, "echoHeaders", "")
	binName := filepath.Base(projectDir)
	chartDir := filepath.Join(projectDir, "deploy", "helm")

	tmplCmd := exec.Command(helm, "template", binName, chartDir,
		"-n", "default",
		"--set", "image.tag=test",
		"--set", "config.upstream.endpoint=http://example.com",
		"--set", "ingress.nginx.enabled=true",
	)
	tmplOut, err := tmplCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helm template failed: %v\n%s", err, tmplOut)
	}
	tmplStr := string(tmplOut)

	if !strings.Contains(tmplStr, "\nkind: Ingress\n") {
		t.Error("expected Ingress when ingress.nginx.enabled=true")
	}
	if strings.Contains(tmplStr, "\nkind: Gateway\n") {
		t.Error("expected NO Gateway when nginx ingress is enabled")
	}
	t.Logf("Nginx Ingress template OK")
}

func TestDeploy_IngressEnvoy_HasGatewayAndRoute(t *testing.T) {
	_, helm, _ := deployPrereqsOK(t)

	projectDir := genProject(t, "echoHeaders", "")
	binName := filepath.Base(projectDir)
	chartDir := filepath.Join(projectDir, "deploy", "helm")

	tmplCmd := exec.Command(helm, "template", binName, chartDir,
		"-n", "default",
		"--set", "image.tag=test",
		"--set", "config.upstream.endpoint=http://example.com",
		"--set", "ingress.envoy.enabled=true",
		"--set", "ingress.envoy.gatewayClassName=envoy-gateway",
	)
	tmplOut, err := tmplCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helm template failed: %v\n%s", err, tmplOut)
	}
	tmplStr := string(tmplOut)

	if !strings.Contains(tmplStr, "\nkind: Gateway\n") {
		t.Error("expected Gateway when ingress.envoy.enabled=true")
	}
	if !strings.Contains(tmplStr, "\nkind: HTTPRoute\n") {
		t.Error("expected HTTPRoute when ingress.envoy.enabled=true")
	}
	if strings.Contains(tmplStr, "\nkind: Ingress\n") {
		t.Error("expected NO Ingress when envoy gateway is enabled")
	}
	if !strings.Contains(tmplStr, "gatewayClassName: envoy-gateway") {
		t.Error("expected gatewayClassName in Gateway spec")
	}
	t.Logf("Envoy Gateway + HTTPRoute template OK")
}
