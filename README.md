# Enhanced Enterprise-grade MCP server Generator from OpenAPI Schema

Generate production-ready Model Context Protocol (MCP) servers from OpenAPI specs. Each API operation becomes an AI tool that forwards requests to your upstream service.

## Features

- **OAS 2.x / 3.x support** — JSON or YAML, one-command generation. Every OpenAPI operation becomes a typed MCP tool with structured input/output schemas.
- **Virtual Tools** — Declaratively compose native tools into pipelines (e.g `call → jq → foreach → emit → return`). Drastically reduces LLM token consumption for multi-step API workflows.
- **Enterprise Auth** — OIDC (client credentials + password grants), LDAP service-account bind, and static bearer/cookie tokens. Auth credentials are decoupled and automatically injected into upstream calls.
- **Prometheus Metrics** — Standard `mcp_tool_call_duration_seconds` histogram exported for every native and virtual tool invocation, with configurable boundaries and static labels.
- **OTel Distributed Tracing** — Optional OpenTelemetry tracing via OTLP gRPC (`-tags otel`). W3C trace context is propagated to upstream APIs for end-to-end visibility across agent teams.
- **Deployment Artifacts** — Generated projects ship with Dockerfile, Helm chart, Makefile targets, and K8s manifests (ConfigMap, Secret, HPA, Ingress/Gateway, SecretProviderClass for GCP).

## Quick Start

### Build Generator

```sh
make
```

### Generate MCP server (e.g: Confluence)

```sh
./bin/mcpgen -v \
    -i examples/swaggers/confluence-server-v10.2.14.oas.v3.0.1.json \
    -o examples/confluence-mcp \
    --includes "listSpaces,createPage,updatePage,deletePage"
```

### Start in `HTTP` mode (example)

- The server defaults to httpbin.org which echoes requests — great for quick verification:

```sh
export MCP__UPSTREAM__ENDPOINT=https://api.example.com

# Optional 1: setup token from env
export MCP__AUTH__STATIC__BEARER_TOKEN=your-token
examples/confluence-mcp/bin/confluence-mcp -v 10 --transport http --port 8080

# Optional 2: setup token from file (e.g: echo -n "YOUR_TOKEN" > /path/to/.credentials)
export MCP__AUTH__STATIC__BEARER_TOKEN_FILE=/path/to/.credentials
examples/confluence-mcp/bin/confluence-mcp -v 10 --transport http --port 8080
```

- Test with `mcpclient.sh` (for `HTTP` transport only)

```sh
./mcpclient.sh list-tools
./mcpclient.sh call GetPage '{"id": "123456"}'
```

### Usage in `CLI` Mode (example)

Invoke tools directly from the command line — no MCP agent needed. Useful for debugging, scripting, and manual API exploration. The CLI reuses the same `mcptools` handlers as the MCP server, so every call makes a real HTTP request upstream.

```sh
# Set your upstream endpoint (required for real API calls)
export MCP__UPSTREAM__ENDPOINT=https://api.example.com
export MCP__AUTH__STATIC__BEARER_TOKEN=your-token

# First call: list available tools
examples/confluence-mcp/bin/confluence-mcp -t cli list

# First tool call: fetch a page by ID
examples/confluence-mcp/bin/confluence-mcp -t cli Getpage --id 123456

# Show tool-specific help (GNU-style usage)
examples/confluence-mcp/bin/confluence-mcp -t cli Getpage --help

# Call a tool with GNU-style --flag arguments
examples/confluence-mcp/bin/confluence-mcp -t cli ListSpaces --limit=5 --type global
examples/confluence-mcp/bin/confluence-mcp -t cli SearchContent --cql 'type=page AND text~"API"' --limit 10

# Call a tool without arguments (for tools that have no required params)
examples/confluence-mcp/bin/confluence-mcp -t cli ListSpaces
```

## Generator Self Configuration

```sh
./bin/mcpgen -i spec.yaml -o output-dir [--includes op1,op2] [--excludes op3] [-v]
```

| Flag | Description | Example |
|---|---|---|
| `-i, --input` | Path to the OpenAPI specification file (JSON or YAML) | `spec.yaml` |
| `-o, --output` | Path to the output MCP server directory | `./my-mcp` |
| `--includes` | Comma-separated `operationId` values to generate (omit for all) | `listSpaces,createPage` |
| `--excludes` | Comma-separated `operationId` values to skip | `healthCheck,status` |
| `-v, --verbose` | Print step-by-step generation details | |
| `--validation` | Enable OpenAPI schema validation | |

Values are matched against the `operationId` field in the OpenAPI spec (exact string match). An `operationId` appearing in both `--includes` and `--excludes` triggers an error.

### Filtering

