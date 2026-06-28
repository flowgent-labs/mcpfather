package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wl4g-ai/mcpgen/internal/generator/mcpaggregator/config"
	"github.com/wl4g-ai/mcpgen/internal/generator/mcpaggregator/engine"
	"github.com/wl4g-ai/mcpgen/internal/generator/mcpaggregator/pipeline"
)

// ===========================================================================
// Helpers for in-process scenario tests
// ===========================================================================

type scenarioRunner struct {
	registry *callRecorder
	executor *engine.Executor
}

type callRecorder struct {
	mu      sync.Mutex
	results map[string]string
	callLog []callEntry
	failOn  map[string]error
}

type callEntry struct {
	Tool string
	Args map[string]interface{}
}

func (r *callRecorder) CallTool(ctx context.Context, name string, args map[string]interface{}) (*pipeline.CallToolResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.callLog = append(r.callLog, callEntry{Tool: name, Args: args})
	if err, ok := r.failOn[name]; ok {
		return nil, err
	}
	text, ok := r.results[name]
	if !ok {
		return nil, fmt.Errorf("tool %q not found", name)
	}
	return &pipeline.CallToolResult{
		Content: []pipeline.ContentItem{{Type: "text", Text: text}},
	}, nil
}

func (r *callRecorder) calls() []callEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]callEntry, len(r.callLog))
	copy(out, r.callLog)
	return out
}

func (r *callRecorder) callCount(name string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, e := range r.callLog {
		if e.Tool == name {
			n++
		}
	}
	return n
}

func newScenarioRunner(results map[string]string) *scenarioRunner {
	r := &callRecorder{
		results: make(map[string]string),
		callLog: make([]callEntry, 0),
		failOn:  make(map[string]error),
	}
	for k, v := range results {
		r.results[k] = v
	}
	return &scenarioRunner{
		registry: r,
		executor: engine.NewExecutor(r),
	}
}

func (s *scenarioRunner) setFail(tool string, err error) {
	s.registry.failOn[tool] = err
}

func (s *scenarioRunner) execute(t *testing.T, steps []pipeline.StepConfig, input map[string]interface{}) *pipeline.CallToolResult {
	t.Helper()
	result, err := s.executor.Execute(context.Background(), steps, input)
	if err != nil {
		t.Fatalf("pipeline execution failed: %v", err)
	}
	return result
}

func (s *scenarioRunner) executeErr(t *testing.T, steps []pipeline.StepConfig, input map[string]interface{}) error {
	t.Helper()
	_, err := s.executor.Execute(context.Background(), steps, input)
	return err
}

func mustJSON(t *testing.T, s string) map[string]interface{} {
	t.Helper()
	var v map[string]interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, s)
	}
	return v
}

func mustJSONArray(t *testing.T, s string) []interface{} {
	t.Helper()
	var v []interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("invalid JSON array: %v\n%s", err, s)
	}
	return v
}

func resultText(t *testing.T, r *pipeline.CallToolResult) string {
	t.Helper()
	if len(r.Content) == 0 {
		t.Fatal("result has no content items")
	}
	return r.Content[0].Text
}

func resultJSON(t *testing.T, r *pipeline.CallToolResult) map[string]interface{} {
	t.Helper()
	return mustJSON(t, resultText(t, r))
}

func resultJSONArray(t *testing.T, r *pipeline.CallToolResult) []interface{} {
	t.Helper()
	return mustJSONArray(t, resultText(t, r))
}

// ===========================================================================
// SECTION 1-9: In-process pipeline engine scenario tests
// ===========================================================================

func TestScenario_ConfigLoad_ValidAggregatedTools(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
aggregateTools:
  - name: enrich_report
    description: Enrich a report with extra data
    inputSchema:
      type: object
      properties:
        reportId:
          type: string
    pipeline:
      - id: fetch_report
        kind: call
        spec:
          tool: getReport
          args:
            id: $input.reportId
      - id: fetch_metadata
        kind: call
        spec:
          tool: getMetadata
          args:
            id: $input.reportId
      - id: enrich
        kind: jq
        spec:
          from: $fetch_report
          vars:
            meta: $fetch_metadata
          expr: '. + {metadata: $meta}'
      - id: clean
        kind: jq
        spec:
          from: $enrich
          expr: 'del(.internalId, .debugInfo) | . + {status: "unknown"}'
      - id: done
        kind: return
        spec:
          from: $clean
  - name: health_summary
    description: Get health summary
    inputSchema:
      type: object
      properties: {}
    pipeline:
      - id: check
        kind: call
        spec:
          tool: healthCheck
          args: {}
      - id: done
        kind: return
        spec:
          from: $check
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if len(cfg.AggregateTools) != 2 {
		t.Fatalf("expected 2 aggregated tools, got %d", len(cfg.AggregateTools))
	}
	t1 := cfg.AggregateTools[0]
	if t1.Name != "enrich_report" {
		t.Errorf("tool name = %q, want enrich_report", t1.Name)
	}
	if len(t1.Pipeline) != 5 {
		t.Errorf("expected 5 pipeline steps, got %d", len(t1.Pipeline))
	}
	t2 := cfg.AggregateTools[1]
	if t2.Name != "health_summary" {
		t.Errorf("tool name = %q, want health_summary", t2.Name)
	}
	if len(t2.Pipeline) != 2 {
		t.Errorf("expected 2 pipeline steps, got %d", len(t2.Pipeline))
	}
}

func TestScenario_Validate_DuplicateStepIDs(t *testing.T) {
	steps := []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "t1", Args: map[string]interface{}{}}},
		{ID: "transform", Kind: "jq", Spec: pipeline.StepSpec{From: "$fetch", Expr: "."}},
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "t2", Args: map[string]interface{}{}}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$fetch"}},
	}
	if err := pipeline.Validate(steps); err == nil {
		t.Fatal("expected validation error for duplicate step ids")
	} else if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("error should mention 'duplicate', got: %v", err)
	}
}

func TestScenario_Validate_UnknownStepKind(t *testing.T) {
	steps := []pipeline.StepConfig{
		{ID: "step1", Kind: "call", Spec: pipeline.StepSpec{Tool: "t1", Args: map[string]interface{}{}}},
		{ID: "step2", Kind: "filter"},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$step1"}},
	}
	if err := pipeline.Validate(steps); err == nil {
		t.Fatal("expected validation error for unknown step kind")
	}
}

