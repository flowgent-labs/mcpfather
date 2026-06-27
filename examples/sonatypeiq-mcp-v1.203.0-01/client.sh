#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# MCP Server Client Script
# Quick test helper for the generated MCP server.
#
# Usage:
#   ./client.sh                  Show this help message
#   ./client.sh help             Show this help message
#   ./client.sh list-tools       List all available tools
#   ./client.sh call <tool> [argsJson] [--file <path>]
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
client.sh — MCP Server Client Script

Usage:
  ./client.sh [command] [arguments]

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
      MCP_SERVER_ENDPOINT=http://localhost:9090/mcp ./client.sh
  - If authentication is required:
      MCP_UPSTREAM_TOKEN=your-token ./client.sh call <tool>
  - Download tools auto-save to $DOWNLOAD_DIR
  - The script auto-initializes a session on first call

Examples:
USAGE
  cat <<'EOEX'
  # Add (POST)
  ./client.sh call Add '{"body": {"firstName": "firstName_value", "lastName": "lastName_value", "password": "password_value", "realm": "realm_value", "username": "username_value", "email": "email_value"}}'
EOEX
  cat <<'EOEX'
  # AddApplication (POST)
  ./client.sh call AddApplication '{"body": {"id": "id_value", "name": "name_value", "organizationId": "organizationId_value", "publicId": "publicId_value", "applicationTags": [], "contactUserName": "contactUserName_value"}}'
EOEX
  cat <<'EOEX'
  # AddArtifactoryConnection (POST)
  ./client.sh call AddArtifactoryConnection '{"internalOwnerId": "internalOwnerId_value", "ownerType": "application", "body": {"ownerType": "application", "password": "password_value", "username": "username_value", "artifactoryConnectionId": "artifactoryConnectionId_value", "baseUrl": "baseUrl_value", "isAnonymous": false, "ownerId": "ownerId_value"}}'
EOEX
  cat <<'EOEX'
  # AddAutoPolicyWaiveExclusion (POST)
  ./client.sh call AddAutoPolicyWaiveExclusion '{"ownerId": "ownerId_value", "ownerType": "application", "body": {"autoPolicyWaiverId": "autoPolicyWaiverId_value", "matchStrategy": "EXACT_COMPONENT", "ownerId": "ownerId_value", "policyViolationId": "policyViolationId_value", "scanId": "scanId_value", "applicationPublicId": "applicationPublicId_value"}}'
EOEX
  cat <<'EOEX'
  # AddAutoPolicyWaiver (POST)
  ./client.sh call AddAutoPolicyWaiver '{"ownerId": "ownerId_value", "ownerType": "application", "body": {"threatLevel": 0, "createTime": "2025-01-01", "creatorName": "creatorName_value", "publicId": "publicId_value", "ownerId": "ownerId_value", "ownerName": "ownerName_value", "pathForward": false, "scopesOperatorAny": false, "autoPolicyWaiverId": "autoPolicyWaiverId_value", "creatorId": "creatorId_value", "ownerType": "ownerType_value", "reachability": false}}'
EOEX
  cat <<'EOEX'
  # AddAutoPolicyWaivers (POST)
  ./client.sh call AddAutoPolicyWaivers '{"ownerId": "ownerId_value", "ownerType": "application", "body": "value"}'
EOEX
  cat <<'EOEX'
  # AddBulkPolicyWaivers (POST)
  ./client.sh call AddBulkPolicyWaivers '{"ownerId": "ownerId_value", "ownerType": "application", "body": {"apiWaiverOptionsDTO": {}, "violationIds": ["violation-id-1","violation-id-2","violation-id-3"]}}'
EOEX
  cat <<'EOEX'
  # AddLabel (POST)
  ./client.sh call AddLabel '{"ownerId": "ownerId_value", "ownerType": "application", "body": {"description": "description_value", "id": "id_value", "label": "label_value", "ownerId": "ownerId_value", "ownerType": "ownerType_value", "color": "color_value"}}'
EOEX
  cat <<'EOEX'
  # AddLicenseOverride (POST)
  ./client.sh call AddLicenseOverride '{"ownerId": "ownerId_value", "ownerType": "application", "where": "where_value", "body": {"componentIdentifier": {}, "id": "id_value", "licenseIds": [], "ownerId": "ownerId_value", "status": "OPEN", "comment": "comment_value"}}'
EOEX
  cat <<'EOEX'
  # AddOrganization (POST)
  ./client.sh call AddOrganization '{"body": {"tags": [], "id": "id_value", "name": "name_value", "parentOrganizationId": "parentOrganizationId_value"}}'
