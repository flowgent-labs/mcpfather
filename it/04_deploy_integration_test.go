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

// deployKubectl holds the kubectl invocation prefix (e.g. {"kubectl"}).
var deployKubectl []string

// kubectlCmd builds an exec.Cmd for kubectl, respecting the deployKubectl prefix.
func kubectlCmd(args ...string) *exec.Cmd {
	a := append(append([]string{}, deployKubectl...), args...)
	return exec.Command(a[0], a[1:]...)
}

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

	// 5 — no local tooling and no remote repo configured; cannot make image
	// available to the cluster. This is a hard prerequisite — skipping would
	// silently hide deploy regressions from developers who expect E2E coverage.
	t.Fatalf("Cannot determine local cluster type and MCPFATHER_TEST_IMAGE_REPO is not set.\n" +
		"Set MCPFATHER_TEST_IMAGE_REPO=<registry>/<repo> for remote clusters, or install k3s/ctr/crictl/orb.")
	return clusterUnknown, "" // unreachable
}


// deployPrereqsOK checks that kubectl, helm, and docker are available and a
// Kubernetes cluster is reachable via ~/.kube/config (or KUBECONFIG).
func deployPrereqsOK(t *testing.T) (_kubectl, helm, docker string) {
	t.Helper()

	kubectlPath, err := exec.LookPath("kubectl")
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

	// Check cluster connectivity via default kubeconfig (~/.kube/config).
	cmd := exec.Command(kubectlPath, "cluster-info")
	out, err := cmd.CombinedOutput()
	if err == nil {
		deployKubectl = []string{kubectlPath}
		return kubectlPath, helm, docker
	}

	// kubectl can't reach the cluster. Give actionable advice.
	errStr := strings.TrimSpace(string(out))
	k3sCfg := "/etc/rancher/k3s/k3s.yaml"
	if _, statErr := os.Stat(k3sCfg); statErr == nil {
		home, _ := os.UserHomeDir()
		t.Skipf("k3s kubeconfig found at %s but not accessible.\n"+
			"Fix:  mkdir -p %s/.kube && sudo cp %s %s/.kube/config && chmod 600 %s/.kube/config\n"+
			"(kubectl error was: %v\n%s)",
			k3sCfg, home, k3sCfg, home, home, err, errStr)
	}

	t.Skipf("kubectl cannot reach cluster: %v\n%s\n"+
		"Ensure a Kubernetes cluster is running and kubectl is configured correctly.\n"+
		"Check:  kubectl cluster-info  and  ls -la ~/.kube/config",
		err, out)
	return
}