func TestScenario_Validate_MissingRequiredFields(t *testing.T) {
	steps := []pipeline.StepConfig{
		{ID: "step1", Kind: "call", Spec: pipeline.StepSpec{Tool: "", Args: map[string]interface{}{}}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$step1"}},
	}
	if err := pipeline.Validate(steps); err == nil {
		t.Fatal("expected validation error for call with empty tool")
	}
}

func TestScenario_Context_ResolveNestedInputPaths(t *testing.T) {
	ctx := engine.NewContext(map[string]interface{}{
		"user": map[string]interface{}{
			"profile": map[string]interface{}{"email": "alice@example.com", "memberId": 12345},
		},
		"filters": []interface{}{"active", "verified"},
	})
	tests := []struct {
		path     string
		expected interface{}
	}{
		{"$input.user.profile.email", "alice@example.com"},
		{"$input.user.profile.memberId", 12345},
		{"$input.filters", []interface{}{"active", "verified"}},
	}
	for _, tc := range tests {
		v, err := ctx.Resolve(tc.path)
		if err != nil {
			t.Errorf("resolve %q: %v", tc.path, err)
			continue
		}
		if fmt.Sprintf("%v", v) != fmt.Sprintf("%v", tc.expected) {
			t.Errorf("resolve %q: got %v, want %v", tc.path, v, tc.expected)
		}
	}
}

func TestScenario_Context_ResolveWithArrayIndexNavigation(t *testing.T) {
	ctx := engine.NewContext(map[string]interface{}{
		"items": []interface{}{map[string]interface{}{"name": "first"}, map[string]interface{}{"name": "second"}},
	})
	v, err := ctx.ResolvePath("$input.items.0.name")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if v != "first" {
		t.Errorf("expected 'first', got %v", v)
	}
}

func TestScenario_Context_ItemContextInForeachIteration(t *testing.T) {
	baseCtx := engine.NewContext(map[string]interface{}{"batch": "batch-001"})
	itemCtx := baseCtx.WithItem(map[string]interface{}{"id": "item-42", "data": map[string]interface{}{"color": "red"}}, "item")
	v, err := itemCtx.Resolve("$item.id")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if v != "item-42" {
		t.Errorf("expected 'item-42', got %v", v)
	}
	v, err = itemCtx.ResolvePath("$item.data.color")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if v != "red" {
		t.Errorf("expected 'red', got %v", v)
	}
}

func TestScenario_Context_UnresolvedReferenceErrors(t *testing.T) {
	ctx := engine.NewContext(map[string]interface{}{"name": "test"})
	_, err := ctx.Resolve("$nonexistent.field")
	if err == nil {
		t.Fatal("expected error for unresolved step reference")
	}
	if !strings.Contains(err.Error(), "unresolved") {
		t.Errorf("error should say 'unresolved', got: %v", err)
	}
}

func TestScenario_Context_ResolveMapWithMixedReferences(t *testing.T) {
	ctx := engine.NewContext(map[string]interface{}{"app": "myapp", "env": "production"})
	resolved, err := ctx.ResolveMap(map[string]interface{}{
		"appName": "$input.app",
		"nested":  map[string]interface{}{"env": "$input.env"},
	})
	if err != nil {
		t.Fatalf("ResolveMap: %v", err)
	}
	if resolved["appName"] != "myapp" {
		t.Errorf("appName = %v", resolved["appName"])
	}
}

func TestScenario_Call_SimpleArgResolution(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getUser": `{"id":"123","name":"Alice","role":"admin"}`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getUser", Args: map[string]interface{}{"userId": "$input.id"}}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$fetch"}},
	}, map[string]interface{}{"id": "123"})
	data := resultJSON(t, result)
	if data["name"] != "Alice" {
		t.Errorf("name = %v", data["name"])
	}
}

func TestScenario_Call_NestedObjectArgResolution(t *testing.T) {
	s := newScenarioRunner(map[string]string{"search": `{"total":5}`})
	_ = s.execute(t, []pipeline.StepConfig{
		{ID: "query", Kind: "call", Spec: pipeline.StepSpec{
			Tool: "search",
			Args: map[string]interface{}{"filter": map[string]interface{}{"app": "$input.app"}, "sort": "$input.sortBy"},
		}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$query"}},
	}, map[string]interface{}{"app": "catalog", "sortBy": "name"})
	calls := s.registry.calls()
	filter := calls[0].Args["filter"].(map[string]interface{})
	if filter["app"] != "catalog" {
		t.Errorf("filter.app = %v", filter["app"])
	}
}

func TestScenario_Call_ErrorPropagation(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getUser": `{}`})
	s.setFail("getUser", fmt.Errorf("upstream 503 Service Unavailable"))
	err := s.executeErr(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getUser", Args: map[string]interface{}{}}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$fetch"}},
	}, map[string]interface{}{})
	if err == nil || !strings.Contains(err.Error(), "503") {
		t.Errorf("expected 503 error, got %v", err)
	}
}

func TestScenario_JQ_ProjectFields(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getReport": `{"id":"r1","title":"Q4","author":"Bob","internalId":"secret","score":85}`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getReport", Args: map[string]interface{}{}}},
		{ID: "select", Kind: "jq", Spec: pipeline.StepSpec{From: "$fetch", Expr: "{id, title, score}"}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$select"}},
	}, nil)
	data := resultJSON(t, result)
	if _, ok := data["internalId"]; ok {
		t.Error("internalId should be projected out")
	}
	if len(data) != 3 {
		t.Errorf("expected 3 fields, got %d", len(data))
	}
}

func TestScenario_JQ_ProjectArrayOfObjects(t *testing.T) {
	s := newScenarioRunner(map[string]string{"listUsers": `[{"id":"1","name":"A","password":"xx"},{"id":"2","name":"B","password":"yy"}]`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "listUsers", Args: map[string]interface{}{}}},
		{ID: "sanitize", Kind: "jq", Spec: pipeline.StepSpec{From: "$fetch", Expr: "[.[] | {id, name}]"}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$sanitize"}},
	}, nil)
	items := resultJSONArray(t, result)
	for i, item := range items {
		m := item.(map[string]interface{})
		if _, ok := m["password"]; ok {
			t.Errorf("item[%d]: password should be removed", i)
		}
	}
}

func TestScenario_JQ_RemoveSensitiveFields(t *testing.T) {
	ctx := engine.NewContext(nil)
	ctx.SetOutput("data", map[string]interface{}{"username": "alice", "password": "secret123", "token": "abc", "email": "alice@x.com"})
	s := &pipeline.StepConfig{ID: "clean", Kind: "jq", Spec: pipeline.StepSpec{From: "$data", Expr: `del(.password, .token)`}}
	result, err := engine.NewExecutor(nil).ExecuteStep(context.Background(), s, ctx)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if _, ok := m["password"]; ok {
		t.Error("password should be removed")
	}
}

func TestScenario_JQ_RenameFields(t *testing.T) {
	ctx := engine.NewContext(nil)
	ctx.SetOutput("data", map[string]interface{}{"first_name": "John", "last_name": "Doe"})
	s := &pipeline.StepConfig{ID: "rename", Kind: "jq", Spec: pipeline.StepSpec{From: "$data", Expr: `{firstName: .first_name, lastName: .last_name}`}}
	result, err := engine.NewExecutor(nil).ExecuteStep(context.Background(), s, ctx)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["firstName"] != "John" || m["lastName"] != "Doe" {
		t.Errorf("rename failed: %v", m)
	}
}

