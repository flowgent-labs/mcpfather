# Project Conventions

## Development

- **All builds MUST use `make` from the repository root.** `make help` shows all targets.
  - `make build` — build for current GOOS/GOARCH → `bin/`
  - `make test-unit` — unit tests (`./pkg/... ./cmd/...`)
  - `make test-integration` — integration tests (`./it/...`)
  - `make build-all` — cross-compile all platforms
  - `make clean` — remove `bin/`
- **NEVER run `go build` directly** outside of `make build`.
- Binaries under `bin/` are git-ignored. Do NOT commit them.
- **Git commit messages must be concise.** Keep subject under 72 chars.
  - **No `Co-Authored-By` trailers.**
- **Any Code, config, or documentation change must follow high cohesion, low coupling.**
  Logical structure must be clear — concise without losing core logic.
  If related dependent modules exist, they MUST be updated synchronously.

## Build & Test

- **Any code change MUST pass build and tests.**
- **GFW network issues**: set `HTTPS_PROXY=http://127.0.0.1:8800` before builds that pull dependencies.
- **E2E tests requiring middleware or tokens**: `source ~/.wl4gshrc.sec` to load credentials (SONARQUBE_URL, SONARQUBE_TOKEN, DEEPSEEK_API_KEY, etc.). Do NOT inspect or read the secrets file.
- **Docker images** for CI integration tests use Aliyun CR mirrors:
  - Dex: `registry.cn-shenzhen.aliyuncs.com/wl4g/dex:v2.41.1`
  - Glauth: `registry.cn-shenzhen.aliyuncs.com/wl4g/glauth:v2.5.0`

## Test Architecture

- Unit tests: `*_test.go` files alongside source under `pkg/` and `cmd/`
- Integration tests: `it/` directory, files named `01_xxx`, `02_xxx`, etc.
- **OIDC integration tests**: Real Dex (Docker) for discovery and connectivity. Mock OIDC provider (`it/cmd/testoidc`) for `client_credentials` token exchange since Dex v2.41+ removed password grant.
- **LDAP integration tests**: All use real Glauth (Docker). No mock LDAP server.
- Generated MCP server E2E tests: generate project via `genProject()`, build via `buildServer()`, start as subprocess, verify tool calls reach mock upstream with correct auth headers.

---

# Architecture

mcpgen-go generates production-ready MCP (Model Context Protocol) servers from OpenAPI specifications.
Each OpenAPI operation becomes a typed MCP tool. Generated servers include enterprise auth, metrics, tracing, and deployment artifacts.

## Data Flow

```
OpenAPI Spec (OAS 2.x/3.x, JSON/YAML)
  │
  ├─ [converter.Parser] OAS 3.1→3.0 normalization, parsing, validation
  ├─ [converter.Converter] Path×Method → Tool (input schema, response templates, filters)
  ▼
  MCPConfig{Tools}
  │
  ├─ [generator.GenerateMCP] Code generation pipeline
  │   ├─ go.mod, main.go, server.go
  │   ├─ pkg/mcptools/ (per-tool handlers, registry)
  │   ├─ pkg/mcpserver/ (MCP server: stdio + HTTP)
  │   ├─ pkg/mcpcli/ (CLI runner)
  │   ├─ pkg/mcpvirtual/ (virtual tools engine, embedded)
  │   ├─ pkg/helpers/ (auth, config, client, metrics, trace)
  │   ├─ deploy/ (Dockerfile + Helm chart)
  │   └─ .agents/skills/virtual-tool-creator/
  ▼
  Standalone Go project (generated)
```

## Virtual Tools DSL

A declarative 5-step pipeline language composing native tools into higher-level operations:

| Step | Kind | Purpose |
|------|------|---------|
| 1 | `call` | Invoke a native MCP tool |
| 2 | `jq` | Transform data via jq expression |
| 3 | `foreach` | Iterate array with concurrent sub-pipeline |
| 4 | `emit` | Yield elements within foreach |
| 5 | `return` | Final pipeline result |

Variable references: `$input.*` (tool args), `$stepId.*` (step output), `$foreachVar` (iteration item).

## Generated Project Auth Chain

```
STATIC_TOKEN → LDAP bind → OIDC client_credentials → keychain file
```
Token precedence is resolved by `GetUpstreamToken()` in `config.templ`.
The generated server inherits all parent process env vars → `MCP__` env var overrides apply.

---

# Documentation Index

## Design & Reference

| Doc | Summary |
|-----|---------|
| [README.md](../README.md) | Project overview, quick start, CLI flags, configuration, Virtual Tools DSL, agent integration guides |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Contribution guidelines |

## Skills

| Doc | Summary |
|-----|---------|
| [.agents/skills/virtual-tool-creator/SKILL.md](skills/virtual-tool-creator/SKILL.md) | Virtual tool pipeline configuration skill for AI agents |
| [.agents/skills/virtual-tool-creator/references/bash-to-pipeline-mapping.md](skills/virtual-tool-creator/references/bash-to-pipeline-mapping.md) | Bash-to-DSL translation reference |

---

# Key Directories

```
pkg/
  converter/           OpenAPI → MCP conversion (parser, converter, schemas, OAS 3.1 preprocessor)
  generator/           Code generation engine (templates, virtual tools engine, deploy artifacts)
    templates/           Go source templates (.templ) for generated MCP servers
    mcpvirtual/          Embedded virtual tools runtime (config, engine, pipeline, node)
      config/              YAML config loading for virtual tool pipelines
      engine/              Pipeline execution engine, context/variable resolution
      node/                Step implementations: call, jq, foreach, emit, return
      pipeline/            Pipeline types, validation, tool registry interface
    deploy/              Deployment artifacts (Dockerfile, Helm chart)
    skills/              Skill resources shipped to generated projects
cmd/
  mcpgen/              CLI entry point for the code generator
  gen-config-dsl-schema/  JSON Schema generator for virtual tool DSL
it/                    Integration tests + Docker fixtures (dex, glauth)
  cmd/testoidc/          Standalone test OIDC provider (client_credentials grant)
  testdata/              OpenAPI spec fixtures (OAS 2.0, 3.0, 3.1, cyclic, binary)
  docker/                Dex OIDC + Glauth LDAP configs
examples/              Pre-generated MCP server examples (confluence, jira, sonarqube, sonatypeiq)
.agents/               Claude Code agent skills (virtual-tool-creator)
.github/workflows/     CI (pr.yml: build, unit, OIDC, LDAP, virtual-tools, core-e2e) + CD (release.yml)
```

---

# CI/CD

| Workflow | Trigger | Jobs |
|----------|---------|------|
| `pr.yml` | PR/push to main, weekly schedule | build-and-test, integration-oidc, integration-ldap, integration-virtual-tools, integration-core-e2e |
| `release.yml` | PR merge to main | Auto-version bump (semver based on commit prefix), cross-compile, GitHub Release |
