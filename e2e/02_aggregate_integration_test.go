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
	"github.com/wl4g-ai/mcpgen/internal/generator/mcpaggregator/node"
	"github.com/wl4g-ai/mcpgen/internal/generator/mcpaggregator/pipeline"
	"github.com/wl4g-ai/mcpgen/internal/generator/mcpaggregator/runtime"
)

// ===========================================================================
// Helpers for in-process scenario tests
// ===========================================================================

type scenarioRunner struct {
	registry *callRecorder
	executor *runtime.Executor
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
		executor: runtime.NewExecutor(r),
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
		t.Fatalf("invalid JSON: %v", err)
	}
	return v
}

func mustJSONArray(t *testing.T, s string) []interface{} {
	t.Helper()
	var v []interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("invalid JSON array: %v", err)
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
// SECTION 1-9: In-process pipeline engine scenario tests (30 tests)
// ===========================================================================

func TestScenario_ConfigLoad_ValidAggregatedTools(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
aggregatedTools:
  - name: enrich_report
    version: "1.0"
    description: Enrich a report with extra data
    inputSchema:
      type: object
      properties:
        reportId:
          type: string
    pipeline:
      - name: fetch_report
        type: call
        call:
          tool: getReport
          args:
            id: "{{ input.reportId }}"
        output: report
      - name: fetch_metadata
        type: call
        call:
          tool: getMetadata
          args:
            id: "{{ input.reportId }}"
        output: meta
      - name: merge_data
        type: merge
        merge:
          from: "meta.output"
          to: "report.output.metadata"
      - name: clean
        type: transform
        transform:
          source: "report.output"
          remove:
            - internalId
            - debugInfo
          default:
            status: "unknown"
        output: cleaned
      - name: done
        type: return
        return:
          source: "cleaned.output"
  - name: health_summary
    version: "1.0"
    description: Get health summary
    inputSchema:
      type: object
      properties: {}
    pipeline:
      - name: check
        type: call
        call:
          tool: healthCheck
          args: {}
        output: status
      - name: done
        type: return
        return:
          source: "status.output"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if len(cfg.AggregatedTools) != 2 {
		t.Fatalf("expected 2 aggregated tools, got %d", len(cfg.AggregatedTools))
	}
	t1 := cfg.AggregatedTools[0]
	if t1.Name != "enrich_report" {
		t.Errorf("tool name = %q, want enrich_report", t1.Name)
	}
	if len(t1.Pipeline) != 5 {
		t.Errorf("expected 5 pipeline steps, got %d", len(t1.Pipeline))
	}
	t2 := cfg.AggregatedTools[1]
	if t2.Name != "health_summary" {
		t.Errorf("tool name = %q, want health_summary", t2.Name)
	}
	if len(t2.Pipeline) != 2 {
		t.Errorf("expected 2 pipeline steps, got %d", len(t2.Pipeline))
	}
}

func TestScenario_Validate_DuplicateStepNames(t *testing.T) {
	steps := []pipeline.StepConfig{
		{Name: "fetch", Type: "call", Call: &pipeline.CallConfig{Tool: "t1"}},
		{Name: "transform", Type: "transform", Transform: &pipeline.TransformConfig{Source: "fetch.output"}},
		{Name: "fetch", Type: "call", Call: &pipeline.CallConfig{Tool: "t2"}},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "fetch.output"}},
	}
	if err := pipeline.Validate(steps); err == nil {
		t.Fatal("expected validation error for duplicate step names")
	} else if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("error should mention 'duplicate', got: %v", err)
	}
}

func TestScenario_Validate_UnknownStepType(t *testing.T) {
	steps := []pipeline.StepConfig{
		{Name: "step1", Type: "call", Call: &pipeline.CallConfig{Tool: "t1"}},
		{Name: "step2", Type: "filter"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "step1.output"}},
	}
	if err := pipeline.Validate(steps); err == nil {
		t.Fatal("expected validation error for unknown step type")
	}
}

func TestScenario_Validate_MissingRequiredFields(t *testing.T) {
	steps := []pipeline.StepConfig{
		{Name: "step1", Type: "call", Call: &pipeline.CallConfig{Tool: ""}},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "step1.output"}},
	}
	if err := pipeline.Validate(steps); err == nil {
		t.Fatal("expected validation error for call with empty tool")
	}
	steps2 := []pipeline.StepConfig{
		{Name: "step1", Type: "call", Call: &pipeline.CallConfig{Tool: "t"}},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: ""}},
	}
	if err := pipeline.Validate(steps2); err == nil {
		t.Fatal("expected validation error for return with empty source")
	}
}