func TestScenario_JQ_FlattenNestedStructures(t *testing.T) {
	ctx := engine.NewContext(nil)
	ctx.SetOutput("data", map[string]interface{}{"name": "report", "metadata": map[string]interface{}{"author": "Alice", "version": 2}, "stats": map[string]interface{}{"views": 100}})
	s := &pipeline.StepConfig{ID: "flat", Kind: "jq", Spec: pipeline.StepSpec{From: "$data", Expr: `. + .metadata | del(.metadata)`}}
	result, err := engine.NewExecutor(nil).ExecuteStep(context.Background(), s, ctx)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["author"] != "Alice" {
		t.Errorf("author = %v", m["author"])
	}
	if _, ok := m["metadata"]; ok {
		t.Error("metadata should be gone")
	}
}

func TestScenario_JQ_DefaultValues(t *testing.T) {
	ctx := engine.NewContext(nil)
	ctx.SetOutput("data", map[string]interface{}{"name": "existing"})
	s := &pipeline.StepConfig{ID: "def", Kind: "jq", Spec: pipeline.StepSpec{From: "$data", Expr: `{version: "1.0"} + .`}}
	result, err := engine.NewExecutor(nil).ExecuteStep(context.Background(), s, ctx)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["name"] != "existing" {
		t.Error("existing value should not be overwritten")
	}
	if m["version"] != "1.0" {
		t.Error("missing field should get default")
	}
}

func TestScenario_JQ_ChainedOperations(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getData": `{"first_name":"Jane","last_name":"Smith","password":"xxx","address":{"city":"NYC"}}`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getData", Args: map[string]interface{}{}}},
		{ID: "rename", Kind: "jq", Spec: pipeline.StepSpec{From: "$fetch", Expr: `{firstName: .first_name, lastName: .last_name, password, address}`}},
		{ID: "sanitize", Kind: "jq", Spec: pipeline.StepSpec{From: "$rename", Expr: `del(.password)`}},
		{ID: "flatten", Kind: "jq", Spec: pipeline.StepSpec{From: "$sanitize", Expr: `(. + .address | del(.address)) + {role: "user"}`}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$flatten"}},
	}, nil)
	data := resultJSON(t, result)
	if data["firstName"] != "Jane" {
		t.Errorf("firstName = %v", data["firstName"])
	}
	if data["city"] != "NYC" {
		t.Errorf("city = %v", data["city"])
	}
	if data["role"] != "user" {
		t.Errorf("role = %v", data["role"])
	}
}

func TestScenario_FullPipeline_CallJQReturn(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getData": `{"id":"r1"}`, "getMeta": `{"author":"Alice"}`, "getStats": `{"score":95}`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch_base", Kind: "call", Spec: pipeline.StepSpec{Tool: "getData", Args: map[string]interface{}{}}},
		{ID: "fetch_meta", Kind: "call", Spec: pipeline.StepSpec{Tool: "getMeta", Args: map[string]interface{}{"id": "$fetch_base.id"}}},
		{ID: "fetch_stats", Kind: "call", Spec: pipeline.StepSpec{Tool: "getStats", Args: map[string]interface{}{"id": "$fetch_base.id"}}},
		{ID: "merge_all", Kind: "jq", Spec: pipeline.StepSpec{From: "$fetch_base", Vars: map[string]interface{}{"meta": "$fetch_meta", "stats": "$fetch_stats"}, Expr: `. + {metadata: $meta, statistics: $stats}`}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$merge_all"}},
	}, nil)
	data := resultJSON(t, result)
	if data["id"] != "r1" {
		t.Errorf("id = %v, want r1", data["id"])
	}
	meta, ok := data["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("metadata field missing from result")
	}
	if meta["author"] != "Alice" {
		t.Errorf("metadata.author = %v", meta["author"])
	}
}

func TestScenario_Return_StringValue(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getStatus": `OK`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getStatus", Args: map[string]interface{}{}}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$fetch"}},
	}, nil)
	if !strings.Contains(resultText(t, result), "OK") {
		t.Errorf("expected OK, got %q", resultText(t, result))
	}
}

func TestScenario_Return_ComplexObjectAsJSON(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getConfig": `{"features":{"darkMode":true}}`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getConfig", Args: map[string]interface{}{}}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$fetch"}},
	}, nil)
	data := resultJSON(t, result)
	if data["features"].(map[string]interface{})["darkMode"] != true {
		t.Error("darkMode should be true")
	}
}

func TestScenario_Return_WithJQExpr(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getReport": `{"id":"r1","title":"Report","created":"2024-01-01"}`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getReport", Args: map[string]interface{}{}}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$fetch", Expr: "{title, created}"}},
	}, nil)
	data := resultJSON(t, result)
	if len(data) != 2 || data["title"] != "Report" {
		t.Errorf("expected {title, created}, got %v", data)
	}
}

func TestScenario_Foreach_CallSubPipelinePerItem(t *testing.T) {
	s := newScenarioRunner(map[string]string{"lookupName": `Alice`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "batch", Kind: "foreach", Spec: pipeline.StepSpec{
			In:            "$input.ids",
			As:            "item",
			Concurrency:   2,
			PreserveOrder: true,
			Pipeline: []pipeline.StepConfig{
				{ID: "lookup", Kind: "call", Spec: pipeline.StepSpec{Tool: "lookupName", Args: map[string]interface{}{"id": "$item"}}},
				{ID: "ret", Kind: "emit", Spec: pipeline.StepSpec{From: "$lookup"}},
			},
		}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$batch"}},
	}, map[string]interface{}{"ids": []interface{}{"1", "2", "3"}})
	if len(resultJSONArray(t, result)) != 3 {
		t.Fatal("expected 3 results")
	}
	if s.registry.callCount("lookupName") != 3 {
		t.Errorf("expected 3 calls, got %d", s.registry.callCount("lookupName"))
	}
}

func TestScenario_Foreach_TransformSubPipelinePerItem(t *testing.T) {
	s := newScenarioRunner(map[string]string{"fetchItem": `{"name":"test","password":"secret","role":"admin"}`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "process", Kind: "foreach", Spec: pipeline.StepSpec{
			In:            "$input.items",
			As:            "item",
			Concurrency:   2,
			PreserveOrder: true,
			Pipeline: []pipeline.StepConfig{
				{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "fetchItem", Args: map[string]interface{}{"id": "$item"}}},
				{ID: "ret", Kind: "emit", Spec: pipeline.StepSpec{From: "$fetch", Expr: "{name, role}"}},
			},
		}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$process"}},
	}, map[string]interface{}{"items": []interface{}{"a", "b"}})
	items := resultJSONArray(t, result)
	for i, item := range items {
		m := item.(map[string]interface{})
		if _, ok := m["password"]; ok {
			t.Errorf("item[%d]: password should be removed", i)
		}
	}
}

