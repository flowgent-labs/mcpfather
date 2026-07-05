# SonarQube Virtual Tools — End-to-End Test Report

> Date: 2026-07-04
>
> Goal: End-to-end validation of `get_overall_issues` and `get_newcode_issues` virtual tools against a real SonarQube instance.

## 0. Prerequisites

If you already have a `.env` file, **skip the copy step** — just verify each key below is present, then source it.

If you don't have one yet, create it from the template:

```bash
# Only if .env does NOT exist yet:
cp .env.example .env
# Then edit .env with your real SonarQube URL, token, and project key
```

Before proceeding, check your `.env` has every required key:

```bash
# Verify all keys are set (should print value for each, no empty output)
source .env
echo "MCP_UPSTREAM_ENDPOINT  = $MCP_UPSTREAM_ENDPOINT"
echo "MCP_UPSTREAM_TOKEN     = ${MCP_UPSTREAM_TOKEN:0:10}..."
echo "SONARQUBE_PROJECT_KEY  = $SONARQUBE_PROJECT_KEY"
echo "SONARQUBE_TEST_BRANCH  = $SONARQUBE_TEST_BRANCH"
echo "SONARQUBE_TEST_COMPONENT = $SONARQUBE_TEST_COMPONENT"
echo "SONARQUBE_TEST_PR       = $SONARQUBE_TEST_PR"
echo "HTTPS_PROXY            = $HTTPS_PROXY"
echo "HTTP_PROXY             = $HTTP_PROXY"
```

Key environment variables:

| Variable | Description |
|----------|-------------|
| `MCP_UPSTREAM_ENDPOINT` | Base URL of your SonarQube instance |
| `MCP_UPSTREAM_TOKEN` | SonarQube API token |
| `SONARQUBE_PROJECT_KEY` | Project key under test |
| `SONARQUBE_TEST_BRANCH` | Git branch under test |
| `SONARQUBE_TEST_COMPONENT` | Optional file path filter (component key) |
| `SONARQUBE_TEST_PR` | Optional pull request ID |
| `HTTPS_PROXY` / `HTTP_PROXY` | Optional, for restricted network environments |

## 1. Test Environment

| Item | Value |
|------|-------|
| mcpgen version | Built from source (`make build`) |
| SonarQube target | `$MCP_UPSTREAM_ENDPOINT` (from `.env`) |
| Test project | `$SONARQUBE_PROJECT_KEY` (from `.env`) |
| Test branch | `$SONARQUBE_TEST_BRANCH` (from `.env`) |
| Test file | `$SONARQUBE_TEST_COMPONENT` (from `.env`) |
| OpenAPI spec | `examples/swaggers/sonarqube-v2026.4.0.124573.oas.3.1.0.json` (385 paths) |
| Native tools | 385 generated MCP tools |

## 2. Setup Steps

### 2.1 Build mcpgen

```bash
make build
# → bin/mcpgen-linux-amd64-<version>
```

### 2.2 Generate the sonarqube-mcp project

```bash
bin/mcpgen \
  -i examples/swaggers/sonarqube-v2026.4.0.124573.oas.3.1.0.json \
  -o examples/sonarqube-mcp
```

If behind a restricted network, prefix with proxy env vars:

```bash
HTTPS_PROXY="$HTTPS_PROXY" HTTP_PROXY="$HTTP_PROXY" \
  bin/mcpgen \
    -i examples/swaggers/sonarqube-v2026.4.0.124573.oas.3.1.0.json \
    -o examples/sonarqube-mcp
```

Output: 385 native tools, including the required upstream tools:
- `GetIssuesSearch` (`api/issues/search`)
- `GetSourcesIssueSnippets` (`api/sources/issue_snippets`)
- `GetDuplicationsShow` (`api/duplications/show`)
- `GetSourcesLines` (`api/sources/lines`)
- `GetMeasuresComponent` (`api/measures/component`)

### 2.3 Build the sonarqube-mcp binary