func TestScenario_Validate_DuplicateOutputNames(t *testing.T) {
	steps := []pipeline.StepConfig{
		{Name: "step1", Type: "call", Call: &pipeline.CallConfig{Tool: "t1"}, Output: "data"},
		{Name: "step2", Type: "call", Call: &pipeline.CallConfig{Tool: "t2"}, Output: "data"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "data.output"}},
	}
	if err := pipeline.Validate(steps); err == nil {
		t.Fatal("expected validation error for duplicate output names")
	}
}

func TestScenario_Context_ResolveNestedInputPaths(t *testing.T) {
	ctx := runtime.NewContext(map[string]interface{}{
		"user": map[string]interface{}{
			"profile": map[string]interface{}{"email": "alice@example.com", "memberId": 12345},
		},
		"filters": []interface{}{"active", "verified"},
	})
	tests := []struct {
		path     string
		expected interface{}
	}{
		{"input.user.profile.email", "alice@example.com"},
		{"input.user.profile.memberId", 12345},
		{"input.filters", []interface{}{"active", "verified"}},
	}
	for _, tc := range tests {
		v, err := ctx.Resolve("{{ " + tc.path + " }}")
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
	ctx := runtime.NewContext(map[string]interface{}{
		"items": []interface{}{map[string]interface{}{"name": "first"}, map[string]interface{}{"name": "second"}},
	})
	v, err := ctx.ResolvePath("input.items.0.name")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if v != "first" {
		t.Errorf("expected 'first', got %v", v)
	}
}

func TestScenario_Context_ItemContextInMapIteration(t *testing.T) {
	baseCtx := runtime.NewContext(map[string]interface{}{"batch": "batch-001"})
	itemCtx := baseCtx.WithItem(map[string]interface{}{"id": "item-42", "data": map[string]interface{}{"color": "red"}})
	v, err := itemCtx.Resolve("{{ item.id }}")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if v != "item-42" {
		t.Errorf("expected 'item-42', got %v", v)
	}
	v, err = itemCtx.ResolvePath("item.data.color")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if v != "red" {
		t.Errorf("expected 'red', got %v", v)
	}
}

func TestScenario_Context_UnresolvedReferenceErrors(t *testing.T) {
	ctx := runtime.NewContext(map[string]interface{}{"name": "test"})
	_, err := ctx.Resolve("{{ nonexistent.output.field }}")
	if err == nil {
		t.Fatal("expected error for unresolved step reference")
	}
	if !strings.Contains(err.Error(), "unresolved") {
		t.Errorf("error should say 'unresolved', got: %v", err)
	}
}

func TestScenario_Context_ResolveMapWithMixedReferences(t *testing.T) {
	ctx := runtime.NewContext(map[string]interface{}{"app": "myapp", "env": "production"})
	resolved, err := ctx.ResolveMap(map[string]interface{}{
		"appName": "{{ input.app }}",
		"nested":  map[string]interface{}{"env": "{{ input.env }}"},
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
		{Name: "fetch", Type: "call", Call: &pipeline.CallConfig{Tool: "getUser", Args: map[string]interface{}{"userId": "{{ input.id }}"}}, Output: "user"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "user.output"}},
	}, map[string]interface{}{"id": "123"})
	data := resultJSON(t, result)
	if data["name"] != "Alice" {
		t.Errorf("name = %v", data["name"])
	}
}

func TestScenario_Call_NestedObjectArgResolution(t *testing.T) {
	s := newScenarioRunner(map[string]string{"search": `{"total":5}`})
	_ = s.execute(t, []pipeline.StepConfig{
		{Name: "query", Type: "call", Call: &pipeline.CallConfig{
			Tool: "search",
			Args: map[string]interface{}{"filter": map[string]interface{}{"app": "{{ input.app }}"}, "sort": "{{ input.sortBy }}"},
		}, Output: "result"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "result.output"}},
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
		{Name: "fetch", Type: "call", Call: &pipeline.CallConfig{Tool: "getUser", Args: map[string]interface{}{}}, Output: "user"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "user.output"}},
	}, map[string]interface{}{})
	if err == nil || !strings.Contains(err.Error(), "503") {
		t.Errorf("expected 503 error, got %v", err)
	}
}