func TestScenario_Foreach_ErrorPropagationStopsExecution(t *testing.T) {
	s := newScenarioRunner(map[string]string{"riskyCall": `OK`})
	s.setFail("riskyCall", fmt.Errorf("upstream timeout on item-2"))
	err := s.executeErr(t, []pipeline.StepConfig{
		{ID: "danger", Kind: "foreach", Spec: pipeline.StepSpec{
			In:            "$input.items",
			As:            "item",
			Concurrency:   1,
			PreserveOrder: true,
			Pipeline: []pipeline.StepConfig{
				{ID: "call", Kind: "call", Spec: pipeline.StepSpec{Tool: "riskyCall", Args: map[string]interface{}{"id": "$item"}}},
				{ID: "ret", Kind: "emit", Spec: pipeline.StepSpec{From: "$call"}},
			},
		}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$danger"}},
	}, map[string]interface{}{"items": []interface{}{1, 2, 3}})
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout error, got %v", err)
	}
}

func TestScenario_Foreach_EmptyListReturnsEmptyResults(t *testing.T) {
	s := newScenarioRunner(map[string]string{"never": "called"})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "empty", Kind: "foreach", Spec: pipeline.StepSpec{
			In:            "$input.items",
			As:            "item",
			Concurrency:   2,
			PreserveOrder: true,
			Pipeline: []pipeline.StepConfig{
				{ID: "call", Kind: "call", Spec: pipeline.StepSpec{Tool: "never", Args: map[string]interface{}{"id": "$item"}}},
				{ID: "ret", Kind: "emit", Spec: pipeline.StepSpec{From: "$call"}},
			},
		}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$empty"}},
	}, map[string]interface{}{"items": []interface{}{}})
	if len(resultJSONArray(t, result)) != 0 {
		t.Error("expected empty results")
	}
}

func TestScenario_Foreach_PreserveOrder(t *testing.T) {
	s := newScenarioRunner(map[string]string{"indexItem": `{"value":"x"}`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "ordered", Kind: "foreach", Spec: pipeline.StepSpec{
			In:            "$input.ids",
			As:            "item",
			Concurrency:   4,
			PreserveOrder: true,
			Pipeline: []pipeline.StepConfig{
				{ID: "index", Kind: "call", Spec: pipeline.StepSpec{Tool: "indexItem", Args: map[string]interface{}{"id": "$item"}}},
				{ID: "ret", Kind: "emit", Spec: pipeline.StepSpec{From: "$index"}},
			},
		}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$ordered"}},
	}, map[string]interface{}{"ids": []interface{}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}})
	if len(resultJSONArray(t, result)) != 10 {
		t.Fatal("expected 10 results")
	}
}

func TestScenario_FullPipeline_CallJQReturn_Project(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getViolations": `{"total":10,"violations":[{"severity":"HIGH"},{"severity":"LOW"}],"scanId":"scan-001"}`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getViolations", Args: map[string]interface{}{"scanId": "$input.scanId"}}},
		{ID: "clean", Kind: "jq", Spec: pipeline.StepSpec{From: "$fetch", Expr: `{violations} + {total: 0}`}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$clean"}},
	}, map[string]interface{}{"scanId": "scan-001"})
	data := resultJSON(t, result)
	if _, ok := data["scanId"]; ok {
		t.Error("scanId should be projected out")
	}
	if len(data["violations"].([]interface{})) != 2 {
		t.Errorf("expected 2 violations")
	}
}

func TestScenario_FullPipeline_CallForeachJQEmitReturn(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getComponents": `[{"id":"c1","name":"auth","owner":"team-A","internalRef":"x-1"},{"id":"c2","name":"ui","owner":"team-B","internalRef":"x-2"}]`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch_all", Kind: "call", Spec: pipeline.StepSpec{Tool: "getComponents", Args: map[string]interface{}{}}},
		{ID: "enrich_each", Kind: "foreach", Spec: pipeline.StepSpec{
			In:            "$fetch_all",
			As:            "item",
			Concurrency:   2,
			PreserveOrder: true,
			Pipeline: []pipeline.StepConfig{
				{ID: "emit_item", Kind: "emit", Spec: pipeline.StepSpec{From: "$item", Expr: `{id, name, maintainer: .owner, status: "active"}`}},
			},
		}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$enrich_each"}},
	}, nil)
	items := resultJSONArray(t, result)
	first := items[0].(map[string]interface{})
	if first["maintainer"] != "team-A" {
		t.Errorf("maintainer = %v", first["maintainer"])
	}
	if _, ok := first["owner"]; ok {
		t.Error("owner should be renamed")
	}
}

func TestScenario_FullPipeline_CallForeachCallJQEmitReturn(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getReports": `[{"id":"r1"},{"id":"r2"}]`, "getAnnotations": `{"color":"blue"}`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch_reports", Kind: "call", Spec: pipeline.StepSpec{Tool: "getReports", Args: map[string]interface{}{}}},
		{ID: "annotate_each", Kind: "foreach", Spec: pipeline.StepSpec{
			In:            "$fetch_reports",
			As:            "item",
			Concurrency:   2,
			PreserveOrder: true,
			Pipeline: []pipeline.StepConfig{
				{ID: "get_annotations", Kind: "call", Spec: pipeline.StepSpec{Tool: "getAnnotations", Args: map[string]interface{}{"id": "$item.id"}}},
				{ID: "emit_merged", Kind: "emit", Spec: pipeline.StepSpec{From: "$item", Vars: map[string]interface{}{"ann": "$get_annotations"}, Expr: `. + {annotationData: $ann}`}},
			},
		}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$annotate_each"}},
	}, nil)
	if len(resultJSONArray(t, result)) != 2 {
		t.Fatal("expected 2 results")
	}
}

func TestScenario_FullPipeline_PipelineWithoutReturnErrors(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getData": `{}`})
	err := s.executeErr(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getData", Args: map[string]interface{}{}}},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "without a return step") {
		t.Errorf("expected missing-return error, got %v", err)
	}
}

func TestScenario_FullPipeline_InvalidReferenceFails(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getData": `{"id":"123"}`})
	err := s.executeErr(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getData", Args: map[string]interface{}{}}},
		{ID: "clean", Kind: "jq", Spec: pipeline.StepSpec{From: "$nonexistent_step", Expr: "."}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$clean"}},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "unresolved") {
		t.Errorf("expected unresolved error, got %v", err)
	}
}

func TestScenario_FullPipeline_MultipleAggregatedToolsInEngine(t *testing.T) {
	cfg := &config.Config{
		AggregateTools: []config.AggregatedToolConfig{
			{Name: "tool_a", InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}, Pipeline: []pipeline.StepConfig{
				{ID: "step1", Kind: "call", Spec: pipeline.StepSpec{Tool: "nativeA", Args: map[string]interface{}{"src": "a"}}},
				{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$step1"}},
			}},
			{Name: "tool_b", InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}, Pipeline: []pipeline.StepConfig{
				{ID: "step1", Kind: "call", Spec: pipeline.StepSpec{Tool: "nativeB", Args: map[string]interface{}{"src": "b"}}},
				{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$step1"}},
			}},
		},
	}
	rec := &callRecorder{results: map[string]string{"nativeA": "A", "nativeB": "B"}, callLog: make([]callEntry, 0), failOn: make(map[string]error)}
	eng, err := engine.NewFromConfig(cfg, rec)
	if err != nil {
		t.Fatalf("NewFromConfig: %v", err)
	}
	tools, _ := eng.Tools()
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
	r1, _ := tools[0].Handler(context.Background(), nil)
	r2, _ := tools[1].Handler(context.Background(), nil)
	if r1.Content[0].Text != "A" || r2.Content[0].Text != "B" {
		t.Error("handler results mismatch")
	}
}

func TestScenario_EdgeCase_NonJSONCallResultWrapsAsText(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getMessage": `plain text response`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getMessage", Args: map[string]interface{}{}}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$fetch"}},
	}, nil)
	if resultText(t, result) != "plain text response" {
		t.Errorf("unexpected result text: %s", resultText(t, result))
	}
}