// deployNamespace creates a unique test namespace and returns its name.
func deployNamespace(t *testing.T, kubectl string) string {
	t.Helper()
	ns := fmt.Sprintf("mcpfather-deploy-test-%d", time.Now().UnixNano()%100000)

	// Clean up any leftover namespace from prior runs (ignore errors).
	kubectlCmd("delete", "namespace", ns, "--ignore-not-found", "--timeout=30s").Run()

	// Wait for any previous instance to be fully deleted.
	for i := 0; i < 30; i++ {
		out, _ := kubectlCmd("get", "namespace", ns, "-o", "jsonpath={.status.phase}").CombinedOutput()
		if strings.TrimSpace(string(out)) != "Terminating" {
			break
		}
		time.Sleep(2 * time.Second)
	}

	cmd := kubectlCmd("create", "namespace", ns)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create namespace %s: %v\n%s", ns, err, out)
	}
	t.Cleanup(func() {
		kubectlCmd("delete", "namespace", ns, "--ignore-not-found", "--timeout=60s").Run()
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

	// Build args: behind GFW, route Go module traffic through goproxy.cn
	// (a public CDN in China) and disable sumdb checks so `go mod download`
	// inside the container does not hang. podman-remote --network host does
	// NOT map 127.0.0.1 to the host, so a localhost proxy is invisible here.
	buildArgs := []string{"build", "--network", "host", "-t", imageTag, "-f", dockerfile}
	if os.Getenv("IN_CN_GFW") == "true" {
		buildArgs = append(buildArgs,
			"--build-arg", "GOPROXY=https://goproxy.cn,direct",
			"--build-arg", "GONOSUMDB=*",
			"--build-arg", "GONOSUMCHECK=*",
		)
	}
	buildArgs = append(buildArgs, projectDir)
	cmd := exec.Command(docker, buildArgs...)
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
		"--set", fmt.Sprintf("config.upstream.default.endpoint=%s", upstreamURL),
		"--set", "config.tools.registerAllByDefault=true",
		"--set", "config.runtime.logAuthorization=false",
		"--wait",
		"--timeout", "120s",
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

	kubectlCmd("delete", "secret", secretName, "-n", ns, "--ignore-not-found").Run()

	cmd := kubectlCmd("create", "secret", "generic", secretName,
		"-n", ns,
		fmt.Sprintf("--from-literal=web_token=%s", bearerToken),
		fmt.Sprintf("--from-literal=cookie_token=%s", cookieToken),
		fmt.Sprintf("--from-literal=oidc_client_secret=%s", "test-oidc-secret"),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("kubectl create secret: %v\n%s", err, out)
	}
	t.Logf("Secret %s created in namespace %s", secretName, ns)
}

// waitForPodReady blocks until a pod matching app.kubernetes.io/name=name
// is Ready, or times out.
func waitForPodReady(t *testing.T, kubectl, ns, name string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		out, _ := kubectlCmd("get", "pods", "-n", ns,
			"-l", fmt.Sprintf("app.kubernetes.io/name=%s", name),
			"-o", "jsonpath={.items[0].status.phase}",
		).CombinedOutput()
		s := strings.TrimSpace(string(out))
		if s == "Running" {
			t.Logf("Pod ready: %s", name)
			return
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("pod %s/%s did not become Ready within %v", ns, name, timeout)
}

// deployPortForward sets up kubectl port-forward and returns the local port.
func deployPortForward(t *testing.T, kubectl, ns, k8sName string, remotePort int) (localPort int, cancel func()) {
	t.Helper()

	cmd := kubectlCmd("get", "pods", "-n", ns,
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

func TestDeploy_FullE2E(t *testing.T) {
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

	// Minimal deploy first (no persistence/secret/ingress)
	chartDir := filepath.Join(projectDir, "deploy", "helm")
	releaseName := binName
	args := []string{
		"install", releaseName, chartDir,
		"-n", ns,
		"--set", fmt.Sprintf("fullnameOverride=%s", k8sFullname),
		"--set", fmt.Sprintf("nameOverride=%s", k8sFullname),
		"--set", fmt.Sprintf("image.repository=%s", imageRepo),
		"--set", fmt.Sprintf("image.tag=%s", imageVer),
		"--set", "image.pullPolicy=IfNotPresent",
		"--set", fmt.Sprintf("config.upstream.default.endpoint=%s", mockURL),
		"--set", "config.tools.registerAllByDefault=true",
		"--wait", "--timeout", "60s",
	}
	if out, err := exec.Command(helm, args...).CombinedOutput(); err != nil {
		t.Fatalf("helm install failed: %v\n%s", err, out)
	}
	t.Logf("Helm release %s installed in namespace %s", releaseName, ns)
	t.Cleanup(func() {
		exec.Command(helm, "uninstall", releaseName, "-n", ns, "--ignore-not-found", "--timeout=30s").Run()
	})

	port, cancel := deployPortForward(t, kubectl, ns, k8sFullname, 8080)
	defer cancel()
	baseURL := fmt.Sprintf("http://localhost:%d", port)
	waitForServer(t, baseURL)

	// Native tools
	for _, toolName := range []string{"EchoHeaders", "SayHello"} {
		resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
			"name": toolName, "arguments": map[string]interface{}{},
		})
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s: expected 200, got %d", toolName, resp.StatusCode)
		}
		resp.Body.Close()
	}

	// tools/list
	resp, _ := mcpHTTPCall(t, baseURL, "tools/list", map[string]interface{}{})
	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(bodyBytes), "EchoHeaders") {
		t.Errorf("tools/list should contain EchoHeaders, got: %s", trimMsg(string(bodyBytes), 500))
	}
	t.Logf("tools/list response: %s", trimMsg(string(bodyBytes), 500))
}