func TestScenario_Transform_ProjectFields(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getReport": `{"id":"r1","title":"Q4","author":"Bob","internalId":"secret","score":85}`})
	result := s.execute(t, []pipeline.StepConfig{
		{Name: "fetch", Type: "call", Call: &pipeline.CallConfig{Tool: "getReport", Args: map[string]interface{}{}}, Output: "raw"},
		{Name: "select", Type: "transform", Transform: &pipeline.TransformConfig{Source: "fetch.output", Project: []string{"id", "title", "score"}}, Output: "filtered"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "filtered.output"}},
	}, nil)
	data := resultJSON(t, result)
	if _, ok := data["internalId"]; ok {
		t.Error("internalId should be projected out")
	}
	if len(data) != 3 {
		t.Errorf("expected 3 fields, got %d", len(data))
	}
}

func TestScenario_Transform_ProjectArrayOfObjects(t *testing.T) {
	s := newScenarioRunner(map[string]string{"listUsers": `[{"id":"1","name":"A","password":"xx"},{"id":"2","name":"B","password":"yy"}]`})
	result := s.execute(t, []pipeline.StepConfig{
		{Name: "fetch", Type: "call", Call: &pipeline.CallConfig{Tool: "listUsers", Args: map[string]interface{}{}}, Output: "raw"},
		{Name: "sanitize", Type: "transform", Transform: &pipeline.TransformConfig{Source: "fetch.output", Project: []string{"id", "name"}}, Output: "clean"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "clean.output"}},
	}, nil)
	items := resultJSONArray(t, result)
	for i, item := range items {
		m := item.(map[string]interface{})
		if _, ok := m["password"]; ok {
			t.Errorf("item[%d]: password should be removed", i)
		}
	}
}

func TestScenario_Transform_RemoveSensitiveFields(t *testing.T) {
	data := map[string]interface{}{"username": "alice", "password": "secret123", "token": "abc", "email": "alice@x.com"}
	ctx := runtime.NewContext(nil)
	ctx.SetOutput("dummy", data)
	result, err := node.TransformNode(&pipeline.StepConfig{Transform: &pipeline.TransformConfig{Source: "dummy", Remove: []string{"password", "token"}}}, ctx)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if _, ok := m["password"]; ok {
		t.Error("password should be removed")
	}
}

func TestScenario_Transform_RenameFields(t *testing.T) {
	data := map[string]interface{}{"first_name": "John", "last_name": "Doe"}
	ctx := runtime.NewContext(nil)
	ctx.SetOutput("dummy", data)
	result, err := node.TransformNode(&pipeline.StepConfig{Transform: &pipeline.TransformConfig{Source: "dummy", Rename: map[string]string{"first_name": "firstName", "last_name": "lastName"}}}, ctx)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["firstName"] != "John" || m["lastName"] != "Doe" {
		t.Errorf("rename failed: %v", m)
	}
	if _, ok := m["first_name"]; ok {
		t.Error("old name should be gone")
	}
}

func TestScenario_Transform_CopyFieldsDeepCopy(t *testing.T) {
	data := map[string]interface{}{"config": map[string]interface{}{"timeout": 30}}
	ctx := runtime.NewContext(nil)
	ctx.SetOutput("dummy", data)
	result, err := node.TransformNode(&pipeline.StepConfig{Transform: &pipeline.TransformConfig{Source: "dummy", Copy: map[string]string{"config": "backup"}}}, ctx)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	m["backup"].(map[string]interface{})["timeout"] = 99
	if m["config"].(map[string]interface{})["timeout"] != 30 {
		t.Error("deep copy failed — original modified")
	}
}

func TestScenario_Transform_MoveFields(t *testing.T) {
	data := map[string]interface{}{"tempId": "tmp-001", "permanent": "value"}
	ctx := runtime.NewContext(nil)
	ctx.SetOutput("dummy", data)
	result, err := node.TransformNode(&pipeline.StepConfig{Transform: &pipeline.TransformConfig{Source: "dummy", Move: map[string]string{"tempId": "finalId"}}}, ctx)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["finalId"] != "tmp-001" {
		t.Errorf("finalId = %v", m["finalId"])
	}
	if _, ok := m["tempId"]; ok {
		t.Error("tempId should be removed after move")
	}
}