func TestScenario_EdgeCase_CallWithNoArgs(t *testing.T) {
	s := newScenarioRunner(map[string]string{"ping": `{"status":"ok"}`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "pinger", Kind: "call", Spec: pipeline.StepSpec{Tool: "ping", Args: map[string]interface{}{}}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$pinger"}},
	}, nil)
	if resultJSON(t, result)["status"] != "ok" {
		t.Error("expected ok")
	}
}

func TestScenario_Require_NonEmpty(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getData": `null`})
	err := s.executeErr(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getData", Args: map[string]interface{}{}},
			Require: &pipeline.RequireConfig{NonEmpty: true, Message: "Data must not be empty"}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$fetch"}},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "Data must not be empty") {
		t.Errorf("expected require validation error, got %v", err)
	}
}

func TestScenario_Require_NonEmptyOnJQ(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getData": `{"items":[]}`})
	err := s.executeErr(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getData", Args: map[string]interface{}{}}},
		{ID: "extract", Kind: "jq", Spec: pipeline.StepSpec{From: "$fetch", Expr: ".items"},
			Require: &pipeline.RequireConfig{NonEmpty: true, Message: "Items must not be empty"}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$extract"}},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "Items must not be empty") {
		t.Errorf("expected require validation error on jq step, got %v", err)
	}
}

func TestScenario_Call_ParseJSON(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getJSONString": `{"nested":{"key":"value"}}`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getJSONString", Parse: "json", Args: map[string]interface{}{}}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$fetch", Expr: ".nested.key"}},
	}, nil)
	if resultText(t, result) != "value" {
		t.Errorf("expected 'value' from parsed JSON, got %q", resultText(t, result))
	}
}

func TestScenario_Call_ParseJSON_NoParseFlag(t *testing.T) {
	// Without parse:json, auto-detection still parses JSON content-type responses
	// but this test verifies the behavior for non-JSON content-type strings
	s := newScenarioRunner(map[string]string{"getPlainText": `plain text, not json`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getPlainText", Args: map[string]interface{}{}}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$fetch"}},
	}, nil)
	if resultText(t, result) != "plain text, not json" {
		t.Errorf("without parse:json, plain text should be returned as-is, got %q", resultText(t, result))
	}
}

func TestScenario_Call_BodyObjectArg(t *testing.T) {
	s := newScenarioRunner(map[string]string{"createItem": `{"id":"new-1","status":"created"}`})
	_ = s.execute(t, []pipeline.StepConfig{
		{ID: "create", Kind: "call", Spec: pipeline.StepSpec{
			Tool: "createItem",
			Args: map[string]interface{}{
				"body": map[string]interface{}{"name": "$input.itemName", "tags": []interface{}{"$input.tag"}},
				"query": map[string]interface{}{"dryRun": true},
			},
		}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$create"}},
	}, map[string]interface{}{"itemName": "test-item", "tag": "urgent"})
	calls := s.registry.calls()
	body := calls[0].Args["body"].(map[string]interface{})
	if body["name"] != "test-item" {
		t.Errorf("body.name = %v, want test-item", body["name"])
	}
	tags := body["tags"].([]interface{})
	if tags[0] != "urgent" {
		t.Errorf("tags[0] = %v, want urgent", tags[0])
	}
	query := calls[0].Args["query"].(map[string]interface{})
	if query["dryRun"] != true {
		t.Errorf("query.dryRun = %v, want true", query["dryRun"])
	}
}

func TestScenario_Foreach_ConcurrencyFromInput(t *testing.T) {
	s := newScenarioRunner(map[string]string{"procItem": `{"result":"ok"}`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "batch", Kind: "foreach", Spec: pipeline.StepSpec{
			In:            "$input.ids",
			As:            "item",
			Concurrency:   "$input.concurrency",
			PreserveOrder: true,
			Pipeline: []pipeline.StepConfig{
				{ID: "proc", Kind: "call", Spec: pipeline.StepSpec{Tool: "procItem", Args: map[string]interface{}{"id": "$item"}}},
				{ID: "out", Kind: "emit", Spec: pipeline.StepSpec{From: "$proc"}},
			},
		}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$batch"}},
	}, map[string]interface{}{"ids": []interface{}{1, 2, 3, 4, 5}, "concurrency": 3})
	if len(resultJSONArray(t, result)) != 5 {
		t.Fatal("expected 5 results")
	}
}

func TestScenario_Return_WithVarsAndComplexExpr(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getReport": `{"id":"r1","title":"Q4 Report","score":95}`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch_report", Kind: "call", Spec: pipeline.StepSpec{Tool: "getReport", Args: map[string]interface{}{}}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{
			From: "$fetch_report",
			Vars: map[string]interface{}{
				"reportId":    "$fetch_report.id",
				"reportTitle": "$fetch_report.title",
			},
			Expr: `{summary: {source_report_id: $reportId, source_report_title: $reportTitle, score: .score}, generated_at: "now"}`,
		}},
	}, nil)
	data := resultJSON(t, result)
	summary, ok := data["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("summary field missing or not an object")
	}
	if summary["source_report_id"] != "r1" {
		t.Errorf("source_report_id = %v", summary["source_report_id"])
	}
	if summary["score"] != float64(95) {
		t.Errorf("score = %v", summary["score"])
	}
	if data["generated_at"] != "now" {
		t.Errorf("generated_at = %v", data["generated_at"])
	}
}

func TestScenario_FullPipeline_CallWithJQRequire(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getData": `[{"name":"valid","status":"ok"},{"name":"alsoOK","status":"good"}]`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch", Kind: "call", Spec: pipeline.StepSpec{Tool: "getData", Args: map[string]interface{}{}},
			Require: &pipeline.RequireConfig{NonEmpty: true, Message: "Fetch result must not be empty"}},
		{ID: "filter", Kind: "jq", Spec: pipeline.StepSpec{From: "$fetch", Expr: `[.[] | select(.status == "ok")]`},
			Require: &pipeline.RequireConfig{NonEmpty: true, Message: "No items with status=ok"}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$filter"}},
	}, nil)
	items := resultJSONArray(t, result)
	if len(items) != 1 {
		t.Fatalf("expected 1 filtered item, got %d", len(items))
	}
}

