# confluence-mcp

## Quick Start

### Build from source

```sh
make
```

### Usage example

```sh
# Set your upstream endpoint and authentication
export MCP_UPSTREAM_ENDPOINT=https://api.example.com
# Token-based authentication
export MCP_UPSTREAM_TOKEN='your-token'
# Cookie-based authentication (for legacy app compatibility)
#export MCP_UPSTREAM_COOKIE='JSESSIONID=your-session-id'

# Run the HTTP mode
./bin/confluence-mcp --transport http --port 8080 &

# Run the CLI mode
./bin/confluence-mcp -t cli list

```

## Authentication

### Bearer / Basic Token (Authorization header)

The server attaches an Authorization header to upstream requests using this priority:

1. Authorization header from the client request (forwarded)
2. `MCP_UPSTREAM_TOKEN` environment variable
3. `MCP_UPSTREAM_TOKEN_FILE` file (set `MCP_UPSTREAM_TOKEN_FILE=.credentials`)
4. macOS Keychain / Windows Credential Manager

### Cookie / Session (Cookie header)

For session-based auth (e.g. JSESSIONID), set a Cookie header on upstream requests:

- `MCP_UPSTREAM_COOKIE` environment variable
- `MCP_UPSTREAM_COOKIE_FILE` file (read cookie value from file)

Both token and cookie can be set simultaneously — they are independent headers.

To use a credentials file for your token:

```sh
echo -n "your-token" > .credentials
export MCP_UPSTREAM_TOKEN_FILE=.credentials
```

### Tool filtering

For APIs with many operations, limit which tools AI agents discover:

```sh
# Print the default config template
./bin/confluence-mcp --print-default-config

# Create and edit your config
mkdir -p ~/.confluence-mcp
./bin/confluence-mcp --print-default-config > ~/.confluence-mcp/config.yaml
```

Edit `~/.confluence-mcp/config.yaml` and set `tools.include` to the operation IDs you want:

```yaml
tools:
  include:
    - ListSpaces
    - SearchContent
```

When `tools.include` is non-empty, only those tools are registered and shown in `-t cli list`.

### Virtual Tools (Composition)

Compose multiple native tools into a single AI-callable tool via a declarative 5-step pipeline DSL. Add a `virtualTools` list to your config file:

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

Pipeline step kinds: `call`, `jq`, `foreach`, `emit`, `return`. See the virtual-tool-creator skill for details.

## Agent Integration

All env vars from [Authentication](#authentication) above (including `MCP_UPSTREAM_COOKIE` / `MCP_UPSTREAM_COOKIE_FILE`) can be set in the `env` block of any agent config below.

### Local Mode (stdio)

Run the MCP server as a child process — recommended for local development.

### OpenCode

`~/.config/opencode/config.json`:

```json
{
  "mcp": {
    "confluence-mcp": {
      "type": "local",
      "command": ["./confluence-mcp"],
      "args": ["--transport", "stdio"],
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://api.example.com",
        "MCP_UPSTREAM_TOKEN": "your-token"
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
      "command": "./confluence-mcp",
      "args": ["--transport", "stdio"],
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://api.example.com",
        "MCP_UPSTREAM_TOKEN": "your-token"
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
      "command": ["./confluence-mcp"],
      "args": ["--transport", "stdio"],
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://api.example.com",
        "MCP_UPSTREAM_TOKEN": "your-token"
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
      command: ./confluence-mcp
      args: ["--transport", "stdio"]
      env:
        MCP_UPSTREAM_ENDPOINT: https://api.example.com
        MCP_UPSTREAM_TOKEN: your-token
```

### Cursor

`~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "confluence-mcp": {
      "command": "./confluence-mcp",
      "args": ["--transport", "stdio"],
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://api.example.com",
        "MCP_UPSTREAM_TOKEN": "your-token"
      }
    }
  }
}
```

### Remote Mode (HTTP)

Run the server separately and connect agents via HTTP transport.

Start the server:

```sh
export MCP_UPSTREAM_ENDPOINT=https://api.example.com
export MCP_UPSTREAM_TOKEN=your-token
./confluence-mcp --transport http --port 8080
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
        "MCP_UPSTREAM_TOKEN": "your-token"
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
        "MCP_UPSTREAM_TOKEN": "your-token"
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
        "MCP_UPSTREAM_TOKEN": "your-token"
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
        "MCP_UPSTREAM_TOKEN": "your-token"
      }
    }
  }
}
```