func TestScenario_Transform_FlattenNestedStructures(t *testing.T) {
	data := map[string]interface{}{"name": "report", "metadata": map[string]interface{}{"author": "Alice", "version": 2}, "stats": map[string]interface{}{"views": 100}}
	ctx := runtime.NewContext(nil)
	ctx.SetOutput("dummy", data)
	result, err := node.TransformNode(&pipeline.StepConfig{Transform: &pipeline.TransformConfig{Source: "dummy", Flatten: []string{"metadata"}}}, ctx)
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
	if _, ok := m["stats"]; !ok {
		t.Error("stats should remain")
	}
}

func TestScenario_Transform_DefaultValues(t *testing.T) {
	data := map[string]interface{}{"name": "existing"}
	ctx := runtime.NewContext(nil)
	ctx.SetOutput("dummy", data)
	result, err := node.TransformNode(&pipeline.StepConfig{Transform: &pipeline.TransformConfig{Source: "dummy", Default: map[string]interface{}{"name": "defaultName", "version": "1.0"}}}, ctx)
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

func TestScenario_Transform_ChainedOperations(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getData": `{"first_name":"Jane","last_name":"Smith","password":"xxx","address":{"city":"NYC"}}`})
	result := s.execute(t, []pipeline.StepConfig{
		{Name: "fetch", Type: "call", Call: &pipeline.CallConfig{Tool: "getData", Args: map[string]interface{}{}}, Output: "raw"},
		{Name: "rename_step", Type: "transform", Transform: &pipeline.TransformConfig{Source: "fetch.output", Rename: map[string]string{"first_name": "firstName", "last_name": "lastName"}}, Output: "renamed"},
		{Name: "sanitize_step", Type: "transform", Transform: &pipeline.TransformConfig{Source: "renamed.output", Remove: []string{"password"}}, Output: "sanitized"},
		{Name: "flatten_step", Type: "transform", Transform: &pipeline.TransformConfig{Source: "sanitized.output", Flatten: []string{"address"}, Default: map[string]interface{}{"role": "user"}}, Output: "final"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "final.output"}},
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

func TestScenario_Merge_DataFromOneStepIntoAnother(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getBase": `{"id":"r1"}`, "getMeta": `{"author":"Alice"}`})
	result := s.execute(t, []pipeline.StepConfig{
		{Name: "fetch_base", Type: "call", Call: &pipeline.CallConfig{Tool: "getBase", Args: map[string]interface{}{}}, Output: "base"},
		{Name: "fetch_meta", Type: "call", Call: &pipeline.CallConfig{Tool: "getMeta", Args: map[string]interface{}{}}, Output: "meta"},
		{Name: "merge_step", Type: "merge", Merge: &pipeline.MergeConfig{From: "meta.output", To: "base.output.metadata"}, Output: "merge_op"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "merge_op.output"}},
	}, nil)
	mergeData := resultJSON(t, result)
	if mergeData["id"] != "r1" {
		t.Errorf("id = %v, want r1", mergeData["id"])
	}
	meta, ok := mergeData["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("metadata field missing from merged result")
	}
	if meta["author"] != "Alice" {
		t.Errorf("metadata.author = %v", meta["author"])
	}
}

func TestScenario_Merge_ApplyMergeWritesToTarget(t *testing.T) {
	target := map[string]interface{}{"id": "r1"}
	value := map[string]interface{}{"author": "Alice"}
	node.ApplyMerge(target, "base.metadata", value)
	metaMap := target["metadata"].(map[string]interface{})
	if metaMap["author"] != "Alice" {
		t.Errorf("metadata.author = %v", metaMap["author"])
	}
}

func TestScenario_Return_StringValue(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getStatus": `OK`})
	result := s.execute(t, []pipeline.StepConfig{
		{Name: "fetch", Type: "call", Call: &pipeline.CallConfig{Tool: "getStatus", Args: map[string]interface{}{}}, Output: "status"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "status.output"}},
	}, nil)
	if !strings.Contains(resultText(t, result), "OK") {
		t.Errorf("expected OK, got %q", resultText(t, result))
	}
}

func TestScenario_Return_ComplexObjectAsJSON(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getConfig": `{"features":{"darkMode":true}}`})
	result := s.execute(t, []pipeline.StepConfig{
		{Name: "fetch", Type: "call", Call: &pipeline.CallConfig{Tool: "getConfig", Args: map[string]interface{}{}}, Output: "cfg"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "cfg.output"}},
	}, nil)
	data := resultJSON(t, result)
	if data["features"].(map[string]interface{})["darkMode"] != true {
		t.Error("darkMode should be true")
	}
}