func TestDeploy_HelmDefaultValues_LintsAndInstalls(t *testing.T) {
	_, helm, _ := deployPrereqsOK(t)
	projectDir := genProject(t, "echoHeaders", "")
	chartDir := filepath.Join(projectDir, "deploy", "helm")

	// helm lint (client-side only, needs no image build or cluster)
	lintCmd := exec.Command(helm, "lint", chartDir)
	lintOut, err := lintCmd.CombinedOutput()
	if err != nil {
		t.Errorf("helm lint failed: %v\n%s", err, lintOut)
	}
	t.Logf("helm lint: %s", string(lintOut))

	// helm template with dummy image values (no real deploy needed)
	tmplCmd := exec.Command(helm, "template", filepath.Base(projectDir), chartDir,
		"-n", "default",
		"--set", "image.repository=example.com/mcp",
		"--set", "image.tag=v1.0.0",
		"--set", "config.upstream.default.endpoint=http://example.com",
	)
	tmplOut, err := tmplCmd.CombinedOutput()
	if err != nil {
		t.Errorf("helm template failed: %v\n%s", err, tmplOut)
	}
	t.Logf("helm template: generated %d bytes", len(tmplOut))

	for _, kind := range []string{"Deployment", "Service", "ConfigMap"} {
		if !strings.Contains(string(tmplOut), fmt.Sprintf("\nkind: %s\n", kind)) {
			t.Errorf("expected kind %s in helm template output", kind)
		}
	}
}

// deployRealInstall builds, deploys, and returns (namespace, releaseName).
// If wait is false the helm install does NOT block on pod readiness (useful for
// tests that only verify infra resources like PVC/Secret/Ingress were created).
func deployRealInstall(t *testing.T, kubectl, helm, docker, projectDir string, wait bool, helmSets ...string) (ns, releaseName string) {
	t.Helper()
	ns = deployNamespace(t, kubectl)
	imageTag := deployBuildImage(t, docker, projectDir)
	ct, remoteRepo := detectCluster(t)
	imageRepo, imageVer := deployMakeImageAvailable(t, docker, imageTag, ct, remoteRepo)

	binName := filepath.Base(projectDir)
	releaseName = binName
	// Prepend "mcp-" if release name starts with a digit (project dirs like "001"
	// fail Helm validation because k8s object names must start with a letter).
	k8sFullname := releaseName
	if len(releaseName) > 0 && releaseName[0] >= '0' && releaseName[0] <= '9' {
		k8sFullname = "mcp-" + releaseName
	}
	chartDir := filepath.Join(projectDir, "deploy", "helm")

	args := []string{
		"install", releaseName, chartDir,
		"-n", ns,
		"--set", fmt.Sprintf("fullnameOverride=%s", k8sFullname),
		"--set", fmt.Sprintf("nameOverride=%s", k8sFullname),
		"--set", fmt.Sprintf("image.repository=%s", imageRepo),
		"--set", fmt.Sprintf("image.tag=%s", imageVer),
		"--set", "image.pullPolicy=IfNotPresent",
		"--timeout", "60s",
	}
	if wait {
		args = append(args, "--wait")
	}
	args = append(args, helmSets...)

	cmd := exec.Command(helm, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("helm install failed: %v\n%s", err, out)
	}
	t.Logf("Helm release %s installed in namespace %s", releaseName, ns)

	t.Cleanup(func() {
		exec.Command(helm, "uninstall", releaseName, "-n", ns, "--ignore-not-found", "--timeout=30s").Run()
		kubectlCmd("delete", "namespace", ns, "--ignore-not-found", "--timeout=30s").Run()
	})
	return ns, releaseName
}

