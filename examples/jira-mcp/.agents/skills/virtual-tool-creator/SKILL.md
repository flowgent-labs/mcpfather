---
name: virtual-tool-creator
description: Generate virtual tool pipeline configurations (YAML) that compose multiple native MCP tools into a single AI-callable tool.
---

# Virtual Tool Creator

Generate `virtualTools` YAML configurations conforming to the [dsl-schema.json](resources/dsl-schema.json) specification from natural language descriptions or existing scripts (bash/jq/yq).

## Two Modes

### Mode A: Build from Scratch

The developer **describes a business scenario** in natural language. Iteratively build the configuration from scratch over multiple turns.

### Mode B: Script Translation

The developer provides **existing bash / jq / yq scripts**. Translate their API orchestration logic into a virtual tool pipeline. Consult [bash-to-pipeline-mapping.md](references/bash-to-pipeline-mapping.md) for patterns.

---

## General Workflow (Both Modes)

### Phase 1: Gather Information

Before writing any configuration, confirm:

1. **Native MCP tool names** — virtual tools invoke generated tools via `call` steps. Discover them with:
   - `ls <project>/internal/mcptools/`
   - `./<binary> -t cli list`
   - `grep -r "func.*InputSchema" internal/mcptools/`
2. **API call chain** — invocation order, data dependencies, which outputs feed into subsequent inputs
3. **Key data structures** — field paths in upstream responses, array positions, fields to keep / drop / rename

### Phase 2: Pipeline Design

Five step kinds (full constraints in [dsl-schema.json](resources/dsl-schema.json)):

| Kind | Purpose | When to Use |
|------|---------|------------|
| `call` | Invoke a native MCP tool | Each upstream API call |
| `jq` | jq expression transformation (`from` + `expr` + `vars`) | Reshape fields, filter, project, construct new objects after a `call` |
| `foreach` | Iterate over an array, executing a sub-pipeline concurrently per element | Enrich each element with additional data |
| `emit` | Output a single element inside a `foreach` sub-pipeline | Send the processed element out of the `foreach` |
| `return` | Return the final result (optional jq `expr`) | The top-level pipeline **must** end with this |

### Phase 3: Schema Validation (Required)

```bash
pip install check-jsonschema  # one-time setup

check-jsonschema \
  --schemafile .agents/skills/virtual-tool-creator/resources/dsl-schema.json \
  ~/.<binary-name>/config.yaml
```

Regenerate the schema when Go structs change: `make gen-config-dsl-schema`

### Phase 4: Deliver

Output:
1. Complete `virtualTools` YAML configuration
2. Deployment path: `$HOME/.<binary-name>/config.yaml`
3. Differences / limitations compared to the original requirements

---

## Core Rules

### 1. Reference Syntax

Use `$`-prefixed references to access pipeline data:

| Location | Format | Examples |
|----------|--------|----------|
| `spec.args` string values | `$root.path` | `$input.userId`, `$policy.application.id` |
| `spec.from` | `$stepId` or `$varName` | `$history`, `$component` |
| `spec.in` (foreach) | `$ref` | `$threatComponents` |
| `spec.vars` values | `$ref` | `vars: { scanId: $input.scanId }` |

The first path segment after `$` resolves to:
- `input` — the tool's input arguments
- A step `id` — that step's output
- A foreach `as` name — the current iteration element

### 2. Pipeline Structure Constraints

- Every step `id` must be unique within its pipeline
- The top-level pipeline **must** end with a `return` step
- Every `foreach` sub-pipeline **must** end with an `emit` step
- `foreach` sub-pipelines must not contain `return` (use `emit` instead)
- Nested `foreach` is not allowed

### 3. jq Expressions

`jq`, `return`, and `emit` steps support jq expressions:

- `from` — the jq input (`.`) — a `$ref` to a previous step or variable
- `vars` — jq variables (`$varName`) — map of names to `$ref` values
- `expr` — the jq expression string

### 4. Response Parsing

- Add `parse: json` to a `call` step to unmarshal upstream JSON responses automatically
- Without `parse: json`, the response body is kept as a raw text string
- Parsed JSON enables dot-path access: `$stepId.field.subfield`

---

## Reference Resources

| Resource | Description |
|----------|------------|
| [dsl-schema.json](resources/dsl-schema.json) | **Authoritative schema** — generated from Go structs by `cmd/gen-config-dsl-schema/main.go` |
| [bash-to-pipeline-mapping.md](references/bash-to-pipeline-mapping.md) | Bash/jq → DSL translation reference |
| `make gen-config-dsl-schema` | Regenerate the JSON schema from Go source |
