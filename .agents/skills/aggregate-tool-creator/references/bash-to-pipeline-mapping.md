# Bash/jq → Pipeline DSL 映射参考

将 bash 脚本中常见的 API 编排模式翻译为 aggregator pipeline DSL。

## 单一 API 调用

**Bash**:
```bash
curl "$BASE/api/v2/apps/$ID" -o result.json
```

**Pipeline**:
```yaml
- name: getApp
  type: call
  call:
    tool: GetApplication
    args:
      applicationId: "{{ input.appId }}"
  output: app
```

## 链式调用（B 依赖 A 的返回值）

**Bash**:
```bash
APP="$(curl "$BASE/api/v2/apps/$PUBLIC_ID" | jq -r '.id')"
curl "$BASE/api/v2/apps/$APP/details"
```

**Pipeline**:
```yaml
- name: getApp
  type: call
  call:
    tool: GetApplication
    args:
      applicationId: "{{ input.publicAppId }}"
  output: app

- name: getDetails
  type: call
  call:
    tool: GetApplicationDetails
    args:
      internalAppId: "{{ app.id }}"
  output: details
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
- name: enrich
  type: map
  map:
    source: input.items
    pipeline:
      - name: getDetail
        type: call
        call:
          tool: GetItemDetail
          args:
            id: "{{ item.id }}"
        output: detail
      - name: done
        type: return
        return:
          source: detail.output
  output: results
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
- name: getData
  type: call
  call:
    tool: GetData
    args:
      id: "{{ input.dataId }}"
  output: data

- name: enrich
  type: map
  map:
    source: data.items
    pipeline:
      - name: snapshot
        type: transform
        transform:
          source: item
        output: orig

      - name: getDetail
        type: call
        call:
          tool: GetItemDetail
          args:
            id: "{{ item.id }}"
        output: detail

      - name: mergeDetail
        type: merge
        merge:
          from: "detail.output"
          to: "orig.output.detail"
        output: merged

      - name: done
        type: return
        return:
          source: merged.output
  output: enriched

- name: done
  type: return
  return:
    source: enriched.output
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
- name: project
  type: transform
  transform:
    source: data.output
    project: [name, email]
```

## 字段重命名

**Bash**:
```bash
jq '{display_name: .displayName, url: .packageUrl}' data.json
```

**Pipeline**:
```yaml
- name: rename
  type: transform
  transform:
    source: data.output
    rename:
      displayName: display_name
      packageUrl: url
```

## 删除字段

**Bash**:
```bash
jq 'del(.internal, ._links)' data.json
```

**Pipeline**:
```yaml
- name: cleanup
  type: transform
  transform:
    source: data.output
    remove: [internal, _links]
```

## 嵌套字段展平

**Bash**:
```bash
jq '. + .metadata' data.json | jq 'del(.metadata)'
```

**Pipeline**:
```yaml
- name: flatten
  type: transform
  transform:
    source: data.output
    flatten: [metadata]
```

## 数组第一个元素

**Bash**:
```bash
jq '.reports[0].stage' history.json
```

**Pipeline**:
```yaml
# 在路径中直接使用数字索引
source: history.reports.0.stage
```

## 构造顶部 JSON

**Bash**:
```bash
jq -n '{app: $app, items: $items, count: ($items | length)}'
```

**Pipeline**: 不支持。`transform` 只在已有对象上操作字段，不能构造全新的顶层结构。需在客户端后处理。

---

## 无法直接翻译的操作

以下 bash/jq 操作在 pipeline DSL 中**无法表达**：

| Bash/jq 操作 | 原因 | 建议 |
|-------------|------|------|
| `jq 'select(.field >= N)'` | 无条件过滤 | 客户端后处理 |
| `jq '.[] \| length'` | 无聚合函数 | 客户端后处理 |
| `jq 'max_by(.field)'` | 无排序/聚合 | 客户端后处理 |
| `jq -s '...'` | 无多输入合并构造 | 客户端后处理 |
| `if ...; then ...; fi` | 无条件分支 | 拆分为多个工具 |
| 字符串拼接/格式化 | 无字符串操作 | 客户端后处理 |
| `jq '.a + .b'` | 无算术运算 | 客户端后处理 |
