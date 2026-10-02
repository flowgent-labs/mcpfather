package node

import (
	"context"
	"reflect"
	"testing"

	"github.com/flowgent-labs/mcpfather/pkg/generator/mcpvirtual/pipeline"
)

type stubHTTPClient struct {
	response *pipeline.HTTPResponse
	pool     string
	retry    string
}

func (c *stubHTTPClient) Call(_ context.Context, _, pool, retry, _, _ string, _ map[string]string, _ map[string]string, _ interface{}) (*pipeline.HTTPResponse, error) {
	c.pool = pool
	c.retry = retry
	return c.response, nil
}

func TestHTTPNode_WrapsBodyAndSingleValueHeaders(t *testing.T) {
	client := &stubHTTPClient{response: &pipeline.HTTPResponse{
		StatusCode: 200,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
			"Set-Cookie":   {"CLM-CSRF-TOKEN=csrf-123"},
		},
		Body: []byte(`{"status":"ok"}`),
	}}
	step := &pipeline.StepConfig{
		Kind: "http",
		Spec: pipeline.StepSpec{Upstream: "iq", Pool: "iq-pool", Retry: "iq-read", Method: "GET", Path: "/assets/index.html", Parse: "json"},
	}

	got, err := HTTPNode(context.Background(), step, newMockCtx(nil), client)
	if err != nil {
		t.Fatalf("HTTPNode returned an error: %v", err)
	}
	want := map[string]interface{}{
		"body": map[string]interface{}{"status": "ok"},
		"headers": map[string]interface{}{
			"Content-Type": "application/json",
			"Set-Cookie":   "CLM-CSRF-TOKEN=csrf-123",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("wrapped HTTP output mismatch:\n got: %#v\nwant: %#v", got, want)
	}
	if client.pool != "iq-pool" || client.retry != "iq-read" {
		t.Fatalf("middleware references = pool:%q retry:%q", client.pool, client.retry)
	}
}

func TestHTTPNode_NormalizesMultiValueResponseHeaders(t *testing.T) {
	client := &stubHTTPClient{response: &pipeline.HTTPResponse{
		StatusCode: 200,
		Headers: map[string][]string{
			"Set-Cookie": {"CLM-CSRF-TOKEN=csrf-123"},
			"Vary":       {"Accept-Encoding", "Origin"},
		},
		Body: []byte(`{"status":"ok"}`),
	}}
	step := &pipeline.StepConfig{
		Kind: "http",
		Spec: pipeline.StepSpec{
			Upstream: "iq",
			Method:   "GET",
			Path:     "/assets/index.html",
			Parse:    "json",
		},
	}

	got, err := HTTPNode(context.Background(), step, newMockCtx(nil), client)
	if err != nil {
		t.Fatalf("HTTPNode returned an error: %v", err)
	}
	want := map[string]interface{}{
		"body": map[string]interface{}{"status": "ok"},
		"headers": map[string]interface{}{
			"Set-Cookie": "CLM-CSRF-TOKEN=csrf-123",
			"Vary":       []interface{}{"Accept-Encoding", "Origin"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("wrapped HTTP output mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}
