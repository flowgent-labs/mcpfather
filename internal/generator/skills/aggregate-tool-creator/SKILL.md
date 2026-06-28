---
name: aggregate-tool-creator
description: 根据开发者描述的业务场景，自动生成符合 mcpgen aggregated tools 规范的聚合工具配置 (YAML)。
---

# Aggregate Tool Creator

根据开发者的自然语言描述或现有脚本（bash/jq/yq），生成符合 [dsl-schema.json](resources/dsl-schema.json) 规范的 `aggregateTools` YAML 配置。

## 两种工作模式

### 模式 A：从零新建

开发者**描述业务场景**（自然语言），从空白开始逐步构建配置。支持多轮对话迭代。

### 模式 B：脚本翻译

开发者提供**现有的 bash / jq / yq 脚本**，将其 API 编排逻辑翻译为聚合工具配置。参考 [bash-to-pipeline-mapping.md](references/bash-to-pipeline-mapping.md)。

---

## 通用工作流（两种模式共用）

### Phase 1: 信息收集

在编写配置之前，必须确认：

1. **原生 MCP Tool 名称** — 聚合工具通过 `call` 步骤调用已生成的工具。获取方式：
   - `ls <project>/internal/mcptools/`
   - `./<binary> -t cli list`
   - `grep -r "func.*InputSchema" internal/mcptools/`
2. **API 调用链路** — 调用顺序、数据依赖关系、哪些输出作为后续输入
3. **关键数据结构** — 上游响应中的字段路径、数组位置、需保留/删除/重命名的字段

### Phase 2: 流水线设计

5 种步骤类型（完整约束见 [dsl-schema.json](resources/dsl-schema.json)）：

| 类型 | 用途 | 何时使用 |
|------|------|---------|
| `call` | 调用原生 MCP tool | 每个上游 API 调用 |
| `jq` | jq 表达式数据变换 (from + expr + vars) | call 之后整理字段、过滤、投影、构造新对象 |
| `foreach` | 遍历数组，对每个元素并发执行子流水线 | 对列表逐元素补充数据 |
| `emit` | foreach 子流水线中输出单个元素 | foreach 内部将处理后的元素送出 |
| `return` | 返回最终结果（支持可选的 jq expr） | 顶层流水线**必须以此结束** |

### Phase 3: Schema 校验（必须）

```bash
pip install check-jsonschema  # 仅需一次

check-jsonschema \
  --schemafile .agents/skills/aggregate-tool-creator/resources/dsl-schema.json \
  ~/.<binary-name>/config.yaml
```

Schema 过时时运行：`make gen-config-dsl-schema`

### Phase 4: 交付

输出内容：
1. 完整 `aggregateTools` YAML 配置
2. 部署路径：`$HOME/.<binary-name>/config.yaml`
3. 与原始需求相比的差异/限制说明

---

## 核心规则

### 1. 引用语法

使用 `$` 前缀引用 pipeline 中的数据：

| 位置 | 格式 | 示例 |
|------|------|------|
| `spec.args` 字符串值 | `$root.path` | `$input.userId`, `$policy.application.id` |
| `spec.from` | `$stepId` 或 `$varName` | `$history`, `$component` |
| `spec.in` (foreach) | `$ref` | `$threatComponents` |
| `spec.vars` 值 | `$ref` | `vars: { scanId: $input.scanId }` |

`$` 后的第一个路径段为 root：
- `input` → 工具输入参数
- step `id` → 之前步骤的输出
- foreach `as` 名称 → 当前遍历元素

### 2. 流水线结构约束

- 每个步骤 `id` 唯一
- 顶层流水线必须以 `type: return` 结束
- foreach 子流水线必须以 `type: emit` 结束
- foreach 中不允许 `return`（用 `emit` 代替）
- foreach 中不允许嵌套 `foreach`

### 3. jq 表达式

`jq`、`return`、`emit` 步骤均支持 jq 表达式：

- `from` → jq 的输入数据 (`.`)
- `vars` → jq 变量 (`$varName`)
- `expr` → jq 表达式字符串

---

## 参考资源

| 资源 | 说明 |
|------|------|
| [dsl-schema.json](resources/dsl-schema.json) | **权威结构定义** — 由 `cmd/aggregate-tool-dsl-schema-gen/main.go` 从 Go struct 生成 |
| [bash-to-pipeline-mapping.md](references/bash-to-pipeline-mapping.md) | Bash/jq → DSL 翻译速查 |
| `make gen-config-dsl-schema` | 从 Go 源码重新生成 schema |