Use `--includes` and `--excludes` to control which operations generate MCP tools. Values are the `operationId` strings from your OpenAPI spec.

- for example - Generate the Atlassian Confluence MCP server

```sh
# Only generate tools for specific operations
./bin/mcpgen -i examples/swaggers/confluence-server-v10.2.14.oas.v3.0.1.json \
    -o examples/confluence-mcp --includes "listSpaces,createPage,getSpaceContent"

# Generate all tools except health checks
./bin/mcpgen -i examples/swaggers/confluence-server-v10.2.14.oas.v3.0.1.json \
    -o examples/confluence-mcp --excludes "healthCheck,status"

# Generate all tools except a few
./bin/mcpgen -i examples/swaggers/confluence-server-v10.2.14.oas.v3.0.1.json \
    -o examples/confluence-mcp --excludes "uploadAttachment,removeLabel"

# Preview what gets included/excluded
./bin/mcpgen -i examples/swaggers/confluence-server-v10.2.14.oas.v3.0.1.json \
    -o examples/confluence-mcp --includes "listSpaces" -v
```

### Tool name truncation

Long `operationId` values are automatically truncated to 125 characters with a hash suffix to preserve uniqueness, ensuring compatibility with MCP tool name limits.

## Generated MCP Server - Configuration

| Flag | Description | Default |
|---|---|---|
| `--transport <stdio\|http\|cli>` | Transport mode | `stdio` |
| `-p, --port <number>` | HTTP server port | `8080` |
| `-v, --verbose <0-10>` | Request logging verbosity | `0` |
| `--print-default-config` | Print default config.yaml to stdout and exit | |

### Logging levels

| Level | Output |
|---|---|
| `0` | Silent |
| `1` | HTTP access log: `[http] sid=- 200 POST /mcp (1ms)` |
| `2` | MCP request log: `[mcp] tool=SearchContent args={...}`, upstream method + URL |
| `3` | + upstream query params |
| `5` | + request/response headers |
| `7` | + request/response body |
| `9` | + pretty-printed JSON body |
| `10` | Same as 9 (full debug) |

### Environment variables

| Variable | Description |
|---|---|
| `MCP__UPSTREAM__ENDPOINT` | Base URL of the upstream API (default: `https://httpbin.org/anything`) |
| `MCP__AUTH__STATIC__BEARER_TOKEN` | Bearer token for upstream auth (fallback when no Authorization header from client) |
| `MCP__AUTH__STATIC__BEARER_TOKEN_FILE` | Path to a file containing the bearer token (alternative to `MCP__AUTH__STATIC__BEARER_TOKEN`) |
| `MCP__AUTH__STATIC__COOKIE_TOKEN` | Legacy cookie token for upstream auth (fallback when no Authorization header from client) |
| `MCP__AUTH__STATIC__COOKIE_TOKEN_FILE` | Path to a file containing the Legacy cookie token (alternative to `MCP__AUTH__STATIC__COOKIE_TOKEN`) |

### Token retrieval priority

> The server tries to obtain a Bearer token in this order:

1. Authorization header from the client's HTTP request (forwarded)
2. `MCP__AUTH__STATIC__BEARER_TOKEN` environment variable
3. `MCP__AUTH__STATIC__BEARER_TOKEN_FILE` (read from file — ideal for Kubernetes secrets)
4. macOS Keychain (`security find-generic-password -s mcpgen-upstream -wa ""`)
5. Windows Credential Manager (`cmdkey /get:mcpgen-upstream`)

### Token format

- The token value is inspected for a recognized prefix. If the value already starts with `Bearer ` or `Basic ` (case-insensitive), it is used as-is in the `Authorization` header. Otherwise, `Bearer ` is automatically prepended.

### Tool filtering

- For specs with many operations, limit which tools AI agents can discover via an optional config file:

```sh
# Print the default config template
examples/confluence-mcp/bin/confluence-mcp --print-default-config

# Edit: ~/.confluence-mcp/config.yaml and list only the tools you want
```

- `$HOME/.{binaryName}/config.yaml`:

```yaml
# ---- Native MCP Tools ----
tools:
  expose:
    # When true, all native tools are registered by default
    # (individual tools can be hidden via excludes).
    # When false (the default), only tools listed in includes
    # are exposed.
    register-all-tools-by-default: false

    # Explicitly expose these tools (operationId values).
    # Takes precedence: tools listed here are always exposed,
    # even if register-all-tools-by-default is false.
    includes:
      - ListSpaces
      - SearchContent

    # Explicitly hide these tools from external agents (operationId values).
    # Takes precedence over register-all-tools-by-default.
    # A tool listed in BOTH includes and excludes will cause
    # the server to fail at startup.
    excludes: []
```

