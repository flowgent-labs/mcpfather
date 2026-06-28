package node

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"jira-mcp-v10.7.4/internal/mcpaggregator/pipeline"
)

const defaultConcurrency = 4

// ForeachNode executes a foreach step — iterating over a list and running a
// sub-pipeline per item with configurable concurrency.
func ForeachNode(ctx context.Context, step *pipeline.StepConfig, rctx pipeline.StepContext, executor pipeline.StepExecutor) (interface{}, error) {
	spec := step.Spec

	source, err := rctx.ResolvePath(spec.In)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve foreach in: %w", err)
	}

	list, ok := source.([]interface{})
	if !ok {
		return nil, fmt.Errorf("foreach in must resolve to an array, got %T", source)
	}

	concurrency := resolveConcurrency(spec.Concurrency, rctx)
	if concurrency < 1 {
		concurrency = 1
	}

	itemName := spec.As
	if itemName == "" {
		itemName = "item"
	}

	if spec.PreserveOrder {
		return runForeachOrdered(ctx, list, itemName, concurrency, spec.Pipeline, rctx, executor)
	}
	return runForeachUnordered(ctx, list, itemName, concurrency, spec.Pipeline, rctx, executor)
}

func resolveConcurrency(raw interface{}, rctx pipeline.StepContext) int {
	if raw == nil {
		return defaultConcurrency
	}
	switch v := raw.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if strings.HasPrefix(v, "$") {
			if resolved, err := rctx.Resolve(v); err == nil {
				return resolveConcurrency(resolved, rctx)
			}
		}
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultConcurrency
}

func runForeachOrdered(ctx context.Context, list []interface{}, itemName string, concurrency int, pipeline []pipeline.StepConfig, rctx pipeline.StepContext, executor pipeline.StepExecutor) ([]interface{}, error) {
	results := make([]interface{}, len(list))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	var firstErr error

	for i, item := range list {
		if firstErr != nil {
			break
		}
		wg.Add(1)
		go func(idx int, item interface{}) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				mu.Lock()
				if firstErr == nil {
					firstErr = ctx.Err()
				}
				mu.Unlock()
				return
			}

			val, err := runSubPipeline(ctx, idx, item, itemName, pipeline, rctx, executor)
			mu.Lock()
			if err != nil && firstErr == nil {
				firstErr = err
			}
			if val != nil {
				results[idx] = val
			}
			mu.Unlock()
		}(i, item)
	}

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	// Compact nil entries
	compacted := make([]interface{}, 0, len(results))
	for _, r := range results {
		if r != nil {
			compacted = append(compacted, r)
		}
	}
	return compacted, nil
}

func runForeachUnordered(ctx context.Context, list []interface{}, itemName string, concurrency int, pipeline []pipeline.StepConfig, rctx pipeline.StepContext, executor pipeline.StepExecutor) ([]interface{}, error) {
	var results []interface{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	var firstErr error

	for i, item := range list {
		if firstErr != nil {
			break
		}
		wg.Add(1)
		go func(idx int, item interface{}) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				mu.Lock()
				if firstErr == nil {
					firstErr = ctx.Err()
				}
				mu.Unlock()
				return
			}

			val, err := runSubPipeline(ctx, idx, item, itemName, pipeline, rctx, executor)
			mu.Lock()
			if err != nil && firstErr == nil {
				firstErr = err
			}
			if val != nil {
				results = append(results, val)
			}
			mu.Unlock()
		}(i, item)
	}

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return results, nil
}

func runSubPipeline(ctx context.Context, idx int, item interface{}, itemName string, pipeline []pipeline.StepConfig, rctx pipeline.StepContext, executor pipeline.StepExecutor) (interface{}, error) {
	itemCtx := rctx.WithItem(item, itemName)
	for _, subStep := range pipeline {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		s := subStep
		val, err := executor.ExecuteStep(ctx, &s, itemCtx)
		if err != nil {
			return nil, fmt.Errorf("foreach[%d] step %q: %w", idx, s.ID, err)
		}
		itemCtx.SetOutput(s.ID, val)
		if s.Kind == "emit" {
			return val, nil
		}
	}
	return nil, fmt.Errorf("foreach[%d]: sub-pipeline completed without emit step", idx)
}
