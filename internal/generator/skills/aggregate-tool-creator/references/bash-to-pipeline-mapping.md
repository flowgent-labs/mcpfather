# Bash/jq → Pipeline DSL 映射参考

将 bash 脚本中常见的 API 编排模式翻译为 aggregator pipeline DSL。

## 单一 API 调用

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

## 链式调用（B 依赖 A 的返回值）

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
    args:
      applicationId: $input.publicAppId

- id: getDetails
  kind: call
  spec:
    tool: GetApplicationDetails
    args:
      internalAppId: $getApp.id
```

## 遍历列表 + 每个元素调用 API

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

## 获取 A → 遍历 A 的元素 → 每个元素调用 B → 合并结果

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

## 字段过滤 (select)

**Bash**:
```bash
jq '[.[] | select(.policyThreatLevel >= 5)]' data.json
```

**Pipeline** (jq 步骤):
```yaml
- id: filterThreat
  kind: jq
  spec:
    from: $data
    vars:
      min: $input.minThreatLevel
    expr: '[.[] | select(.policyThreatLevel >= $min)]'
```

## 字段投影（只保留部分字段）

**Bash**:
```bash
jq '{name, email}' data.json
# 或对数组:
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

## 字段重命名

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

## 删除字段

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

## 数组第一个元素

**Bash**:
```bash
jq '.reports[0].stage' history.json
```

**Pipeline**:
```yaml
- id: firstStage
  kind: jq
  spec:
    from: $history
    expr: '.reports[0].stage // null'
```

## 构造顶部 JSON

**Bash**:
```bash
jq -n '{app: $app, items: $items, count: ($items | length)}'
```

**Pipeline** (return 步骤 + jq expr):
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

## 聚合统计

**Bash**:
```bash
jq 'length' data.json
jq '[.[] | select(.waived != true)] | length' data.json
```

**Pipeline** (jq 步骤):
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

## 条件分支 (if/then/else)

**Bash**:
```bash
if [ "$(echo "$data" | jq '.count > 0')" = "true" ]; then
  # process
else
  echo '{"components":[],"totalCount":0}'
fi
```

**Pipeline** (jq 表达式内处理):
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

## 无法直接翻译的操作

| Bash/jq 操作 | 原因 | 建议 |
|-------------|------|------|
| `io.ReadAll` 大文件下载 | 无流式下载 | call 步骤本身处理 |
| 字符串拼接/格式化 | 无字符串操作 | jq 表达式内用 `+` 处理 |