- When `tools.expose.include` is non-empty, only those tools are registered with the MCP server and shown in `-t cli list`. When absent or empty, all tools are available.

### Virtual Tools (Composition)

- Compose multiple native tools into a single AI-callable tool via a declarative 5-step pipeline DSL. Add a `virtualTools` list to your config file:

- Real Enterprise Virtual Tools definitions:
  - [sonarqube-example-config.yaml](.agents/skills/virtual-tool-creator/resources/sonarqube-example-config.yaml)
  - [sonatypeiq-example-config.yaml](.agents/skills/virtual-tool-creator/resources/sonatypeiq-example-config.yaml)

`$HOME/.{BIN_NAME}/config.yaml`:

```yaml
# ---- Virtual Tools ----
# Compose multiple native tools into a single virtual tool via a declarative
# e.g: 5-step pipeline (call → jq → foreach → emit → return).
# Schema: https://github.com/flowgent-labs/mcpgen-go/blob/main/.agents/skills/virtual-tool-creator/resources/dsl-schema.json
virtualTools:
  - id: getData
    kind: call
    spec:
      tool: GetData
      parse: json
      args:
        id: $input.dataId
  - id: enrich
    kind: foreach
    spec:
      in: $getData.items
      as: item
      concurrency: 8
      preserveOrder: true
      pipeline:
        - id: getDetail
          kind: call
          spec:
            tool: GetItemDetail
            args:
              id: $item.id
        - id: emitEnriched
          kind: emit
          spec:
            from: $item
            vars:
              detail: $getDetail
            expr: '. + {detail: $detail}'
  - id: done
    kind: return
    spec:
      from: $enrich
```

- Pipeline step kinds: `call` (invoke an MCP tool), `jq` (jq expression transform), `foreach` (concurrent iteration over arrays), `emit` (output within foreach), and `return` (final result). Full documentation in [.agents/skills/virtual-tool-creator/](.agents/skills/virtual-tool-creator/).

## Generated MCP Server - Agent Integration

### OpenCode

- `~/.config/opencode/opencode.json`:

```json
{
  "mcp": {
    "confluence-mcp": {
      "type": "local",
      "command": ["bash", "-c", "/path/to/confluence-mcp"],
      "args": ["--transport", "stdio"],
      "env": {
        "MCP__UPSTREAM__ENDPOINT": "https://api.example.com",
        "MCP__AUTH__STATIC__BEARER_TOKEN": "your-token",
        "MCP__AUTH__STATIC__BEARER_TOKEN_FILE": "/path/to/fallback/.credentials"
      },
      "enabled": true
    }
  }
}
```

### Claude Code

- `~/.claude.json`:

```json
{
  "mcpServers": {
    "confluence-mcp": {
      "command": "/path/to/confluence-mcp",
      "args": ["--transport", "stdio"],
      "env": {
        "MCP__UPSTREAM__ENDPOINT": "https://api.example.com",
        "MCP__AUTH__STATIC__BEARER_TOKEN": "your-token",
        "MCP__AUTH__STATIC__BEARER_TOKEN_FILE": "/path/to/fallback/.credentials"
      }
    }
  }
}
```

### Codex CLI

- `~/.codex/config.toml`:

```toml
[mcp_servers.confluence-mcp]
url = "http://localhost:8080/mcp"
bearer_token_env_var = "MCP__AUTH__STATIC__BEARER_TOKEN"
```

### Cursor

- `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "confluence-mcp": {
      "command": "/path/to/confluence-mcp",
      "args": ["--transport", "stdio"],
      "env": {
        "MCP__UPSTREAM__ENDPOINT": "https://api.example.com",
        "MCP__AUTH__STATIC__BEARER_TOKEN": "your-token",
        "MCP__AUTH__STATIC__BEARER_TOKEN_FILE": "/path/to/fallback/.credentials"
      }
    }
  }
}
```

### OpenCode (Remote)

- `~/.config/opencode/opencode.json`:

```json
{
  "mcp": {
    "confluence-mcp": {
      "type": "remote",
      "url": "http://localhost:8080/mcp",
      "env": {
        "MCP__UPSTREAM__ENDPOINT": "https://api.example.com",
        "MCP__AUTH__STATIC__BEARER_TOKEN": "your-token",
        "MCP__AUTH__STATIC__BEARER_TOKEN_FILE": "/path/to/fallback/.credentials"
      }
    }
  }
}
```

### Claude Code (Remote)

- `~/.claude.json`:

```json
{
  "mcpServers": {
    "confluence-mcp": {
      "url": "http://localhost:8080/mcp",
      "env": {
        "MCP__UPSTREAM__ENDPOINT": "https://api.example.com",
        "MCP__AUTH__STATIC__BEARER_TOKEN": "your-token",
        "MCP__AUTH__STATIC__BEARER_TOKEN_FILE": "/path/to/fallback/.credentials"
      }
    }
  }
}
```