func TestScenario_JQ_WithVarsFromMultipleSources(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getUser": `{"id":"u1","name":"Alice"}`, "getDept": `{"id":"d1","title":"Engineering"}`})
	result := s.execute(t, []pipeline.StepConfig{
		{ID: "fetch_user", Kind: "call", Spec: pipeline.StepSpec{Tool: "getUser", Args: map[string]interface{}{}}},
		{ID: "fetch_dept", Kind: "call", Spec: pipeline.StepSpec{Tool: "getDept", Args: map[string]interface{}{}}},
		{ID: "merge", Kind: "jq", Spec: pipeline.StepSpec{
			From: "$fetch_user",
			Vars: map[string]interface{}{"deptInfo": "$fetch_dept", "env": "production"},
			Expr: `{name, department: $deptInfo.title, environment: $env}`,
		}},
		{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$merge"}},
	}, nil)
	data := resultJSON(t, result)
	if data["name"] != "Alice" {
		t.Errorf("name = %v", data["name"])
	}
	if data["department"] != "Engineering" {
		t.Errorf("department = %v", data["department"])
	}
	if data["environment"] != "production" {
		t.Errorf("environment = %v", data["environment"])
	}
}

// ===========================================================================
// SECTION 10: End-to-End Integration Tests (generate → build → run → verify)
// ===========================================================================

func writeAggregatedConfig(t *testing.T, homeDir, binaryName, yamlContent string) {
	t.Helper()
	configDir := filepath.Join(homeDir, "."+binaryName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
}

func mcpCallAggregatedTool(t *testing.T, baseURL string, toolName string, args map[string]interface{}) string {
	t.Helper()
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      toolName,
		"arguments": args,
	})
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var rpcResp struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		t.Fatalf("failed to parse MCP response: %v\nbody: %s", err, body)
	}
	if rpcResp.Error != nil {
		t.Fatalf("MCP error: %s (code %d)", rpcResp.Error.Message, rpcResp.Error.Code)
	}
	if rpcResp.Result.IsError {
		for _, c := range rpcResp.Result.Content {
			t.Logf("tool error: %s", c.Text)
		}
		t.Fatal("aggregated tool returned error")
	}
	if len(rpcResp.Result.Content) == 0 {
		t.Fatal("empty content in result")
	}
	return rpcResp.Result.Content[0].Text
}

func startAggTestServer(t *testing.T, projectDir string, mockURL string, homeDir string) (cleanup func(), baseURL string) {
	t.Helper()
	binPath := buildServer(t, projectDir)
	port := fmt.Sprintf("%d", 19000+(time.Now().UnixNano()%1000))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port, "-v", "1")
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"MCP_UPSTREAM_ENDPOINT="+mockURL,
	)
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start HTTP server: %v", err)
	}

	cleanup = func() {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}
	baseURL = "http://localhost:" + port
	waitForServer(t, baseURL)
	return
}

// ---------------------------------------------------------------------------
// E2E Test 1: Aggregated tool with call → jq(project) → return
// ---------------------------------------------------------------------------

func TestE2E_AggregatedTool_CallJQReturn(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","message":"hello","internalToken":"secret123","timestamp":"2024-01-01T00:00:00Z"}`))
	})
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	aggConfig := `
aggregateTools:
  - name: agg_clean_echo
    description: Echo with cleanup
    inputSchema:
      type: object
      properties: {}
    pipeline:
      - id: fetch
        kind: call
        spec:
          tool: EchoHeaders
          args: {}
      - id: clean
        kind: jq
        spec:
          from: $fetch
          expr: '{status, message, timestamp}'
      - id: done
        kind: return
        spec:
          from: $clean
`
	writeAggregatedConfig(t, homeDir, binaryName, aggConfig)

	cleanup, baseURL := startAggTestServer(t, dir, mock.server.URL, homeDir)
	defer cleanup()

	result := mcpCallAggregatedTool(t, baseURL, "agg_clean_echo", map[string]interface{}{})

	data := mustJSON(t, result)
	if data["status"] != "ok" {
		t.Errorf("status = %v, want ok", data["status"])
	}
	if _, ok := data["internalToken"]; ok {
		t.Error("internalToken should have been projected out")
	}
	if len(mock.requests) != 1 {
		t.Errorf("expected 1 upstream request, got %d", len(mock.requests))
	}
}

// ---------------------------------------------------------------------------
// E2E Test 2: Aggregated tool chaining two native tools with jq merge
// ---------------------------------------------------------------------------

func TestE2E_AggregatedTool_ChainedNativeTools(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/echo") {
			w.Write([]byte(`{"echo_status":"done","trace_id":"abc-123"}`))
		} else if strings.Contains(r.URL.Path, "/hello") {
			name := r.URL.Query().Get("name")
			w.Write([]byte(fmt.Sprintf(`{"greeting":"Hello, %s!","code":200}`, name)))
		} else {
			w.Write([]byte(`{}`))
		}
	})
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	aggConfig := `
aggregateTools:
  - name: agg_chain
    description: Chain echo and greet
    inputSchema:
      type: object
      properties:
        name:
          type: string
      required:
        - name
    pipeline:
      - id: echo
        kind: call
        spec:
          tool: EchoHeaders
          args: {}
      - id: greet
        kind: call
        spec:
          tool: SayHello
          args:
            name: $input.name
      - id: merge_step
        kind: jq
        spec:
          from: $echo
          vars:
            g: $greet
          expr: '. + {greeting_data: $g}'
      - id: done
        kind: return
        spec:
          from: $merge_step
`
	writeAggregatedConfig(t, homeDir, binaryName, aggConfig)

	cleanup, baseURL := startAggTestServer(t, dir, mock.server.URL, homeDir)
	defer cleanup()

	result := mcpCallAggregatedTool(t, baseURL, "agg_chain", map[string]interface{}{
		"name": "World",
	})

	data := mustJSON(t, result)
	if data["echo_status"] != "done" {
		t.Errorf("echo_status = %v", data["echo_status"])
	}

	greetingData, ok := data["greeting_data"]
	if !ok {
		t.Fatal("greeting_data not found — merge failed")
	}
	gd := greetingData.(map[string]interface{})
	if gd["greeting"] != "Hello, World!" {
		t.Errorf("greeting = %v", gd["greeting"])
	}

	if len(mock.requests) != 2 {
		t.Errorf("expected 2 upstream requests (echo + hello), got %d", len(mock.requests))
	}
}

// ---------------------------------------------------------------------------
// E2E Test 3: Foreach over input array — calls SayHello for each name
// ---------------------------------------------------------------------------