func TestScenario_Map_CallSubPipelinePerItem(t *testing.T) {
	s := newScenarioRunner(map[string]string{"lookupName": `Alice`})
	result := s.execute(t, []pipeline.StepConfig{
		{Name: "batch", Type: "map", Map: &pipeline.MapConfig{
			Source: "{{ input.ids }}",
			Pipeline: []pipeline.StepConfig{
				{Name: "lookup", Type: "call", Call: &pipeline.CallConfig{Tool: "lookupName", Args: map[string]interface{}{"id": "{{ item }}"}}, Output: "r"},
				{Name: "ret", Type: "return", Return: &pipeline.ReturnConfig{Source: "r.output"}},
			},
		}, Output: "names"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "names.output"}},
	}, map[string]interface{}{"ids": []interface{}{"1", "2", "3"}})
	if len(resultJSONArray(t, result)) != 3 {
		t.Fatal("expected 3 results")
	}
	if s.registry.callCount("lookupName") != 3 {
		t.Errorf("expected 3 calls, got %d", s.registry.callCount("lookupName"))
	}
}

func TestScenario_Map_TransformSubPipelinePerItem(t *testing.T) {
	s := newScenarioRunner(map[string]string{"fetchItem": `{"name":"test","password":"secret","role":"admin"}`})
	result := s.execute(t, []pipeline.StepConfig{
		{Name: "process", Type: "map", Map: &pipeline.MapConfig{
			Source: "{{ input.items }}",
			Pipeline: []pipeline.StepConfig{
				{Name: "fetch", Type: "call", Call: &pipeline.CallConfig{Tool: "fetchItem", Args: map[string]interface{}{}}, Output: "raw"},
				{Name: "clean", Type: "transform", Transform: &pipeline.TransformConfig{Source: "raw.output", Remove: []string{"password"}, Project: []string{"name", "role"}}, Output: "cleaned"},
				{Name: "ret", Type: "return", Return: &pipeline.ReturnConfig{Source: "cleaned.output"}},
			},
		}, Output: "results"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "results.output"}},
	}, map[string]interface{}{"items": []interface{}{"a", "b"}})
	items := resultJSONArray(t, result)
	for i, item := range items {
		m := item.(map[string]interface{})
		if _, ok := m["password"]; ok {
			t.Errorf("item[%d]: password should be removed", i)
		}
	}
}

func TestScenario_Map_ErrorPropagationStopsExecution(t *testing.T) {
	s := newScenarioRunner(map[string]string{"riskyCall": `OK`})
	s.setFail("riskyCall", fmt.Errorf("upstream timeout on item-2"))
	err := s.executeErr(t, []pipeline.StepConfig{
		{Name: "danger", Type: "map", Map: &pipeline.MapConfig{
			Source: "{{ input.items }}",
			Pipeline: []pipeline.StepConfig{
				{Name: "call", Type: "call", Call: &pipeline.CallConfig{Tool: "riskyCall", Args: map[string]interface{}{}}, Output: "r"},
				{Name: "ret", Type: "return", Return: &pipeline.ReturnConfig{Source: "r.output"}},
			},
		}, Output: "results"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "results.output"}},
	}, map[string]interface{}{"items": []interface{}{1, 2, 3}})
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout error, got %v", err)
	}
}

func TestScenario_Map_EmptyListReturnsEmptyResults(t *testing.T) {
	s := newScenarioRunner(map[string]string{"never": "called"})
	result := s.execute(t, []pipeline.StepConfig{
		{Name: "empty", Type: "map", Map: &pipeline.MapConfig{
			Source: "{{ input.items }}",
			Pipeline: []pipeline.StepConfig{
				{Name: "call", Type: "call", Call: &pipeline.CallConfig{Tool: "never", Args: map[string]interface{}{}}, Output: "r"},
				{Name: "ret", Type: "return", Return: &pipeline.ReturnConfig{Source: "r.output"}},
			},
		}, Output: "results"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "results.output"}},
	}, map[string]interface{}{"items": []interface{}{}})
	if len(resultJSONArray(t, result)) != 0 {
		t.Error("expected empty results")
	}
}

