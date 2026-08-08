.PHONY: build build-all test-ut test-it install clean gen-config-dsl-schema \
	build-image push-image help

BINARY_NAME := mcpfather
CMD_PATH := ./cmd/mcpfather
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-s -w -X main.versionStr=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildDate=$(BUILD_DATE)"
BUILD_FLAGS := -v -trimpath

GOPROXY ?= https://goproxy.cn,direct
IT_GOMAXPROCS ?= 2
IT_GOFLAGS ?= -p=1
IT_TEST_FLAGS ?= -v -count=1 -timeout 600s -parallel 1
IT_BATCH_TOTAL := 12

IT_CORE_GEN_AUTH_RE := ^Test(Generator_|Auth_).*$$
IT_CORE_LOG_DOWNLOAD_RE := ^Test(Logging_|Download_).*$$
IT_CORE_UPLOAD_FORM_RE := ^Test(Upload_|Form|Multipart|FileRef).*$$
IT_CORE_CASE_CLIENT_RE := ^Test(Case|Mcpclient).*$$
IT_CORE_RUNTIME_RE := ^Test(CLI_|CyclicRef_|Regression_|E2E_Core_).*$$
IT_CORE_CONFIG_RE := ^Test(Config_|IFS_|LoggingConfig_|EnvOverride_).*$$
IT_VIRTUAL_DSL_RE := ^Test(Scenario_|DSLSchema_).*$$
IT_VIRTUAL_E2E_RE := ^Test(E2E_SonarQube_|E2E_VirtualTool_|E2E_MCP_|E2E_SonatypeIQ_|E2E_NexusFirewall_|E2E_HTTPStep_).*$$
IT_OIDC_SERVER_RE := ^TestServer_.*$$
IT_OIDC_UPSTREAM_RE := ^TestOIDC.*$$
IT_DEPLOY_FULL_RE := ^TestDeploy_FullE2E$$
IT_DEPLOY_MISC_RE := ^TestDeploy_(HelmDefaultValues|SecretGCP|IngressEnvoy).*$$

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

IMAGE_REPO ?= ghcr.io/flowgent-labs/$(BINARY_NAME)
IMAGE_TAG  ?= $(VERSION)

BIN := bin/$(BINARY_NAME)-$(GOOS)-$(GOARCH)-$(VERSION)$(if $(filter windows,$(GOOS)),.exe,)

help:
	@echo "Usage:"
	@echo "  make build                  Build $(BINARY_NAME) for current platform"
	@echo "  make build-all              Cross-compile for all platforms"
	@echo "  make build-image            Build Docker image"
	@echo "  make push-image             Build and push Docker image to ghcr.io"
	@echo "  make test-ut                 Run unit tests"
	@echo "  make test-it                 Run integration tests"
	@echo "  make install                Install $(BINARY_NAME) to GOPATH/bin"
	@echo "  make clean                  Remove build artifacts"
	@echo "  make gen-config-dsl-schema  Regenerate JSON Schema for virtual tool DSL"
	@echo ""

build:
	GOPROXY=$(GOPROXY) GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BIN) $(CMD_PATH)
	@ln -sf $(notdir $(BIN)) bin/$(BINARY_NAME)

build-all:
	GOPROXY=$(GOPROXY) GOOS=linux   GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64-$(VERSION)   $(CMD_PATH)
	GOPROXY=$(GOPROXY) GOOS=linux   GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64-$(VERSION)   $(CMD_PATH)
	GOPROXY=$(GOPROXY) GOOS=darwin  GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64-$(VERSION)  $(CMD_PATH)
	GOPROXY=$(GOPROXY) GOOS=darwin  GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64-$(VERSION)  $(CMD_PATH)
	GOPROXY=$(GOPROXY) GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64-$(VERSION).exe $(CMD_PATH)
	GOPROXY=$(GOPROXY) GOOS=windows GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-arm64-$(VERSION).exe $(CMD_PATH)

test-ut:
	GOPROXY=$(GOPROXY) go test -v -count=1 -timeout 300s ./pkg/... ./cmd/...

