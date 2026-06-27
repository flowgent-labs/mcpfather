#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# MCP Server Client Script
# Quick test helper for the generated MCP server.
#
# Usage:
#   ./mcpclient.sh                  Show this help message
#   ./mcpclient.sh help             Show this help message
#   ./mcpclient.sh list-tools       List all available tools
#   ./mcpclient.sh call <tool> [argsJson] [--file <path>]
#
# Environment variables:
#   MCP_UPSTREAM_TOKEN    - Bearer token for MCP server auth
#   MCP_SERVER_ENDPOINT        - Server URL (default: http://localhost:8080/mcp)
#   MCP_SERVER_DOWNLOAD_DIR      - Directory for download responses (default: ./downloads)
# ============================================================

SERVER_URL="${MCP_SERVER_ENDPOINT:-http://localhost:8080/mcp}"
SESSION_ID=""
DOWNLOAD_DIR="${MCP_SERVER_DOWNLOAD_DIR:-downloads}"

usage() {
  cat <<'USAGE'
mcpclient.sh — MCP Server Client Script

Usage:
  ./mcpclient.sh [command] [arguments]

Commands:
  (no args)           Show this help message
  help                Show this help message
  list-tools          List all available tools
  call <tool> [argsJson] [--file <path>]  Call a tool

  --file <path>       Use for upload tools to specify a local file to upload

Environment:
  MCP_SERVER_ENDPOINT        Override server URL (default: http://localhost:8080/mcp)
  MCP_UPSTREAM_TOKEN    Bearer token for server auth
  MCP_SERVER_DOWNLOAD_DIR      Directory for file downloads (default: ./downloads)

Tips:
  - Always uses --noproxy '*' to avoid proxy issues with localhost
  - If the server is running on a different port:
      MCP_SERVER_ENDPOINT=http://localhost:9090/mcp ./mcpclient.sh
  - If authentication is required:
      MCP_UPSTREAM_TOKEN=your-token ./mcpclient.sh call <tool>
  - Download tools auto-save to $DOWNLOAD_DIR
  - The script auto-initializes a session on first call

Examples:
USAGE
  cat <<'EOEX'
  # Add (POST)
  ./mcpclient.sh call Add '{"labelName": "labelName_value", "spaceKey": "spaceKey_value"}'
EOEX
  cat <<'EOEX'
  # AddContentWatcher (POST)
  ./mcpclient.sh call AddContentWatcher '{"contentId": "contentId_value", "key": "value", "username": "username_value"}'
EOEX
  cat <<'EOEX'
  # AddLabels (POST)
  ./mcpclient.sh call AddLabels '{"id": "id_value", "body": "value"}'
EOEX
  cat <<'EOEX'
  # AddSpaceWatch (POST)
  ./mcpclient.sh call AddSpaceWatch '{"contentType": [], "key": "value", "spaceKey": "spaceKey_value", "username": "username_value"}'
EOEX
  cat <<'EOEX'
  # Archive (PUT)
  ./mcpclient.sh call Archive '{"spaceKey": "spaceKey_value"}'
EOEX
  cat <<'EOEX'
  # ByOperation (GET)
  ./mcpclient.sh call ByOperation '{"expand": "expand_value", "id": "id_value"}'
EOEX
  cat <<'EOEX'
  # CancelAllQueuedJobs (PUT)
  ./mcpclient.sh call CancelAllQueuedJobs '{}'
EOEX
  cat <<'EOEX'
  # CancelJob (PUT)
  ./mcpclient.sh call CancelJob '{"jobId": 0}'
EOEX
  cat <<'EOEX'
  # ChangePassword (POST)
  ./mcpclient.sh call ChangePassword '{"username": "username_value", "body": {"password": "password"}}'
EOEX
  cat <<'EOEX'
  # ChangePassword1 (POST)
  ./mcpclient.sh call ChangePassword1 '{"body": {"newPassword": "newPassword", "oldPassword": "oldPassword"}}'
EOEX
  cat <<'EOEX'
  # Children (GET)
  ./mcpclient.sh call Children '{"expand": "expand_value", "id": "id_value", "limit": 25, "parentVersion": 0, "start": 0}'
EOEX
  cat <<'EOEX'
  # ChildrenOfType (GET)
  ./mcpclient.sh call ChildrenOfType '{"expand": "expand_value", "id": "id_value", "limit": 25, "parentVersion": 0, "start": 0, "type": "type_value"}'
EOEX
  cat <<'EOEX'
  # CommentsOfContent (GET)
  ./mcpclient.sh call CommentsOfContent '{"depth": "depth_value", "expand": "expand_value", "id": "id_value", "limit": 25, "location": [], "parentVersion": 0, "start": 0}'
EOEX
  cat <<'EOEX'
  # Contents (GET)
  ./mcpclient.sh call Contents '{"depth": "depth_value", "expand": "expand_value", "limit": 25, "spaceKey": "spaceKey_value", "start": 0}'
EOEX
  cat <<'EOEX'
  # ContentsWithType (GET)
  ./mcpclient.sh call ContentsWithType '{"cursor": "cursor_value", "expand": "expand_value", "limit": 25, "spaceKey": "spaceKey_value"}'
EOEX
  cat <<'EOEX'
  # ContentsWithType1 (GET)
  ./mcpclient.sh call ContentsWithType1 '{"depth": "depth_value", "expand": "expand_value", "limit": 25, "spaceKey": "spaceKey_value", "start": 0, "type": "type_value"}'
EOEX
  cat <<'EOEX'
  # Convert (POST)
  ./mcpclient.sh call Convert '{"expand": "expand_value", "to": "to_value", "body": "value"}'
EOEX
  cat <<'EOEX'
  # Create (POST)
  ./mcpclient.sh call Create '{"body": "value"}'
EOEX
  cat <<'EOEX'
  # Create1 (POST)
  ./mcpclient.sh call Create1 '{"id": "id_value", "body": "value"}'
EOEX
  cat <<'EOEX'
  # Create2 (POST)
  ./mcpclient.sh call Create2 '{"id": "id_value", "key": "key_value", "body": "value"}'
EOEX
}

# --- Auth helper ---

get_auth_header() {
  if [ -n "${MCP_UPSTREAM_TOKEN:-}" ]; then
    printf '%s' "Authorization: Bearer ${MCP_UPSTREAM_TOKEN}"
  fi
}

# --- Session helpers ---

init_session() {
  echo "[*] Initializing MCP session at $SERVER_URL ..." >&2
  local headers_file
  headers_file=$(mktemp)

  local curl_args=(
    -s -D "$headers_file"
    --noproxy '*'
    -X POST "$SERVER_URL"
    -H "Content-Type: application/json"
    -d '{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"client","version":"1.0"}}}'
  )

  local auth_header
  auth_header=$(get_auth_header)
  if [ -n "$auth_header" ]; then
    curl_args+=(-H "$auth_header")
  fi

  local body
  body=$(curl "${curl_args[@]}")

  SESSION_ID=$(grep -oi "Mcp-Session-Id: [^ ]*" "$headers_file" | head -1 | awk '{print $2}' | tr -d '"\r' || true)
  rm -f "$headers_file"
  if [ -z "$SESSION_ID" ]; then
    echo "[!] Failed to get session ID. Is the server running?" >&2
    echo "[!] Response: $body" >&2
    return 1
  fi
  echo "[+] Session: $SESSION_ID" >&2
}

# --- MCP JSON-RPC helpers ---

mcp_request() {
  local method="$1"
  local id="${2:-1}"
  local params
  if [ $# -ge 3 ]; then params="$3"; else params='{ }'; fi

  if [ -z "$SESSION_ID" ]; then
    init_session
  fi

  local curl_args=(
    -s --noproxy '*'
    -X POST "$SERVER_URL"
    -H "Content-Type: application/json"
    -H "Mcp-Session-Id: $SESSION_ID"
    -d "{\"jsonrpc\":\"2.0\",\"id\":$id,\"method\":\"$method\",\"params\":$params}"
  )

  local auth_header
  auth_header=$(get_auth_header)
  if [ -n "$auth_header" ]; then
    curl_args+=(-H "$auth_header")
  fi

  curl "${curl_args[@]}"
}

# --- Tool helpers ---

list_tools() {
  echo "[*] Listing tools ..." >&2
  local result
  result=$(mcp_request tools/list 1)
  echo "$result" | python3 -m json.tool 2>/dev/null || echo "$result"
}

call_tool() {
  local tool_name="${1:?Usage: call_tool <tool-name> [json-args] [--file <path>]}"
  shift
  local args='{}'
  local file_path=""

  # Parse --file flag
  while [ $# -gt 0 ]; do
    case "$1" in
      --file)
        file_path="${2:?--file requires a path argument}"
        shift 2
        ;;
      *)
        args="$1"
        shift
        ;;
    esac
  done

  echo "[*] Calling tool: $tool_name" >&2
  echo "[*] Args: $args" >&2

  # If --file is provided, add local_file_path to the args
  if [ -n "$file_path" ]; then
    if [ ! -f "$file_path" ]; then
      echo "[!] File not found: $file_path" >&2
      return 1
    fi
    local file_size
    file_size=$(wc -c < "$file_path" | tr -d ' ')
    echo "[*] Uploading file: $file_path ($file_size bytes)" >&2

    # Add local_file_path to args JSON
    if command -v python3 >/dev/null 2>&1; then
      args=$(python3 -c "
import json, sys
args = json.loads('$args')
args['local_file_path'] = '$file_path'
print(json.dumps(args))
" 2>/dev/null || echo "$args")
    else
      # Simple jq-based approach or fallback
      args=$(echo "$args" | sed 's/}$/}/' | sed "s/}\"$/,\"local_file_path\":\"$file_path\"}/" | sed "s/{}/{\"local_file_path\":\"$file_path\"}/")
    fi
  fi

  local result
  result=$(mcp_request tools/call 1 "{\"name\":\"$tool_name\",\"arguments\":$args}")

  # Check for error
  if echo "$result" | grep -q '"isError":true'; then
    echo "[!] Tool returned an error:" >&2
    echo "$result" | python3 -m json.tool 2>/dev/null || echo "$result"
    return 1
  fi

  # Check if result indicates a saved file (download tools return "Saved to: <path>")
  if echo "$result" | grep -q '"Saved to:'; then
    local saved_path
    saved_path=$(echo "$result" | grep -o 'Saved to: [^"]*' | sed 's/Saved to: //')
    if [ -n "$saved_path" ] && [ -f "$saved_path" ]; then
      local fsize
      fsize=$(wc -c < "$saved_path" | tr -d ' ')
      echo "[+] Downloaded: $saved_path ($fsize bytes)"
      echo "$result" | python3 -m json.tool 2>/dev/null || echo "$result"
      return 0
    fi
  fi

  # Pretty print JSON response
  echo "$result" | python3 -m json.tool 2>/dev/null || echo "$result"
}

# --- Main ---

case "${1:-help}" in
  help|--help|-h)
    usage
    ;;
  list-tools|list)
    init_session
    list_tools
    ;;
  call)
    shift
    init_session
    call_tool "$@"
    ;;
  *)
    init_session
    call_tool "$@"
    ;;
esac