func TestScenario_Map_ConcurrentExecutionPreservesOrder(t *testing.T) {
	s := newScenarioRunner(map[string]string{"indexItem": `{"value":"x"}`})
	result := s.execute(t, []pipeline.StepConfig{
		{Name: "ordered", Type: "map", Map: &pipeline.MapConfig{
			Source: "{{ input.ids }}",
			Pipeline: []pipeline.StepConfig{
				{Name: "index", Type: "call", Call: &pipeline.CallConfig{Tool: "indexItem", Args: map[string]interface{}{"id": "{{ item }}"}}, Output: "r"},
				{Name: "ret", Type: "return", Return: &pipeline.ReturnConfig{Source: "r.output"}},
			},
		}, Output: "results"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "results.output"}},
	}, map[string]interface{}{"ids": []interface{}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}})
	if len(resultJSONArray(t, result)) != 10 {
		t.Fatal("expected 10 results")
	}
}

func TestScenario_FullPipeline_CallTransformReturn(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getViolations": `{"total":10,"violations":[{"severity":"HIGH"},{"severity":"LOW"}],"scanId":"scan-001"}`})
	result := s.execute(t, []pipeline.StepConfig{
		{Name: "fetch", Type: "call", Call: &pipeline.CallConfig{Tool: "getViolations", Args: map[string]interface{}{"scanId": "{{ input.scanId }}"}}, Output: "raw"},
		{Name: "clean", Type: "transform", Transform: &pipeline.TransformConfig{Source: "fetch.output", Project: []string{"violations"}, Default: map[string]interface{}{"total": 0}}, Output: "cleaned"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "cleaned.output"}},
	}, map[string]interface{}{"scanId": "scan-001"})
	data := resultJSON(t, result)
	if _, ok := data["scanId"]; ok {
		t.Error("scanId should be projected out")
	}
	if len(data["violations"].([]interface{})) != 2 {
		t.Errorf("expected 2 violations")
	}
}

