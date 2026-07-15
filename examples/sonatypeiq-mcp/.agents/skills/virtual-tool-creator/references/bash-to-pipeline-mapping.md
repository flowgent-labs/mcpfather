# Bash/jq → Virtual Tool Pipeline DSL Mapping

Translate common bash API orchestration patterns into the virtual tool pipeline DSL.

## Single API Call

**Bash**:
```bash
curl "$BASE/api/v2/apps/$ID" -o result.json
```

**Pipeline**:
```yaml
- id: getApp
  kind: call
  spec:
    tool: GetApplication
    args:
      applicationId: $input.appId
```

If the upstream returns JSON and downstream steps need parsed data, add `parse: json`:

```yaml
- id: getApp
  kind: call
  spec:
    tool: GetApplication
    parse: json
    args:
      applicationId: $input.appId
```

Without `parse: json`, the response is kept as a raw text string.

## Chained Calls (B depends on A)

**Bash**:
```bash
APP="$(curl "$BASE/api/v2/apps/$PUBLIC_ID" | jq -r '.id')"
curl "$BASE/api/v2/apps/$APP/details"
```

**Pipeline**:
```yaml
- id: getApp
  kind: call
  spec:
    tool: GetApplication
    parse: json
    args:
      applicationId: $input.publicAppId

- id: getDetails
  kind: call
  spec:
    tool: GetApplicationDetails
    args:
      internalAppId: $getApp.id
```

## Iterate a List + Call an API per Element

**Bash**:
```bash
jq -c '.[]' items.json | while read item; do
  ID="$(echo "$item" | jq -r '.id')"
  curl "$BASE/api/v2/items/$ID/details" >> results.jsonl
done
```

**Pipeline**:
```yaml
- id: enrich
  kind: foreach
  spec:
    in: $input.items
    as: item
    concurrency: 4
    preserveOrder: true
    pipeline:
      - id: getDetail
        kind: call
        spec:
          tool: GetItemDetail
          args:
            id: $item.id
      - id: emitResult
        kind: emit
        spec:
          from: $getDetail
```

## Fetch A → Foreach over A's Items → Call B per Item → Merge Results

**Bash**:
```bash
curl "$BASE/api/v2/data" | jq -c '.items[]' | while read item; do
  ID="$(echo "$item" | jq -r '.id')"
  DETAIL="$(curl "$BASE/api/v2/items/$ID/detail")"
  echo "$item" | jq --argjson detail "$DETAIL" '. + {detail: $detail}'
done
```

**Pipeline**:
```yaml
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

## Field Filtering (select)

**Bash**:
```bash
jq '[.[] | select(.policyThreatLevel >= 5)]' data.json
```

**Pipeline** (jq step):
```yaml
- id: filterThreat
  kind: jq
  spec:
    from: $data
    vars:
      min: $input.minThreatLevel
    expr: '[.[] | select(.policyThreatLevel >= $min)]'
```

## Field Projection (keep only selected fields)

**Bash**:
```bash
jq '{name, email}' data.json
# or for arrays:
jq '[.[] | {name, email}]' data.json
```

**Pipeline**:
```yaml
- id: project
  kind: jq
  spec:
    from: $data
    expr: '{name, email}'
```

## Field Renaming

**Bash**:
```bash
jq '{display_name: .displayName, url: .packageUrl}' data.json
```

**Pipeline**:
```yaml
- id: rename
  kind: jq
  spec:
    from: $data
    expr: '{display_name: .displayName, url: .packageUrl}'
```

## Delete Fields

**Bash**:
```bash
jq 'del(.internal, ._links)' data.json
```

**Pipeline**:
```yaml
- id: cleanup
  kind: jq
  spec:
    from: $data
    expr: 'del(.internal, ._links)'
```

## Pick First Array Element or Default

**Bash**:
```bash
jq '.reports[0].stage' history.json
# with fallback:
jq '(.reports[0].stage) // null' history.json
```

**Pipeline**:
```yaml
- id: firstStage
  kind: jq
  spec:
    from: $history
    vars:
      scanId: $input.scanId
    expr: '([.reports[] | select(.scanId == $scanId) | .stage][0] // null)'
```

## Construct a Top-Level Result Object

**Bash**:
```bash
jq -n '{app: $app, items: $items, count: ($items | length)}'
```

**Pipeline** (return step + jq expr):
```yaml
- id: buildSummary
  kind: return
  spec:
    from: $items
    vars:
      app: $app
    expr: |
      {
        app: $app,
        items: .,
        count: length
      }
```

## Compute Aggregate Statistics

**Bash**:
```bash
jq 'length' data.json
jq '[.[] | select(.waived != true)] | length' data.json
```

**Pipeline** (jq step):
```yaml
- id: computeStats
  kind: jq
  spec:
    from: $components
    expr: |
      {
        totalCount: length,
        activeCount: ([.[] | select(.waived != true)] | length)
      }
```

## Conditional Branching (if/then/else)

**Bash**:
```bash
if [ "$(echo "$data" | jq '.count > 0')" = "true" ]; then
  # process
else
  echo '{"components":[],"totalCount":0}'
fi
```

**Pipeline** (handled inside a jq expression):
```yaml
- id: handle
  kind: jq
  spec:
    from: $data
    expr: |
      if .count > 0 then
        {components: .components, totalCount: .count}
      else
        {components: [], totalCount: 0}
      end
```

## Post-Step Validation (require non-empty)

**Bash**:
```bash
result="$(jq '.app.id' response.json)"
if [ -z "$result" ]; then
  echo "ERROR: missing app id" >&2
  exit 1
fi
```

**Pipeline**:
```yaml
- id: appId
  kind: jq
  spec:
    from: $policy
    expr: '.application.id'
  require:
    nonEmpty: true
    message: "Cannot find application id in policy response."
```

## Operations Without a Direct DSL Equivalent

| Bash/jq operation | Reason | Suggestion |
|---|---|---|
| Streaming large file download | No streaming in pipeline | Handled by the `call` step; the native tool performs the download |
| String concatenation / formatting | No dedicated string step | Use `+` inside a jq expression |
| Date arithmetic | No date functions | Compute inside a jq expression with `now` and `strftime`, or push logic into the native tool |
| External process / sub-shell | DSL is declarative only | Embed logic in the native MCP tool implementation |

## Reference Syntax Quick Reference

| Context | Syntax | Examples |
|---|---|---|
| `spec.args` values | `$root.path` | `$input.userId`, `$getApp.id`, `$item.name` |
| `spec.from` | `$stepId` or `$varName` | `$history`, `$component` |
| `spec.in` (foreach) | `$ref` | `$threatComponents` |
| `spec.vars` keys | any string | `vars: { scanId: $input.scanId }` |
| `spec.vars` values | `$ref` | `minLevel: $input.minThreatLevel` |

The first path segment after `$` resolves to:
- `input` — the virtual tool's input arguments
- A step `id` — that step's output
- A foreach `as` name — the current iteration element
