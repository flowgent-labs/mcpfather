# Project Conventions

- **All builds use `make` from the repository root.** `make help` for targets.
  `make build`, `make test-unit`, `make test-integration`, `make build-all`, `make clean`.
- **Git commit messages must be concise.** Subject under 72 chars. **No `Co-Authored-By` trailers.**
- **Any code, config, or documentation change must follow high cohesion, low coupling.**
  Structure must be clear and concise. Related dependent modules MUST be updated synchronously.
- **Code changes MUST pass `make test-unit`.**

## Test Architecture (non-obvious from code alone)

- **OIDC**: Real Keycloak 26.7.0 (Docker) for discovery, connectivity, client_credentials,
  and device_code (RFC 8628) tests. Mock provider (`it/cmd/mockoidcsvc`) for negative
  test cases (expired token, wrong audience, HS256 forgery) via custom /sign endpoint.
- **E2E secrets**: `source ~/.wl4gshrc.sec` to load SonarQube/Deeplake tokens.
- See also: [examples/E2E-Sonarqube-VirtualTools.md](../examples/E2E-Sonarqube-VirtualTools.md)

---

# Documentation Index

| Doc | Summary |
|-----|---------|
| [README.md](../README.md) | Project overview, quick start, CLI flags, config, Virtual Tools DSL, agent integration |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Contribution guidelines |
| [examples/E2E-Sonarqube-VirtualTools.md](../examples/E2E-Sonarqube-VirtualTools.md) | Full E2E test walkthrough: virtual tools against real SonarQube, bug discoveries, design lessons |
| [.agents/skills/virtual-tool-creator/SKILL.md](skills/virtual-tool-creator/SKILL.md) | Virtual tool pipeline configuration skill for AI agents |
| [.agents/skills/virtual-tool-creator/references/bash-to-pipeline-mapping.md](skills/virtual-tool-creator/references/bash-to-pipeline-mapping.md) | Bash→DSL translation reference |

---

# Key Directories

```
pkg/
  converter/           OpenAPI → MCP: parser, converter, schemas, OAS 3.1 preprocessor
  generator/           Code generation engine (templates, virtual tools, deploy artifacts)
    templates/           Go source templates (.templ) for generated MCP servers
    mcpvirtual/          Embedded virtual tools runtime
      config/              YAML config loading
      engine/              Pipeline executor, context/variable resolution
      node/                Step kinds: call, jq, foreach, emit, return
      pipeline/            DSL types, validation, tool registry
    deploy/              Dockerfile + Helm chart (embedded, copied to generated projects)
cmd/
  mcpfather/              CLI entry point
  gen-config-dsl-schema/  JSON Schema generator for virtual tool DSL
it/                    Integration tests + Docker fixtures (keycloak)
  cmd/mockoidcsvc/       Standalone OIDC provider for custom-claims token tests
  testdata/              OpenAPI spec fixtures
  docker/                Keycloak OIDC configs
examples/              Pre-generated MCP server examples (confluence, jira, sonarqube, sonatypeiq)
.agents/               Claude Code agent skills
.github/workflows/     CI (build, unit, OIDC/virtual-tools/core-e2e) + CD (release)
```