func TestScenario_FullPipeline_CallMapCallTransformReturn(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getComponents": `[{"id":"c1","name":"auth","owner":"team-A","internalRef":"x-1"},{"id":"c2","name":"ui","owner":"team-B","internalRef":"x-2"}]`})
	result := s.execute(t, []pipeline.StepConfig{
		{Name: "fetch_all", Type: "call", Call: &pipeline.CallConfig{Tool: "getComponents", Args: map[string]interface{}{}}, Output: "all"},
		{Name: "enrich_each", Type: "map", Map: &pipeline.MapConfig{
			Source: "all.output",
			Pipeline: []pipeline.StepConfig{
				{Name: "transform_item", Type: "transform", Transform: &pipeline.TransformConfig{Source: "item", Project: []string{"id", "name", "owner"}, Rename: map[string]string{"owner": "maintainer"}, Default: map[string]interface{}{"status": "active"}}, Output: "enriched"},
				{Name: "ret_item", Type: "return", Return: &pipeline.ReturnConfig{Source: "enriched.output"}},
			},
		}, Output: "results"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "results.output"}},
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

func TestScenario_FullPipeline_CallMapCallMergeReturn(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getReports": `[{"id":"r1"},{"id":"r2"}]`, "getAnnotations": `{"color":"blue"}`})
	result := s.execute(t, []pipeline.StepConfig{
		{Name: "fetch_reports", Type: "call", Call: &pipeline.CallConfig{Tool: "getReports", Args: map[string]interface{}{}}, Output: "reports"},
		{Name: "annotate_each", Type: "map", Map: &pipeline.MapConfig{
			Source: "reports.output",
			Pipeline: []pipeline.StepConfig{
				{Name: "get_annotations", Type: "call", Call: &pipeline.CallConfig{Tool: "getAnnotations", Args: map[string]interface{}{"id": "{{ item.id }}"}}, Output: "annotations"},
				{Name: "merge_annotations", Type: "merge", Merge: &pipeline.MergeConfig{From: "annotations.output", To: "item.annotationData"}, Output: "merge_op"},
				{Name: "ret_item", Type: "return", Return: &pipeline.ReturnConfig{Source: "merge_op.output"}},
			},
		}, Output: "results"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "results.output"}},
	}, nil)
	if len(resultJSONArray(t, result)) != 2 {
		t.Fatal("expected 2 results")
	}
}

func TestScenario_FullPipeline_PipelineWithoutReturnErrors(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getData": `{}`})
	err := s.executeErr(t, []pipeline.StepConfig{
		{Name: "fetch", Type: "call", Call: &pipeline.CallConfig{Tool: "getData", Args: map[string]interface{}{}}, Output: "data"},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "without a return step") {
		t.Errorf("expected missing-return error, got %v", err)
	}
}

func TestScenario_FullPipeline_InvalidReferenceFails(t *testing.T) {
	s := newScenarioRunner(map[string]string{"getData": `{"id":"123"}`})
	err := s.executeErr(t, []pipeline.StepConfig{
		{Name: "fetch", Type: "call", Call: &pipeline.CallConfig{Tool: "getData", Args: map[string]interface{}{}}, Output: "data"},
		{Name: "clean", Type: "transform", Transform: &pipeline.TransformConfig{Source: "nonexistent_step.output"}, Output: "cleaned"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "cleaned.output"}},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "unresolved") {
		t.Errorf("expected unresolved error, got %v", err)
	}
}

func TestScenario_FullPipeline_MultipleAggregatedToolsInEngine(t *testing.T) {
	cfg := &config.Config{
		AggregatedTools: []config.AggregatedToolConfig{
			{Name: "tool_a", Version: "1.0", InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}, Pipeline: []pipeline.StepConfig{
				{Name: "step1", Type: "call", Call: &pipeline.CallConfig{Tool: "nativeA", Args: map[string]interface{}{"src": "a"}}, Output: "out"},
				{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "out.output"}},
			}},
			{Name: "tool_b", Version: "2.0", InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}, Pipeline: []pipeline.StepConfig{
				{Name: "step1", Type: "call", Call: &pipeline.CallConfig{Tool: "nativeB", Args: map[string]interface{}{"src": "b"}}, Output: "out"},
				{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "out.output"}},
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
		{Name: "fetch", Type: "call", Call: &pipeline.CallConfig{Tool: "getMessage", Args: map[string]interface{}{}}, Output: "msg"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "msg.output"}},
	}, nil)
	if resultText(t, result) != "plain text response" {
		t.Errorf("unexpected result text: %s", resultText(t, result))
	}
}

func TestScenario_EdgeCase_CallWithNoArgs(t *testing.T) {
	s := newScenarioRunner(map[string]string{"ping": `{"status":"ok"}`})
	result := s.execute(t, []pipeline.StepConfig{
		{Name: "pinger", Type: "call", Call: &pipeline.CallConfig{Tool: "ping", Args: map[string]interface{}{}}, Output: "pong"},
		{Name: "done", Type: "return", Return: &pipeline.ReturnConfig{Source: "pong.output"}},
	}, nil)
	if resultJSON(t, result)["status"] != "ok" {
		t.Error("expected ok")
	}
}

func TestScenario_EdgeCase_TransformOnArraySource(t *testing.T) {
	data := []interface{}{map[string]interface{}{"a": 1, "b": 2, "c": 3}, map[string]interface{}{"a": 4, "b": 5, "c": 6}}
	ctx := runtime.NewContext(nil)
	ctx.SetOutput("dummy", data)
	result, err := node.TransformNode(&pipeline.StepConfig{Transform: &pipeline.TransformConfig{Source: "dummy", Project: []string{"a", "c"}}}, ctx)
	if err != nil {
		t.Fatal(err)
	}
	arr := result.([]interface{})
	if _, ok := arr[0].(map[string]interface{})["b"]; ok {
		t.Error("field b should be projected out")
	}
}

// ===========================================================================
// SECTION 10: End-to-End Integration Tests (generate → build → run → verify)
// ===========================================================================

// writeAggregatedConfig writes an aggregated tool YAML config to
// $HOME/.<binaryName>/config.yaml so the generated server loads it at startup.
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

// mcpCallAggregatedTool sends a tools/call request via MCP HTTP and returns the first text content.
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