```bash
make -C examples/sonarqube-mcp
# → bin/sonarqube-mcp
```

(Add `HTTPS_PROXY` / `HTTP_PROXY` prefix if needed.)

### 2.4 Deploy configuration

```bash
mkdir -p ~/.sonarqube-mcp
cp .agents/skills/virtual-tool-creator/resources/sonarqube-example-config.yaml \
   ~/.sonarqube-mcp/config.yaml
```

### 2.5 Source credentials

```bash
source .env
```

> If your `.env` keys are not prefixed with `export`, use `set -a && source .env && set +a` instead so the variables are available to child processes.

> The generated tools read `MCP_UPSTREAM_ENDPOINT` and `MCP_UPSTREAM_TOKEN`. The example config also uses `tools.expose` (not the legacy `tools.activates` key).

## 3. Virtual Tool API Reference

### 3.1 `get_overall_issues`

| Parameter | Type | Required | Default | Description |
|-----------|------|:--------:|---------|-------------|
| `projectKey` | string | yes | - | SonarQube project key |
| `branch` | string | **yes** | - | Git branch name (e.g., `main`) |
| `component` | string | no | `""` | Optional file path filter (SonarQube component key) |
| `limit` | integer | no | 50 | Max results (1–500) |
| `newCodeOnly` | boolean | no | false | Return only new-code-period issues |
| `snippetConcurrency` | integer | no | 6 | Concurrent snippet fetches (1–16) |

**Pipeline**: `GetIssuesSearch` → foreach → `GetSourcesIssueSnippets` → return minimal fields

### 3.2 `get_newcode_issues`

| Parameter | Type | Required | Default | Description |
|-----------|------|:--------:|---------|-------------|
| `projectKey` | string | yes | - | SonarQube project key |
| `branch` | string | **yes** | - | Git branch name |
| `pullRequest` | string | **yes** | - | Pull request ID |
| `component` | string | no | `""` | Optional file path filter |
| `limit` | integer | no | 50 | Max results (1–500) |
| `snippetConcurrency` | integer | no | 6 | Concurrent snippet fetches (1–16) |

**Pipeline**: Same as above, with `pullRequest` forwarded to SonarQube.

## 4. Test Results

### 4.1 `get_overall_issues` — by branch (no component filter)

```bash
source .env && examples/sonarqube-mcp/bin/sonarqube-mcp -v 10 -t cli get_overall_issues \
  --projectKey="$SONARQUBE_PROJECT_KEY" \
  --branch="$SONARQUBE_TEST_BRANCH" \
  --limit="2"
```

**Result**: pass

```json
{
  "project": "<projectKey>",
  "branch": "<branch>",
  "summary": {"total": 1406, "returned": 2},
  "issues": [
    {
      "key": "c27f6c23-...",
      "rule": "docker:S7020",
      "severity": "MINOR",
      "type": "CODE_SMELL",
      "message": "Line is too long...",
      "component": "<projectKey>:path/to/file",
      "startLine": 304, "endLine": 304,
      "debt": "2min",
      "codeSnippet": [{"lineNumber": 299, "code": "&& echo..."}, ...]
    }
  ]
}
```

### 4.2 `get_overall_issues` — by branch + file filter

```bash
source .env && examples/sonarqube-mcp/bin/sonarqube-mcp -v 10 -t cli get_overall_issues \
  --projectKey="$SONARQUBE_PROJECT_KEY" \
  --branch="$SONARQUBE_TEST_BRANCH" \
  --component="$SONARQUBE_TEST_COMPONENT" \
  --limit="5"
```

**Result**: pass — total=1, precisely filtered to a single file

