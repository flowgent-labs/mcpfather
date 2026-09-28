package node

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flowgent-labs/mcpfather/pkg/generator/mcpvirtual/pipeline"
)

// HTTPNode executes an http step — calling an external HTTP API on a named upstream.
func HTTPNode(ctx context.Context, step *pipeline.StepConfig, rctx pipeline.StepContext, client pipeline.HTTPClient) (interface{}, error) {
	spec := step.Spec

	resolvedPath, err := resolveHTTPString(spec.Path, rctx)
	if err != nil {
		return nil, fmt.Errorf("http path: %w", err)
	}

	resolvedQuery, err := resolveHTTPStringMap(spec.Query, rctx)
	if err != nil {
		return nil, fmt.Errorf("http query: %w", err)
	}

	resolvedHeaders, err := resolveHTTPStringMap(spec.Headers, rctx)
	if err != nil {
		return nil, fmt.Errorf("http headers: %w", err)
	}

	resolvedBody, err := resolveHTTPBody(spec.Body, rctx)
	if err != nil {
		return nil, fmt.Errorf("http body: %w", err)
	}

	resp, err := client.Call(ctx, spec.Upstream, spec.Method, resolvedPath, resolvedQuery, resolvedHeaders, resolvedBody)
	if err != nil {
		return nil, fmt.Errorf("http %s %q on upstream %q failed: %w", spec.Method, spec.Path, spec.Upstream, err)
	}
	if resp == nil {
		return nil, fmt.Errorf("http %s %q on upstream %q returned an empty response", spec.Method, spec.Path, spec.Upstream)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %s %q returned %d", spec.Method, spec.Path, resp.StatusCode)
	}

	parsedBody, err := parseHTTPResponseBody(resp.Body, spec.Parse)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"body":    parsedBody,
		"headers": normalizeHTTPResponseHeaders(resp.Headers),
	}, nil
}

func parseHTTPResponseBody(respBody []byte, parse string) (interface{}, error) {
	if parse == "json" {
		var parsed interface{}
		if err := json.Unmarshal(respBody, &parsed); err != nil {
			return nil, fmt.Errorf("failed to parse http response as JSON: %w", err)
		}
		return parsed, nil
	}

	// Auto-detect JSON
	var parsed interface{}
	if err := json.Unmarshal(respBody, &parsed); err == nil {
		return parsed, nil
	}
	return string(respBody), nil
}

func normalizeHTTPResponseHeaders(headers map[string][]string) map[string]interface{} {
	normalized := make(map[string]interface{}, len(headers))
	for name, values := range headers {
		switch len(values) {
		case 0:
			normalized[name] = ""
		case 1:
			normalized[name] = values[0]
		default:
			items := make([]interface{}, len(values))
			for i, value := range values {
				items[i] = value
			}
			normalized[name] = items
		}
	}
	return normalized
}

func resolveHTTPString(s string, rctx pipeline.StepContext) (string, error) {
	resolved, err := rctx.Resolve(s)
	if err != nil {
		return "", err
	}
	s, ok := resolved.(string)
	if !ok {
		data, _ := json.Marshal(resolved)
		return string(data), nil
	}
	return s, nil
}

func resolveHTTPStringMap(m map[string]interface{}, rctx pipeline.StepContext) (map[string]string, error) {
	if len(m) == 0 {
		return nil, nil
	}
	resolved, err := rctx.ResolveMap(m)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(resolved))
	for k, v := range resolved {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out, nil
}

func resolveHTTPBody(body interface{}, rctx pipeline.StepContext) (interface{}, error) {
	if body == nil {
		return nil, nil
	}
	switch v := body.(type) {
	case string:
		return rctx.Resolve(v)
	case map[string]interface{}:
		return rctx.ResolveMap(v)
	default:
		return v, nil
	}
}