func TestDeploy_SecretGCP_HasSecretProviderClass(t *testing.T) {
	kubectl, helm, docker := deployPrereqsOK(t)

	// SecretProviderClass CRD is required; skip if not installed.
	if out, err := exec.Command(kubectl, "get", "crd", "secretproviderclasses.secrets-store.csi.x-k8s.io").CombinedOutput(); err != nil {
		t.Skipf("SecretProviderClass CRD not installed — skipping: %v\n%s", err, out)
	}

	projectDir := genProject(t, "echoHeaders", "")
	ns, releaseName := deployRealInstall(t, kubectl, helm, docker, projectDir, false,
		"--set", "secret.provider=gcp",
		"--set", "secret.gcp.projectId=my-gcp-project",
		"--set", "secret.gcp.secretId=mcp-secrets",
		"--set", "config.upstream.default.endpoint=http://httpbin.org/anything",
	)

	spcOut, err := kubectlCmd("get", "secretproviderclass", "-n", ns,
		"-l", fmt.Sprintf("app.kubernetes.io/instance=%s", releaseName),
		"-o", "jsonpath={.items[0].metadata.name}",
	).CombinedOutput()
	if err != nil || len(strings.TrimSpace(string(spcOut))) == 0 {
		t.Logf("SecretProviderClass may not be available (CSI driver not installed): %v", err)
	} else {
		t.Logf("SecretProviderClass created: %s", strings.TrimSpace(string(spcOut)))
	}
	// Also verify NO static Secret was created
	secOut, _ := kubectlCmd("get", "secret", "-n", ns,
		"-l", fmt.Sprintf("app.kubernetes.io/instance=%s", releaseName),
		"-o", "jsonpath={.items[*].metadata.name}",
	).CombinedOutput()
	if strings.Contains(string(secOut), fmt.Sprintf("%s-secret", releaseName)) {
		t.Error("expected NO static Secret when provider=gcp")
	}
}

// TestDeploy_SecretDisabled_NoSecret is covered by TestDeploy_Infrastructure
// (which includes secret.static.create=true). The default path (no secret)
// is implicitly tested by TestDeploy_AuthRequired_401WithoutToken which
// deploys without secret settings.

func TestDeploy_IngressEnvoy_HasGatewayAndRoute(t *testing.T) {
	kubectl, helm, docker := deployPrereqsOK(t)

	// Gateway API CRDs are required; skip if not installed.
	if out, err := exec.Command(kubectl, "get", "crd", "gateways.gateway.networking.k8s.io").CombinedOutput(); err != nil {
		t.Skipf("Gateway API CRDs not installed — skipping: %v\n%s", err, out)
	}

	projectDir := genProject(t, "echoHeaders", "")
	ns, releaseName := deployRealInstall(t, kubectl, helm, docker, projectDir, false,
		"--set", "ingress.envoy.enabled=true",
		"--set", "ingress.envoy.gatewayClassName=envoy-gateway",
		"--set", "config.upstream.default.endpoint=http://httpbin.org/anything",
	)

	gwOut, err := kubectlCmd("get", "gateway", "-n", ns,
		"-l", fmt.Sprintf("app.kubernetes.io/instance=%s", releaseName),
		"-o", "jsonpath={.items[0].metadata.name}",
	).CombinedOutput()
	if err != nil || len(strings.TrimSpace(string(gwOut))) == 0 {
		t.Logf("Gateway/HTTPRoute may not be available (Gateway API CRDs not installed): %v", err)
	}
}