```json
{
  "summary": {"total": 1, "returned": 1},
  "issues": [
    {
      "key": "94d52855-...",
      "rule": "java:S1192",
      "severity": "CRITICAL",
      "type": "CODE_SMELL",
      "message": "Define a constant instead of duplicating this literal ...",
      "component": "<projectKey>:path/to/ExampleFile.java",
      "startLine": 66, "endLine": 66,
      "debt": "8min",
      "codeSnippet": [
        {"lineNumber": 61, "code": "..."},
        {"lineNumber": 66, "code": "...duplicated literal..."},
        {"lineNumber": 90, "code": "...duplicated literal..."}
      ]
    }
  ]
}
```

### 4.3 `get_overall_issues` — newCodeOnly

```bash
source .env && examples/sonarqube-mcp/bin/sonarqube-mcp -v 10 -t cli get_overall_issues \
  --projectKey="$SONARQUBE_PROJECT_KEY" \
  --branch="$SONARQUBE_TEST_BRANCH" \
  --newCodeOnly="true" --limit="3"
```

**Result**: pass — returns zero if no new-code-period issues exist

```json
{"summary": {"total": 0, "returned": 0}}
```

### 4.4 `get_newcode_issues` — non-existent PR (error path)

```bash
source .env && examples/sonarqube-mcp/bin/sonarqube-mcp -v 10 -t cli get_newcode_issues \
  --projectKey="$SONARQUBE_PROJECT_KEY" \
  --branch="$SONARQUBE_TEST_BRANCH" \
  --pullRequest="$SONARQUBE_TEST_PR" --limit="3"
```

**Result**: SonarQube returns HTTP 500 for a non-existent PR — error is propagated correctly. An actual PR must exist for the happy path to succeed.

### 4.5 Native tool spot-checks (optional)

Look up the component file:

```bash
source .env && examples/sonarqube-mcp/bin/sonarqube-mcp -v 10 -t cli GetComponentsShow \
  --component="$SONARQUBE_TEST_COMPONENT"
```

Fetch raw issues for the same file:

```bash
source .env && examples/sonarqube-mcp/bin/sonarqube-mcp -v 10 -t cli GetIssuesSearch \
  --components="$SONARQUBE_TEST_COMPONENT" \
  --branch="$SONARQUBE_TEST_BRANCH" \
  --types="CODE_SMELL,BUG,VULNERABILITY"
```

### 4.6 HTTP Mode

The same virtual tools work over the MCP HTTP transport (`StreamableHTTPServer`).

#### 4.6.1 Start the server

In one terminal, start the HTTP server:

```bash
source .env && examples/sonarqube-mcp/bin/sonarqube-mcp -v 10 -t http -p 8080
```

Wait for the log line `MCP server listening on :8080/mcp`.

#### 4.6.2 Test via mcpclient.sh

In another terminal, call the virtual tool through the MCP HTTP endpoint:

```bash
source .env && ./examples/sonarqube-mcp/mcpclient.sh call get_overall_issues \
  '{"projectKey":"'$SONARQUBE_PROJECT_KEY'","branch":"'$SONARQUBE_TEST_BRANCH'","component":"'$SONARQUBE_TEST_COMPONENT'"}'
```

**Result**: pass — same response shape as CLI mode (section 4.2), delivered over HTTP.

```json
{
  "summary": {"total": 1, "returned": 1},
  "issues": [
    {
      "key": "94d52855-...",
      "rule": "java:S1192",
      "severity": "CRITICAL",
      "type": "CODE_SMELL",
      "message": "Define a constant instead of duplicating this literal ...",
      "component": "<projectKey>:path/to/ExampleFile.java",
      "startLine": 66, "endLine": 66,
      "debt": "8min",
      "codeSnippet": [
        {"lineNumber": 61, "code": "..."},
        {"lineNumber": 66, "code": "...duplicated literal..."},
        {"lineNumber": 90, "code": "...duplicated literal..."}
      ]
    }
  ]
}
```

#### 4.6.3 List tools via HTTP

```bash
source .env && ./examples/sonarqube-mcp/mcpclient.sh list-tools
```

#### 4.6.4 Native tool via HTTP

