# Go MCP server Generator from OpenAPI Specification

Generate production-ready Model Context Protocol (MCP) servers from OpenAPI specs. Each API operation becomes an AI tool that forwards requests to your upstream service.

## Quick Start

### Building

```sh
make
```

### Generate the MCP server (e.g. Confluence API)

```sh
./bin/mcpgen -v -i testdata/example_confluence_oas_v3.1.yaml -o myconfluence-mcp \
  --includes "listSpaces,createPage,updatePage,deletePage"
cd myconfluence-mcp
```

This produces a complete Go project with tools for every operation:

```
myconfluence-mcp/
├── .credentials                 # file-based token (set MCP_UPSTREAM_TOKEN_FILE)
├── main.go                      # entry point (stdio/http/cli transport)
├── client.sh                    # quick curl-based test script
├── Makefile                     # build / run / clean / test
├── myconfluence-mcp             # compiled binary
└── internal/
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

### Start the server

The server defaults to httpbin.org which echoes requests — great for quick verification:

```sh
./myconfluence-mcp --transport http --port 8080 -v 1
# MCP_UPSTREAM_ENDPOINT=https://httpbin.org/anything
```

Set your actual upstream to enable real API calls:

```sh
export MCP_UPSTREAM_ENDPOINT=https://api.example.com
# Option 1: pass token via env var
MCP_UPSTREAM_TOKEN=your-token ./myconfluence-mcp --transport http --port 8080 -v 1

# Option 2: read token from file (safer, no shell history exposure)
echo -n "your-token" > .credentials
MCP_UPSTREAM_TOKEN_FILE=.credentials ./myconfluence-mcp --transport http --port 8080 -v 1
```

### Test with client.sh for `http` transport

```sh
./client.sh list-tools
./client.sh call GetPage '{"id": "123456"}'
```

## Populars application Swagger

### Atlassian Jira

- https://developer.atlassian.com/cloud/jira/software/rest/intro/#introduction
- https://dac-static.atlassian.com/cloud/jira/software/swagger.v3.json
- https://developer.atlassian.com/cloud/jira/software/on-premise-swagger.json

### Atlassian Confluence

- https://developer.atlassian.com/cloud/confluence/rest/v2/intro/
- https://dac-static.atlassian.com/cloud/confluence/openapi-v2.v3.json

### Sonatype - IQ

- https://help.sonatype.com/en/iq-api-reference.html
- https://sonatype.github.io/sonatype-documentation/api/iq/latest/iq-api.json

### Sonatype - Nexus Repository

- https://help.sonatype.com/en/api-reference.html
- https://sonatype.github.io/sonatype-documentation/api/nexus-repository/latest/nexus-repository-api.json

### Sonarqube (*Not support Swagger*)

- https://next.sonarqube.com/sonarqube/web_api
- https://github.com/sonarsource/sonarqube-mcp-server (official java edition)
- https://github.com/flowgent-labs/go-sonarqube-mcp-server (enhanced go edition based on official above)

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

```sh
# Only generate tools for specific operations
./bin/mcpgen -i spec.yaml -o mymcp --includes "listSpaces,createPage,getSpaceContent"

# Generate all tools except health checks
./bin/mcpgen -i spec.yaml -o mymcp --excludes "healthCheck,status"

# Generate all tools except a few
./bin/mcpgen -i spec.yaml -o mymcp --excludes "uploadAttachment,removeLabel"

# Preview what gets included/excluded
./bin/mcpgen -i spec.yaml -o mymcp --includes "listSpaces" -v
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
./myconfluence-mcp --print-default-config

# Edit ~/.myconfluence-mcp/config.yaml and list only the tools you want
```

`$HOME/.{binaryName}/config.yaml`:

```yaml
tools:
  include:
    - ListSpaces
    - SearchContent
```

When `tools.include` is non-empty, only those tools are registered with the MCP server and shown in `-t cli list`. When absent or empty, all tools are available.


## Generated MCP Server - Agent Integration

### Local Mode (stdio)

Run the MCP server as a child process — recommended for local development.

### OpenCode

`~/.config/opencode/config.json`:

```json
{
  "mcp": {
    "myconfluence": {
      "type": "local",
      "command": ["bash", "-c", "./myconfluence-mcp"],
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
    "myconfluence-r": {
      "command": "./myconfluence-mcp",
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
    "myconfluence-r": {
      "command": ["bash", "-c", "./myconfluence-mcp"],
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
    myconfluence-r:
      command: ./myconfluence-mcp
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
    "myconfluence-r": {
      "command": "./myconfluence-mcp",
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
./myconfluence-mcp --transport http --port 8080 -v 1
```

### OpenCode (remote)

`~/.config/opencode/config.json`:

```json
{
  "mcp": {
    "myconfluence-r": {
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
    "myconfluence-r": {
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
    "myconfluence-r": {
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
    myconfluence-r:
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
    "myconfluence-r": {
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

### CLI Mode examples

Invoke tools directly from the command line — no MCP agent needed. Useful for debugging, scripting, and manual API exploration. The CLI reuses the same `mcptools` handlers as the MCP server, so every call makes a real HTTP request upstream.

```sh
# Set your upstream endpoint (required for real API calls)
export MCP_UPSTREAM_ENDPOINT=https://api.example.com
export MCP_UPSTREAM_TOKEN=your-token

# First call: list available tools
./myconfluence-mcp -t cli list

# First tool call: fetch a page by ID
./myconfluence-mcp -t cli Getpage --id 123456

# Show tool-specific help (GNU-style usage)
./myconfluence-mcp -t cli Getpage --help

# Call a tool with GNU-style --flag arguments
./myconfluence-mcp -t cli ListSpaces --limit=5 --type global
./myconfluence-mcp -t cli SearchContent --cql 'type=page AND text~"API"' --limit 10

# Call a tool without arguments (for tools that have no required params)
./myconfluence-mcp -t cli ListSpaces
```

## License

[MIT License](LICENSE)