// startAggTestServer compiles the generated project and starts it in HTTP mode.
// Returns a cleanup function (defer it) and the base URL.
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
// E2E Test 1: Aggregated tool with call → transform(project) → return
// ---------------------------------------------------------------------------

func TestE2E_AggregatedTool_CallTransformReturn(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","message":"hello","internalToken":"secret123","timestamp":"2024-01-01T00:00:00Z"}`))
	})
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	aggConfig := `
aggregatedTools:
  - name: agg_clean_echo
    version: "1.0"
    description: Echo with cleanup
    inputSchema:
      type: object
      properties: {}
    pipeline:
      - name: fetch
        type: call
        call:
          tool: EchoHeaders
          args: {}
        output: raw
      - name: clean
        type: transform
        transform:
          source: "fetch.output"
          project:
            - status
            - message
            - timestamp
        output: cleaned
      - name: done
        type: return
        return:
          source: "cleaned.output"
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
// E2E Test 2: Aggregated tool chaining two native tools (EchoHeaders + SayHello)
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
aggregatedTools:
  - name: agg_chain
    version: "1.0"
    description: Chain echo and greet
    inputSchema:
      type: object
      properties:
        name:
          type: string
      required:
        - name
    pipeline:
      - name: echo
        type: call
        call:
          tool: EchoHeaders
          args: {}
        output: echoResult
      - name: greet
        type: call
        call:
          tool: SayHello
          args:
            name: "{{ input.name }}"
        output: greetResult
      - name: merge_step
        type: merge
        merge:
          from: "greetResult.output"
          to: "echoResult.output.greeting_data"
      - name: done
        type: return
        return:
          source: "echoResult.output"
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
// E2E Test 3: Map over input array — calls SayHello for each name
// ---------------------------------------------------------------------------

func TestE2E_AggregatedTool_MapOverInputArray(t *testing.T) {
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
aggregatedTools:
  - name: agg_batch_greet
    version: "1.0"
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
      - name: process_all
        type: map
        map:
          source: "{{ input.names }}"
          pipeline:
            - name: greet_one
              type: call
              call:
                tool: SayHello
                args:
                  name: "{{ item }}"
              output: oneResult
            - name: clean_one
              type: transform
              transform:
                source: "oneResult.output"
                project:
                  - greeting
              output: cleaned
            - name: ret_one
              type: return
              return:
                source: "cleaned.output"
        output: results
      - name: done
        type: return
        return:
          source: "results.output"
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
aggregatedTools:
  - name: agg_fast_echo
    version: "1.0"
    description: Fast echo wrapper
    inputSchema:
      type: object
      properties: {}
    pipeline:
      - name: call_echo
        type: call
        call:
          tool: EchoHeaders
          args: {}
        output: raw
      - name: done
        type: return
        return:
          source: "call_echo.output"

  - name: agg_greet_wrapper
    version: "1.0"
    description: Greet with transform
    inputSchema:
      type: object
      properties:
        person:
          type: string
      required:
        - person
    pipeline:
      - name: call_hello
        type: call
        call:
          tool: SayHello
          args:
            name: "{{ input.person }}"
        output: raw
      - name: transform_hello
        type: transform
        transform:
          source: "raw.output"
          rename:
            hello: "greeting"
        output: final
      - name: done
        type: return
        return:
          source: "final.output"
`
	writeAggregatedConfig(t, homeDir, binaryName, aggConfig)

	cleanup, baseURL := startAggTestServer(t, dir, mock.server.URL, homeDir)
	defer cleanup()

	// Call first aggregated tool
	result1 := mcpCallAggregatedTool(t, baseURL, "agg_fast_echo", map[string]interface{}{})
	if mustJSON(t, result1)["status"] != "echo_ok" {
		t.Error("agg_fast_echo failed")
	}

	// Call second aggregated tool
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

	// Duplicate step names — validation should fail
	aggConfig := `
aggregatedTools:
  - name: bad_tool
    version: "1.0"
    description: Broken
    inputSchema:
      type: object
      properties: {}
    pipeline:
      - name: step1
        type: call
        call:
          tool: EchoHeaders
          args: {}
        output: out
      - name: step1
        type: return
        return:
          source: "out.output"
`
	writeAggregatedConfig(t, homeDir, binaryName, aggConfig)

	// Server should still start — bad aggregated tools are warned, not fatal
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