func TestE2E_AggregatedTool_ForeachOverInputArray(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		name := r.URL.Query().Get("name")
		w.Write([]byte(fmt.Sprintf(`{"greeting":"Hi, %s!","length":%d}`, name, len(name))))
	})
	defer mock.Close()

	dir := genProject(t, "sayHello", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	aggConfig := `
aggregateTools:
  - name: agg_batch_greet
    description: Batch greet
    inputSchema:
      type: object
      properties:
        names:
          type: array
          items:
            type: string
      required:
        - names
    pipeline:
      - id: process_all
        kind: foreach
        spec:
          in: $input.names
          as: name
          concurrency: 2
          preserveOrder: true
          pipeline:
            - id: greet_one
              kind: call
              spec:
                tool: SayHello
                args:
                  name: $name
            - id: emit_greeting
              kind: emit
              spec:
                from: $greet_one
                expr: '{greeting}'
      - id: done
        kind: return
        spec:
          from: $process_all
`
	writeAggregatedConfig(t, homeDir, binaryName, aggConfig)

	cleanup, baseURL := startAggTestServer(t, dir, mock.server.URL, homeDir)
	defer cleanup()

	result := mcpCallAggregatedTool(t, baseURL, "agg_batch_greet", map[string]interface{}{
		"names": []interface{}{"Alice", "Bob", "Charlie"},
	})

	items := mustJSONArray(t, result)
	if len(items) != 3 {
		t.Fatalf("expected 3 results, got %d", len(items))
	}
	first := items[0].(map[string]interface{})
	if first["greeting"] != "Hi, Alice!" {
		t.Errorf("first greeting = %v", first["greeting"])
	}
	if _, ok := first["length"]; ok {
		t.Error("length should be projected out")
	}
	if len(mock.requests) != 3 {
		t.Errorf("expected 3 upstream requests, got %d", len(mock.requests))
	}
}

// ---------------------------------------------------------------------------
// E2E Test 4: Multiple aggregated tools coexist with native tools
// ---------------------------------------------------------------------------

func TestE2E_AggregatedTool_CoexistsWithNativeTools(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/echo") {
			w.Write([]byte(`{"status":"echo_ok"}`))
		} else if strings.Contains(r.URL.Path, "/hello") {
			name := r.URL.Query().Get("name")
			w.Write([]byte(fmt.Sprintf(`{"hello":"%s"}`, name)))
		} else {
			w.Write([]byte(`{}`))
		}
	})
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	aggConfig := `
aggregateTools:
  - name: agg_fast_echo
    description: Fast echo wrapper
    inputSchema:
      type: object
      properties: {}
    pipeline:
      - id: call_echo
        kind: call
        spec:
          tool: EchoHeaders
          args: {}
      - id: done
        kind: return
        spec:
          from: $call_echo

  - name: agg_greet_wrapper
    description: Greet with rename
    inputSchema:
      type: object
      properties:
        person:
          type: string
      required:
        - person
    pipeline:
      - id: call_hello
        kind: call
        spec:
          tool: SayHello
          args:
            name: $input.person
      - id: rename_greeting
        kind: jq
        spec:
          from: $call_hello
          expr: '{greeting: .hello}'
      - id: done
        kind: return
        spec:
          from: $rename_greeting
`
	writeAggregatedConfig(t, homeDir, binaryName, aggConfig)

	cleanup, baseURL := startAggTestServer(t, dir, mock.server.URL, homeDir)
	defer cleanup()

	result1 := mcpCallAggregatedTool(t, baseURL, "agg_fast_echo", map[string]interface{}{})
	if mustJSON(t, result1)["status"] != "echo_ok" {
		t.Error("agg_fast_echo failed")
	}

	result2 := mcpCallAggregatedTool(t, baseURL, "agg_greet_wrapper", map[string]interface{}{"person": "Zoe"})
	data2 := mustJSON(t, result2)
	if data2["greeting"] != "Zoe" {
		t.Errorf("greeting = %v", data2["greeting"])
	}
	if _, ok := data2["hello"]; ok {
		t.Error("hello should have been renamed to greeting")
	}

	if len(mock.requests) != 2 {
		t.Errorf("expected 2 upstream requests, got %d", len(mock.requests))
	}
}

// ---------------------------------------------------------------------------
// E2E Test 5: Invalid config — server still starts, native tools available
// ---------------------------------------------------------------------------

func TestE2E_AggregatedTool_InvalidConfigServerStarts(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	// Duplicate step ids — validation should fail
	aggConfig := `
aggregateTools:
  - name: bad_tool
    description: Broken
    inputSchema:
      type: object
      properties: {}
    pipeline:
      - id: step1
        kind: call
        spec:
          tool: EchoHeaders
          args: {}
      - id: step1
        kind: return
        spec:
          from: $step1
`
	writeAggregatedConfig(t, homeDir, binaryName, aggConfig)

	cleanup, baseURL := startAggTestServer(t, dir, mock.server.URL, homeDir)
	defer cleanup()

	// Native EchoHeaders tool should still be callable
	result := mcpCallAggregatedTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	if mustJSON(t, result)["status"] != "ok" {
		t.Error("native EchoHeaders should still work")
	}

	// The bad aggregated tool should not be registered
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "bad_tool",
		"arguments": map[string]interface{}{},
	})
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Tool not found") && !strings.Contains(string(body), "unknown") && !strings.Contains(string(body), "error") {
		t.Errorf("expected tool-not-found for bad aggregated tool, got: %s", string(body))
	}
}

// ---------------------------------------------------------------------------
// E2E Test 6: Call with parse: json — explicit JSON parse on string response
// ---------------------------------------------------------------------------

func TestE2E_AggregatedTool_CallParseJSON(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(`{"nested":{"key":"hello-world","count":42}}`))
	})
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	aggConfig := `
aggregateTools:
  - name: agg_parse_json
    description: Parse JSON from text response
    inputSchema:
      type: object
      properties: {}
    pipeline:
      - id: fetch
        kind: call
        spec:
          tool: EchoHeaders
          parse: json
          args: {}
      - id: done
        kind: return
        spec:
          from: $fetch
          expr: '{key: .nested.key, count: .nested.count}'
`
	writeAggregatedConfig(t, homeDir, binaryName, aggConfig)

	cleanup, baseURL := startAggTestServer(t, dir, mock.server.URL, homeDir)
	defer cleanup()

	result := mcpCallAggregatedTool(t, baseURL, "agg_parse_json", map[string]interface{}{})
	data := mustJSON(t, result)
	if data["key"] != "hello-world" {
		t.Errorf("key = %v, want hello-world", data["key"])
	}
	val, _ := data["count"].(float64)
	if val != 42 {
		t.Errorf("count = %v, want 42", val)
	}
}

// ---------------------------------------------------------------------------
// E2E Test 7: Require nonEmpty validation on jq step
// ---------------------------------------------------------------------------

func TestE2E_AggregatedTool_RequireNonEmptyOnJQ(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"items":[]}`))
	})
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	aggConfig := `
aggregateTools:
  - name: agg_require_jq
    description: Require nonEmpty on jq step
    inputSchema:
      type: object
      properties: {}
    pipeline:
      - id: fetch
        kind: call
        spec:
          tool: EchoHeaders
          args: {}
      - id: extract
        kind: jq
        spec:
          from: $fetch
          expr: '.items'
        require:
          nonEmpty: true
          message: "Items array must not be empty"
      - id: done
        kind: return
        spec:
          from: $extract