```bash
source .env && ./examples/sonarqube-mcp/mcpclient.sh call GetComponentsShow \
  '{"component":"'$SONARQUBE_TEST_COMPONENT'"}'
```

#### 4.6.5 Obtain MCP server metrics

```bash
curl -v localhost:9991/metrics
```

## 5. Bugs Found and Fixed During Testing

| # | Issue | Root cause | Fix |
|---|-------|------------|-----|
| 1 | `codeSnippet` array always empty `[]` | `GetSourcesIssueSnippets` returns `{componentKey: {sources: [...]}}`; jq path `$snippet.sources` could not reach the nested array | Changed jq path to `$snippet \| .[] \| .sources[]?` |
| 2 | Optional params (`component`, etc.) caused nil errors when omitted | `$input.optionalField` fails immediately when the field is absent; `inputSchema.default` is documentation-only | Added `opts` jq preprocessing step: `(.field // default)` normalizes all optional params |
| 3 | `inNewCodePeriod=true` + `projects` returned HTTP 400 | SonarQube requires `components`/`componentKeys` when `inNewCodePeriod` is set | Unified to `components` only (works for both project keys and file keys) |
| 4 | `componentKeys` + `components` conflict when both present | They are aliases for the same parameter; `components` overrides `componentKeys` | Use only `components`, with jq `opts.scope` dynamically selecting project key or file key |
| 5 | `proxy.golang.org` unreachable | Restricted network environment | Set `HTTPS_PROXY` / `HTTP_PROXY` for `go mod tidy` and `go build` |
| 6 | `branch` parameter documented as "Not available in community edition" | Community Edition lacks branch analysis | Target Server/Enterprise Edition where `branch` is supported |

## 6. Key Design Lessons

### 6.1 jq expression patterns

`GetSourcesIssueSnippets` response structure:

```json
{
  "<projectKey>:path/to/file": {
    "component": {...},
    "sources": [{"line": <N>, "code": "..."}]
  }
}
```

Correct extraction path: `$snippet | .[] | .sources[]?`
- `.[]` — iterate over object values (bypasses the unknown component key)
- `.sources[]?` — extract the sources array; `?` avoids errors on empty/missing values

### 6.2 Optional parameter normalization

**Key lesson**: In pipeline args, `$input.optionalField` fails immediately when the field is absent. `inputSchema.default` is documentation-only and is never enforced at runtime.

**Solution**: Normalize all optional params in a first-step jq transformation:

```yaml
- id: opts
  kind: jq
  spec:
    from: $input
    expr: |
      {
        scope: (if (.component // "") != "" then .component else .projectKey end),
        limit: (.limit // 50),
        newCodeOnly: (.newCodeOnly // false),
        snippetConcurrency: (.snippetConcurrency // 6)
      }
```

All downstream steps reference `$opts.xxx` instead of `$input.xxx` for optional params.

### 6.3 SonarQube API parameter constraints

- `inNewCodePeriod` must be paired with `components`/`componentKeys` — cannot use `projects`
- `components` and `componentKeys` are aliases; passing both causes the latter to be silently dropped
- `GetComponentsSearch` qualifiers only support `TRK`
- `pullRequest` on a non-existent PR returns HTTP 500 (not 404)

### 6.4 Proxy-aware builds

In restricted network environments, prefix all `go` commands with proxy env vars:

```bash
HTTPS_PROXY="$HTTPS_PROXY" HTTP_PROXY="$HTTP_PROXY" make -C examples/sonarqube-mcp
```

## 7. Production Config

Configuration files used:

| Path | Purpose |
|------|---------|
| `~/.sonarqube-mcp/config.yaml` | User deployment config |
| `.agents/skills/virtual-tool-creator/resources/sonarqube-example-config.yaml` | Reference example |

Virtual tool feature matrix:

| Tool | branch | component | pullRequest | newCodeOnly |
|------|:------:|:---------:|:-----------:|:-----------:|
| `get_overall_issues` | required | optional | — | optional |
| `get_newcode_issues` | required | optional | required | — |
