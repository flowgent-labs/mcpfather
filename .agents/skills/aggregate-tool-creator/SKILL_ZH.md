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
| `transform` | 声明式数据变换 (project/remove/rename/copy/move/flatten/default) | call 之后整理字段 |
| `map` | 遍历数组，对每个元素并发执行子流水线 | 对列表逐元素补充数据 |
| `merge` | 将值合并到目标对象的指定 key | map 子流水线中合并结果 |
| `return` | 返回最终结果 | 顶层和 map 子流水线**均必须以此结束** |

### Phase 3: Schema 校验（必须）

```bash
pip install check-jsonschema  # 仅需一次

check-jsonschema \
  --schemafile .agents/skills/aggregate-tool-creator/resources/dsl-schema.json \
  ~/.<binary-name>/config.yaml
```

Schema 过时时运行：`make gen-aggregatetool-dsl-schema`

### Phase 4: 交付

输出内容：
1. 完整 `aggregateTools` YAML 配置
2. 部署路径：`$HOME/.<binary-name>/config.yaml`
3. 与原始需求相比的差异/限制说明（尤其是 DSL 不支持的部分）

---

## 四条核心规则

### 1. 引用语法

| 位置 | 格式 | 示例 |
|------|------|------|
| `call.args` 字符串值 | `{{ root.path }}` | `"{{ input.userId }}"`, `"{{ policy.application.id }}"` |
| `source`/`from`/`to`/`return.source` | 裸路径（无 `{{ }}`） | `policy.components`, `comp.output.remediation` |

`{{ }}` 内的引用**根必须是 `output:` 名称**（不是 step `name`），否则 ValidateReferences 校验报错。

### 2. 路径导航

`.` 分隔字段链，数字自动解析为数组索引：`history.reports.0.stage` → `obj["reports"][0]["stage"]`。`.output` 是语法糖，自动跳过。

### 3. 流水线结构约束

- 每个步骤 `name` 唯一；每个 `output` 别名唯一
- 顶层流水线**和** map 子流水线都必须以 `type: return` 结束
- map 的 `source` 必须解析为数组

### 4. Transform 操作顺序

`project → remove → rename → copy → move → flatten → default`

---

## 四种编排模式

识别开发者的场景属于哪种模式，按对应骨架设计流水线：

| 模式 | 特征 | 流水线骨架 |
|------|------|-----------|
| **Chain** | A→B→C，后步依赖前步输出 | `call(A)→call(B,用A.id)→call(C)→return` |
| **Map-Enrich** | 获取列表→遍历每个元素→为每个调用补充API→合并 | `call(list)→map(source) { snapshot→call(detail)→merge→return }→return` |
| **Fan-Out** | 获取ID→分别获取多个数据源→合并 | `call(A)→call(B,用A.id)→call(C,用A.id)→transform→return`（顺序执行，不支持真正并行） |
| **List-Transform** | 获取列表→字段投影/重命名→直接返回 | `call(list)→transform(project/rename)→return` |

---

## DSL 能力边界

以下 bash/jq 操作在 pipeline DSL 中**无法表达**，必须告知开发者需客户端后处理：

| 操作 | 替代方案 |
|------|---------|
| `select(.field >= N)` 条件过滤 | 客户端过滤 |
| `max`/`min`/`length` 聚合计算 | 客户端计算 |
| `if/then/else` 条件分支 | 拆分为多个聚合工具 |
| 算术/字符串运算 | 客户端处理 |
| `jq -s` 自定义顶层 JSON 构造 | transform 仅支持字段级操作 |

---

## 参考资源

| 资源 | 说明 |
|------|------|
| [dsl-schema.json](resources/dsl-schema.json) | **权威结构定义** — 由 `cmd/aggregate-tool-dsl-schema-gen/main.go` 从 Go struct 生成 |
| [bash-to-pipeline-mapping.md](references/bash-to-pipeline-mapping.md) | Bash/jq → DSL 翻译速查 |
| `make gen-aggregatetool-dsl-schema` | 从 Go 源码重新生成 schema |