test-it:
	IT_BATCH='01/$(IT_BATCH_TOTAL) core-generator-auth' GOPROXY=$(GOPROXY) GOMAXPROCS=$(IT_GOMAXPROCS) GOFLAGS=$(IT_GOFLAGS) go test $(IT_TEST_FLAGS) ./it/ -run '$(IT_CORE_GEN_AUTH_RE)'
	IT_BATCH='02/$(IT_BATCH_TOTAL) core-logging-download' GOPROXY=$(GOPROXY) GOMAXPROCS=$(IT_GOMAXPROCS) GOFLAGS=$(IT_GOFLAGS) go test $(IT_TEST_FLAGS) ./it/ -run '$(IT_CORE_LOG_DOWNLOAD_RE)'
	IT_BATCH='03/$(IT_BATCH_TOTAL) core-upload-form-file' GOPROXY=$(GOPROXY) GOMAXPROCS=$(IT_GOMAXPROCS) GOFLAGS=$(IT_GOFLAGS) go test $(IT_TEST_FLAGS) ./it/ -run '$(IT_CORE_UPLOAD_FORM_RE)'
	IT_BATCH='04/$(IT_BATCH_TOTAL) core-case-mcpclient' GOPROXY=$(GOPROXY) GOMAXPROCS=$(IT_GOMAXPROCS) GOFLAGS=$(IT_GOFLAGS) go test $(IT_TEST_FLAGS) ./it/ -run '$(IT_CORE_CASE_CLIENT_RE)'
	IT_BATCH='05/$(IT_BATCH_TOTAL) core-runtime-e2e' GOPROXY=$(GOPROXY) GOMAXPROCS=$(IT_GOMAXPROCS) GOFLAGS=$(IT_GOFLAGS) go test $(IT_TEST_FLAGS) ./it/ -run '$(IT_CORE_RUNTIME_RE)'
	IT_BATCH='06/$(IT_BATCH_TOTAL) core-config-env-ifs' GOPROXY=$(GOPROXY) GOMAXPROCS=$(IT_GOMAXPROCS) GOFLAGS=$(IT_GOFLAGS) go test $(IT_TEST_FLAGS) ./it/ -run '$(IT_CORE_CONFIG_RE)'
	IT_BATCH='07/$(IT_BATCH_TOTAL) virtual-dsl-schema' GOPROXY=$(GOPROXY) GOMAXPROCS=$(IT_GOMAXPROCS) GOFLAGS=$(IT_GOFLAGS) go test $(IT_TEST_FLAGS) ./it/ -run '$(IT_VIRTUAL_DSL_RE)'
	IT_BATCH='08/$(IT_BATCH_TOTAL) virtual-e2e-usecases' GOPROXY=$(GOPROXY) GOMAXPROCS=$(IT_GOMAXPROCS) GOFLAGS=$(IT_GOFLAGS) go test $(IT_TEST_FLAGS) ./it/ -run '$(IT_VIRTUAL_E2E_RE)'
	IT_BATCH='09/$(IT_BATCH_TOTAL) oidc-resource-server' GOPROXY=$(GOPROXY) GOMAXPROCS=$(IT_GOMAXPROCS) GOFLAGS=$(IT_GOFLAGS) go test $(IT_TEST_FLAGS) ./it/ -run '$(IT_OIDC_SERVER_RE)'
	IT_BATCH='10/$(IT_BATCH_TOTAL) oidc-upstream-client' GOPROXY=$(GOPROXY) GOMAXPROCS=$(IT_GOMAXPROCS) GOFLAGS=$(IT_GOFLAGS) go test $(IT_TEST_FLAGS) ./it/ -run '$(IT_OIDC_UPSTREAM_RE)'
	IT_BATCH='11/$(IT_BATCH_TOTAL) deploy-full-e2e' GOPROXY=$(GOPROXY) GOMAXPROCS=$(IT_GOMAXPROCS) GOFLAGS=$(IT_GOFLAGS) go test $(IT_TEST_FLAGS) ./it/ -run '$(IT_DEPLOY_FULL_RE)'
	IT_BATCH='12/$(IT_BATCH_TOTAL) deploy-helm-ingress' GOPROXY=$(GOPROXY) GOMAXPROCS=$(IT_GOMAXPROCS) GOFLAGS=$(IT_GOFLAGS) go test $(IT_TEST_FLAGS) ./it/ -run '$(IT_DEPLOY_MISC_RE)'

install:
	go install $(BUILD_FLAGS) $(LDFLAGS) $(CMD_PATH)

clean:
	rm -rf bin/

# gen-config-dsl-schema regenerates the JSON Schema for virtual tool configuration
# from the Go struct definitions in internal/generator/mcpvirtual/.
# Output is written to the skill resources directory for use by the virtual-tool-creator skill.
gen-config-dsl-schema:
	@mkdir -p bin
	@go build -o bin/gen-config-dsl-schema ./cmd/gen-config-dsl-schema/
	@./bin/gen-config-dsl-schema --output .agents/skills/virtual-tool-creator/resources/dsl-schema.json
	@echo "==> Schema updated: .agents/skills/virtual-tool-creator/resources/dsl-schema.json"

build-image: build
	docker build -t $(IMAGE_REPO):$(IMAGE_TAG) .

push-image: build-image
	docker push $(IMAGE_REPO):$(IMAGE_TAG)