### Codex CLI (Remote)

- `~/.codex/config.toml`:

```toml
[mcp_servers.confluence-mcp]
url = "http://localhost:8080/mcp"
bearer_token_env_var = "MCP__AUTH__STATIC__BEARER_TOKEN"
```

### Cursor (Remote)

- `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "confluence-mcp": {
      "url": "http://localhost:8080/mcp",
      "env": {
        "MCP__UPSTREAM__ENDPOINT": "https://api.example.com",
        "MCP__AUTH__STATIC__BEARER_TOKEN": "your-token",
        "MCP__AUTH__STATIC__BEARER_TOKEN_FILE": "/path/to/fallback/.credentials"
      }
    }
  }
}
```

## References swaggers

### Atlassian - Jira

- Server edition (more: [developer.atlassian.com/server](https://developer.atlassian.com/server))
  - [jira_software_dc_10007_swagger.v3.json](https://dac-static.atlassian.com/server/jira/platform/jira_software_dc_10007_swagger.v3.json) (v10.7.4)
  - [jira_software_dc_11002_swagger.v3.json](https://dac-static.atlassian.com/server/jira/platform/jira_software_dc_11002_swagger.v3.json) (v11.2.1)
    - Other MCP refer: [context7 OpenAPI](https://context7.com/openapi/dac-static_atlassian_server_jira_platform_jira_software_dc_11002_swagger_v3_json)
    - Older specs refer: [jira-rest-plugin.wadl](https://docs.atlassian.com/jira/REST/server/jira-rest-plugin.wadl)

- Cloud edition (More: [developer.atlassian.com/cloud](https://developer.atlassian.com/cloud))
  - [Jira Software REST API intro](https://developer.atlassian.com/cloud/jira/software/rest/intro/#introduction)
  - [swagger.v3.json](https://dac-static.atlassian.com/cloud/jira/software/swagger.v3.json)

### Atlassian - Confluence

- Server edition (More: [developer.atlassian.com/server](https://developer.atlassian.com/server))
  - [Confluence REST v10.2.14 intro](https://developer.atlassian.com/server/confluence/rest/v10214/intro/#about)
  - [10.2.14.swagger.v3.json](https://dac-static.atlassian.com/server/confluence/10.2.14.swagger.v3.json)
    - more docs: [developer.atlassian.com/cloud](https://developer.atlassian.com/cloud)

- Cloud edition (More: [developer.atlassian.com/cloud](https://developer.atlassian.com/cloud))
  - [Confluence REST v2 intro](https://developer.atlassian.com/cloud/confluence/rest/v2/intro/)
  - [openapi-v2.v3.json](https://dac-static.atlassian.com/cloud/confluence/openapi-v2.v3.json)

### Sonatype - IQ

- [IQ API reference](https://help.sonatype.com/en/iq-api-reference.html)
- [iq-api.json (latest)](https://sonatype.github.io/sonatype-documentation/api/iq/latest/iq-api.json)
- [iq-api.json (1.204.2-01)](https://sonatype.github.io/sonatype-documentation/api/iq/1.204.2-01/iq-api.json)
- [iq-api.json (1.203.0-01)](https://sonatype.github.io/sonatype-documentation/api/iq/1.203.0-01/iq-api.json)

### Sonatype - Nexus Repository

- [Nexus Repository API reference](https://help.sonatype.com/en/api-reference.html)
- [nexus-repository-api.json (latest)](https://sonatype.github.io/sonatype-documentation/api/nexus-repository/latest/nexus-repository-api.json)

### Sonarqube (*Not support swagger*)

- [SonarQube API (Schema) - Recommends](https://next.sonarqube.com/sonarqube/api/webservices/list?include_internals=true)
    - It need use tool to convert ([sonarqube-convert-webservice-to-oas.py](examples/swaggers/sonarqube/sonarqube-convert-webservice-to-oas.py)) to OAS format from [sonarqube webservices schema](examples/swaggers/sonarqube/sonarqube-v2026.4.0.124573.webservices.json).
- [SonarQube API (Page) - @Deprecated](https://next.sonarqube.com/sonarqube/web_api) (Many commonly used APIs are missing)
- [SonarQube API (Schema) - @Deprecated](https://next.sonarqube.com/sonarqube/api/v2/api-docs) (Many commonly used APIs are missing)
- [sonarqube-mcp-server - @Deprecated](https://github.com/sonarsource/sonarqube-mcp-server) (official java edition)
- [go-sonarqube-mcp-server - @Deprecated](https://github.com/flowgent-labs/go-sonarqube-mcp-server) (enhanced go edition based on official above)

## License

[MIT License](LICENSE)
