# Go MCP server Generator from OpenAPI Specification

Generate production-ready Model Context Protocol (MCP) servers from OpenAPI specs. Each API operation becomes an AI tool that forwards requests to your upstream service.

## Quick Start

### Building

```sh
make
```

### Generate the Confluence MCP server for examples

```sh
./bin/mcpgen -v -i examples/confluence-mcp/confluence-server-v10.2.14.oas.v3.0.1.json -o examples/confluence-mcp \
  --includes "listSpaces,createPage,updatePage,deletePage"
cd examples/confluence-mcp
```

This produces a complete Go project with tools for every operation:

```plaintext
confluence-mcp/
├── bin
├   └── confluence-mcp           # compiled binary
├── .credentials                 # file-based token (set MCP_UPSTREAM_TOKEN_FILE)
├── main.go                      # entry point (stdio/http/cli transport)
├── mcpclient.sh                 # quick curl-based test script
├── Makefile                     # build / run / clean / test
└── pkg/
    ├── mcpserver/server.go      # MCP server setup + tool registration
    ├── helpers/                 # ForwardRequest, logging, parameter parsing
    └── mcptools/                # one file per API operation
        ├── GetPage.go
        ├── CreatePage.go
        ├── UpdatePage.go
        ├── DeletePage.go
        ├── SearchContent.go
        └── ...
```

### Start in `HTTP` mode

The server defaults to httpbin.org which echoes requests — great for quick verification:

```sh
examples/confluence-mcp --transport http --port 8080 -v 1
# MCP_UPSTREAM_ENDPOINT=https://httpbin.org/anything
```

Set your actual upstream to enable real API calls:

```sh
export MCP_UPSTREAM_ENDPOINT=https://api.example.com
# Optional 1: pass token via env var
MCP_UPSTREAM_TOKEN=your-token /path/to/confluence-mcp --transport http --port 8080 -v 1

# Optional 2: read token from file (safer, no shell history exposure)
echo -n "your-token" > .credentials
MCP_UPSTREAM_TOKEN_FILE=.credentials /path/to/confluence-mcp --transport http --port 8080 -v 1
```

### Test with mcpclient.sh for `HTTP` transport only.

```sh
./mcpclient.sh list-tools
./mcpclient.sh call GetPage '{"id": "123456"}'
```

## Populars application Swagger

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

## Generator Configuration

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
| `MCP_UPSTREAM_ENDPOINT` | Base URL of the upstream API (default: `https://httpbin.org/anything`) |
| `MCP_UPSTREAM_TOKEN` | Bearer token for upstream auth (fallback when no Authorization header from client) |
| `MCP_UPSTREAM_TOKEN_FILE` | Path to a file containing the bearer token (alternative to `MCP_UPSTREAM_TOKEN`) |

### Token retrieval priority

The server tries to obtain a Bearer token in this order:

1. Authorization header from the client's HTTP request (forwarded)
2. `MCP_UPSTREAM_TOKEN` environment variable
3. `MCP_UPSTREAM_TOKEN_FILE` (read from file — ideal for Kubernetes secrets)
4. macOS Keychain (`security find-generic-password -s mcpgen-upstream -wa ""`)
5. Windows Credential Manager (`cmdkey /get:mcpgen-upstream`)

### Token format

The token value is inspected for a recognized prefix. If the value already starts with `Bearer ` or `Basic ` (case-insensitive), it is used as-is in the `Authorization` header. Otherwise, `Bearer ` is automatically prepended.

### Tool filtering

For specs with many operations, limit which tools AI agents can discover via an optional config file:

```sh
# Print the default config template
/path/to/confluence-mcp --print-default-config

# Edit ~/.confluence-mcp/config.yaml and list only the tools you want
```

`$HOME/.{binaryName}/config.yaml`:

```yaml
tools:
  include:
    - ListSpaces
    - SearchContent
```

When `tools.include` is non-empty, only those tools are registered with the MCP server and shown in `-t cli list`. When absent or empty, all tools are available.

### Virtual Tools (Composition)

Compose multiple native tools into a single AI-callable tool via a declarative 5-step pipeline DSL. Add a `virtualTools` list to your config file:

`$HOME/.{binaryName}/config.yaml`:

```yaml
virtualTools:
  - name: MyVirtualTool
    description: Retrieve application details with remediation suggestions
    inputSchema:
      type: object
      properties:
        appId:
          type: string
      required: [appId]
    pipeline:
      - id: getApp
        kind: call
        spec:
          tool: GetApplication
          parse: json
          args:
            applicationId: $input.appId
      - id: appName
        kind: jq
        spec:
          from: $getApp
          expr: .name
      - id: done
        kind: return
        spec:
          from: $appName
```

Pipeline step kinds: `call` (invoke an MCP tool), `jq` (jq expression transform), `foreach` (concurrent iteration over arrays), `emit` (output within foreach), and `return` (final result). Full documentation in [.agents/skills/virtual-tool-creator/](.agents/skills/virtual-tool-creator/).


## Generated MCP Server - Agent Integration

### Local Mode (stdio)

Run the MCP server as a child process — recommended for local development.

### OpenCode

`~/.config/opencode/config.json`:

```json
{
  "mcp": {
    "confluence-mcp": {
      "type": "local",
      "command": ["bash", "-c", "/path/to/confluence-mcp"],
      "args": ["--transport", "stdio"],
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://api.example.com",
        "MCP_UPSTREAM_TOKEN": "your-token",
        "MCP_UPSTREAM_TOKEN_FILE": "/path/to/fallback/.credentials"
      },
      "enabled": true
    }
  }
}
```

### Claude Code

`~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "confluence-mcp": {
      "command": "/path/to/confluence-mcp",
      "args": ["--transport", "stdio"],
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://api.example.com",
        "MCP_UPSTREAM_TOKEN": "your-token",
        "MCP_UPSTREAM_TOKEN_FILE": "/path/to/fallback/.credentials"
      }
    }
  }
}
```

### Claude Desktop

`~/.config/claude-desktop/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "confluence-mcp": {
      "command": ["bash", "-c", "/path/to/confluence-mcp"],
      "args": ["--transport", "stdio"],
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://api.example.com",
        "MCP_UPSTREAM_TOKEN": "your-token",
        "MCP_UPSTREAM_TOKEN_FILE": "/path/to/fallback/.credentials"
      }
    }
  }
}
```

### Codex CLI

`~/.codex/config.yaml`:

```yaml
mcp:
  servers:
    confluence-mcp:
      command: /path/to/confluence-mcp
      args: ["--transport", "stdio"]
      env:
        MCP_UPSTREAM_ENDPOINT: https://api.example.com
        MCP_UPSTREAM_TOKEN: your-token
        MCP_UPSTREAM_TOKEN_FILE: "/path/to/fallback/.credentials"
```

### Cursor

`~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "confluence-mcp": {
      "command": "/path/to/confluence-mcp",
      "args": ["--transport", "stdio"],
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://api.example.com",
        "MCP_UPSTREAM_TOKEN": "your-token",
        "MCP_UPSTREAM_TOKEN_FILE": "/path/to/fallback/.credentials"
      }
    }
  }
}
```

### Remote Mode (HTTP)

Run the server separately and connect agents via HTTP transport. Suitable for shared instances, cloud deployments, or when agent cannot spawn local processes.

Start the server:

```sh
export MCP_UPSTREAM_ENDPOINT=https://api.example.com
export MCP_UPSTREAM_TOKEN=your-token
/path/to/confluence-mcp --transport http --port 8080 -v 1
```

### OpenCode (remote)

`~/.config/opencode/config.json`:

```json
{
  "mcp": {
    "confluence-mcp": {
      "type": "remote",
      "url": "http://localhost:8080/mcp",
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://api.example.com",
        "MCP_UPSTREAM_TOKEN": "your-token",
        "MCP_UPSTREAM_TOKEN_FILE": "/path/to/fallback/.credentials"
      }
    }
  }
}
```

### Claude Code (remote)

`~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "confluence-mcp": {
      "url": "http://localhost:8080/mcp",
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://api.example.com",
        "MCP_UPSTREAM_TOKEN": "your-token",
        "MCP_UPSTREAM_TOKEN_FILE": "/path/to/fallback/.credentials"
      }
    }
  }
}
```

### Claude Desktop (remote)

`~/.config/claude-desktop/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "confluence-mcp": {
      "url": "http://localhost:8080/mcp",
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://api.example.com",
        "MCP_UPSTREAM_TOKEN": "your-token",
        "MCP_UPSTREAM_TOKEN_FILE": "/path/to/fallback/.credentials"
      }
    }
  }
}
```

### Codex CLI (remote)

`~/.codex/config.yaml`:

```yaml
mcp:
  servers:
    confluence-mcp:
      url: http://localhost:8080/mcp
      env:
        MCP_UPSTREAM_ENDPOINT: https://api.example.com
        MCP_UPSTREAM_TOKEN: your-token
        MCP_UPSTREAM_TOKEN_FILE: "/path/to/fallback/.credentials"
```

### Cursor (remote)

`~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "confluence-mcp": {
      "url": "http://localhost:8080/mcp",
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://api.example.com",
        "MCP_UPSTREAM_TOKEN": "your-token",
        "MCP_UPSTREAM_TOKEN_FILE": "/path/to/fallback/.credentials"
      }
    }
  }
}
```

### Usage for CLI Mode (example)

Invoke tools directly from the command line — no MCP agent needed. Useful for debugging, scripting, and manual API exploration. The CLI reuses the same `mcptools` handlers as the MCP server, so every call makes a real HTTP request upstream.

```sh
# Set your upstream endpoint (required for real API calls)
export MCP_UPSTREAM_ENDPOINT=https://api.example.com
export MCP_UPSTREAM_TOKEN=your-token

# First call: list available tools
/path/to/confluence-mcp -t cli list

# First tool call: fetch a page by ID
/path/to/confluence-mcp -t cli Getpage --id 123456

# Show tool-specific help (GNU-style usage)
/path/to/confluence-mcp -t cli Getpage --help

# Call a tool with GNU-style --flag arguments
/path/to/confluence-mcp -t cli ListSpaces --limit=5 --type global
/path/to/confluence-mcp -t cli SearchContent --cql 'type=page AND text~"API"' --limit 10

# Call a tool without arguments (for tools that have no required params)
/path/to/confluence-mcp -t cli ListSpaces
```

## License

[MIT License](LICENSE)