`
	writeAggregatedConfig(t, homeDir, binaryName, aggConfig)

	cleanup, baseURL := startAggTestServer(t, dir, mock.server.URL, homeDir)
	defer cleanup()

	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "agg_require_jq",
		"arguments": map[string]interface{}{},
	})
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Items array must not be empty") {
		t.Errorf("expected require validation error, got: %s", string(body))
	}
}

// ---------------------------------------------------------------------------
// E2E Test 8: Foreach with concurrency from input variable
// ---------------------------------------------------------------------------

func TestE2E_AggregatedTool_ForeachConcurrencyFromInput(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		name := r.URL.Query().Get("name")
		w.Write([]byte(fmt.Sprintf(`{"name":"%s","length":%d}`, name, len(name))))
	})
	defer mock.Close()

	dir := genProject(t, "sayHello", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	aggConfig := `
aggregateTools:
  - name: agg_concurrent
    description: Foreach with input concurrency
    inputSchema:
      type: object
      properties:
        names:
          type: array
          items:
            type: string
        workers:
          type: integer
      required:
        - names
    pipeline:
      - id: batch
        kind: foreach
        spec:
          in: $input.names
          as: name
          concurrency: $input.workers
          preserveOrder: true
          pipeline:
            - id: greet
              kind: call
              spec:
                tool: SayHello
                args:
                  name: $name
            - id: out
              kind: emit
              spec:
                from: $greet
                expr: '{name}'
      - id: done
        kind: return
        spec:
          from: $batch
`
	writeAggregatedConfig(t, homeDir, binaryName, aggConfig)

	cleanup, baseURL := startAggTestServer(t, dir, mock.server.URL, homeDir)
	defer cleanup()

	result := mcpCallAggregatedTool(t, baseURL, "agg_concurrent", map[string]interface{}{
		"names":   []interface{}{"Alice", "Bob", "Charlie", "Diana"},
		"workers": 2,
	})
	items := mustJSONArray(t, result)
	if len(items) != 4 {
		t.Fatalf("expected 4 results, got %d", len(items))
	}
	if len(mock.requests) != 4 {
		t.Errorf("expected 4 upstream calls, got %d", len(mock.requests))
	}
}

// ---------------------------------------------------------------------------
// E2E Test 9: Return with vars + complex expr building summary object
// ---------------------------------------------------------------------------

func TestE2E_AggregatedTool_ReturnWithVarsAndExpr(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/echo") {
			w.Write([]byte(`{"echo_status":"done","trace_id":"abc-123","server":"main"}`))
		} else if strings.Contains(r.URL.Path, "/hello") {
			name := r.URL.Query().Get("name")
			w.Write([]byte(fmt.Sprintf(`{"greeting":"Hello, %s!","code":200,"ts":"2024-01-01"}`, name)))
		} else {
			w.Write([]byte(`{}`))
		}
	})
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	aggConfig := `
aggregateTools:
  - name: agg_summary
    description: Build summary with vars and expr
    inputSchema:
      type: object
      properties:
        name:
          type: string
      required:
        - name
    pipeline:
      - id: echo
        kind: call
        spec:
          tool: EchoHeaders
          args: {}
      - id: greet
        kind: call
        spec:
          tool: SayHello
          args:
            name: $input.name
      - id: done
        kind: return
        spec:
          from: $echo
          vars:
            greeting: $greet.greeting
            code: $greet.code
          expr: '{summary: {origin: .server, greeting_text: $greeting, status_code: $code}, trace: .trace_id}'
`
	writeAggregatedConfig(t, homeDir, binaryName, aggConfig)

	cleanup, baseURL := startAggTestServer(t, dir, mock.server.URL, homeDir)
	defer cleanup()

	result := mcpCallAggregatedTool(t, baseURL, "agg_summary", map[string]interface{}{
		"name": "World",
	})
	data := mustJSON(t, result)
	summary, ok := data["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("summary field missing")
	}
	if summary["origin"] != "main" {
		t.Errorf("origin = %v", summary["origin"])
	}
	if summary["greeting_text"] != "Hello, World!" {
		t.Errorf("greeting_text = %v", summary["greeting_text"])
	}
	status, _ := summary["status_code"].(float64)
	if status != 200 {
		t.Errorf("status_code = %v", status)
	}
	if data["trace"] != "abc-123" {
		t.Errorf("trace = %v", data["trace"])
	}
	if len(mock.requests) != 2 {
		t.Errorf("expected 2 upstream requests, got %d", len(mock.requests))
	}
}

// ---------------------------------------------------------------------------
// E2E Test 10: Aggregated tool with annotations
// ---------------------------------------------------------------------------

func TestE2E_AggregatedTool_AnnotationsPropagated(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	aggConfig := `
aggregateTools:
  - name: agg_annotated
    description: Annotated tool
    annotations:
      readOnlyHint: true
      destructiveHint: false
      idempotentHint: true
    inputSchema:
      type: object
      properties: {}
    pipeline:
      - id: fetch
        kind: call
        spec:
          tool: EchoHeaders
          args: {}
      - id: done
        kind: return
        spec:
          from: $fetch
`
	writeAggregatedConfig(t, homeDir, binaryName, aggConfig)

	cleanup, baseURL := startAggTestServer(t, dir, mock.server.URL, homeDir)
	defer cleanup()

	// List tools and verify annotations are present
	resp, _ := mcpHTTPCall(t, baseURL, "tools/list", map[string]interface{}{})
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "readOnlyHint") {
		t.Errorf("expected readOnlyHint annotation in tools/list response, got: %s", string(body))
	}
	if !strings.Contains(string(body), "idempotentHint") {
		t.Errorf("expected idempotentHint annotation in tools/list response, got: %s", string(body))
	}

	// Also verify the tool is callable
	result := mcpCallAggregatedTool(t, baseURL, "agg_annotated", map[string]interface{}{})
	if mustJSON(t, result)["status"] != "ok" {
		t.Error("agg_annotated should still work")
	}
}

// ---------------------------------------------------------------------------
// E2E Test 11: Require nonEmpty — success case (non-empty result passes)
// ---------------------------------------------------------------------------

func TestE2E_AggregatedTool_RequireNonEmptyPasses(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"items":["a","b","c"]}}`))
	})
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	aggConfig := `
aggregateTools:
  - name: agg_require_ok
    description: Require validation that passes
    inputSchema:
      type: object
      properties: {}
    pipeline:
      - id: fetch
        kind: call
        spec:
          tool: EchoHeaders
          args: {}
      - id: extract_items
        kind: jq
        spec:
          from: $fetch
          expr: '.data.items'
        require:
          nonEmpty: true
          message: "Items must not be empty"
      - id: done
        kind: return
        spec:
          from: $extract_items
`
	writeAggregatedConfig(t, homeDir, binaryName, aggConfig)

	cleanup, baseURL := startAggTestServer(t, dir, mock.server.URL, homeDir)
	defer cleanup()

	result := mcpCallAggregatedTool(t, baseURL, "agg_require_ok", map[string]interface{}{})
	items := mustJSONArray(t, result)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
}