EOEX
  cat <<'EOEX'
  # AddPolicyWaiverByPolicyViolationId (POST)
  ./client.sh call AddPolicyWaiverByPolicyViolationId '{"ownerId": "ownerId_value", "ownerType": "application", "policyViolationId": "policyViolationId_value", "body": {"comment": "False positive - internal tool approved by security team", "expireWhenRemediationAvailable": false, "expiryTime": "2025-01-01", "matcherStrategy": "EXACT_COMPONENT", "waiverReasonId": "waiver-reason-id-123"}}'
EOEX
  cat <<'EOEX'
  # AddPolicyWaiverRequestByPolicyViolationId (POST)
  ./client.sh call AddPolicyWaiverRequestByPolicyViolationId '{"ownerId": "ownerId_value", "ownerType": "application", "policyViolationId": "policyViolationId_value", "body": {"waiverReasonId": "waiverReasonId_value", "comment": "comment_value", "expireWhenRemediationAvailable": false, "expiryTime": "2025-01-01", "matcherStrategy": "DEFAULT", "noteToReviewer": "noteToReviewer_value"}}'
EOEX
  cat <<'EOEX'
  # AddProprietaryComponentNames (POST)
  ./client.sh call AddProprietaryComponentNames '{"format": "format_value", "body": "value"}'
EOEX
  cat <<'EOEX'
  # AddRepositoryManager (POST)
  ./client.sh call AddRepositoryManager '{"body": {"productName": "productName_value", "productVersion": "productVersion_value", "id": "id_value", "instanceId": "instanceId_value", "name": "name_value"}}'
EOEX
  cat <<'EOEX'
  # AddRole (POST)
  ./client.sh call AddRole '{"body": {"description": "description_value", "id": "id_value", "name": "name_value", "permissionCategories": [], "builtIn": false}}'
EOEX
  cat <<'EOEX'
  # AddSourceControl (POST)
  ./client.sh call AddSourceControl '{"internalOwnerId": "internalOwnerId_value", "ownerType": "application", "body": {"nonGoldenPullRequestsEnabled": false, "token": "token_value", "baseBranch": "baseBranch_value", "ownerId": "ownerId_value", "authenticationType": "authenticationType_value", "closePrAfterDays": 0, "remediationPullRequestsEnabled": false, "pullRequestCommentingEnabled": false, "innerSourceAutomatedUpdatesEnabled": false, "provider": "provider_value", "sshEnabled": false, "repositoryUrl": "repositoryUrl_value", "enablePullRequests": false, "githubAppId": "githubAppId_value", "id": "id_value", "enableStatusChecks": false, "statusChecksEnabled": false, "manualPullRequestsEnabled": false, "closePrAfterDaysOpenEnabled": false, "sourceControlEvaluationsEnabled": false, "username": "username_value", "commitStatusEnabled": false, "sourceControlScanTarget": "sourceControlScanTarget_value", "closePrOnFailedChecksEnabled": false}}'
EOEX
  cat <<'EOEX'
  # AddTag (POST)
  ./client.sh call AddTag '{"organizationId": "organizationId_value", "body": {"id": "id_value", "name": "name_value", "organizationId": "organizationId_value", "color": "color_value", "description": "description_value"}}'
EOEX
  cat <<'EOEX'
  # AddUserMappings (POST)
  ./client.sh call AddUserMappings '{"organizationId": "organizationId_value", "body": {"mappings": [], "role": "role_value"}}'
EOEX
  cat <<'EOEX'
  # AddWaiver (POST)
  ./client.sh call AddWaiver '{"containerImageId": "containerImageId_value", "body": {"waiverReasonId": "waiverReasonId_value", "comment": "comment_value", "expiryTime": "2025-01-01"}}'
EOEX
  cat <<'EOEX'
  # AddWaiverToTransitivePolicyViolationsByAppScanComponent (POST)
  ./client.sh call AddWaiverToTransitivePolicyViolationsByAppScanComponent '{"componentIdentifier": {}, "hash": "hash_value", "ownerId": "ownerId_value", "ownerType": "application", "packageUrl": "packageUrl_value", "scanId": "scanId_value", "body": {"matcherStrategy": "EXACT_COMPONENT", "waiverReasonId": "waiver-reason-id-123", "comment": "False positive - internal tool approved by security team", "expireWhenRemediationAvailable": false, "expiryTime": "2025-01-01"}}'
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
