package node

import (
	"context"
	"fmt"
	"sync"

	"confluence-mcp-v10.2.14/internal/mcpaggregator/pipeline"
)

const maxConcurrency = 8

// MapNode executes a map step — iterating over a list and running a sub-pipeline per item.
func MapNode(ctx context.Context, step *pipeline.StepConfig, rctx pipeline.StepContext, executor pipeline.StepExecutor) (interface{}, error) {
	cfg := step.Map

	source, err := rctx.ResolvePath(cfg.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve map source: %w", err)
	}

	list, ok := source.([]interface{})
	if !ok {
		return nil, fmt.Errorf("map source must resolve to an array, got %T", source)
	}

	results := make([]interface{}, len(list))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency)
	var firstErr error

	setFirstErr := func(err error) {
		mu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		mu.Unlock()
	}

	getFirstErr := func() error {
		mu.Lock()
		defer mu.Unlock()
		return firstErr
	}

	for i, item := range list {
		wg.Add(1)
		go func(i int, item interface{}) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				setFirstErr(ctx.Err())
				return
			}

			itemCtx := rctx.WithItem(item)
			for _, subStep := range cfg.Pipeline {
				if getFirstErr() != nil {
					return
				}
				s := subStep
				val, err := executor.ExecuteStep(ctx, &s, itemCtx)
				if err != nil {
					setFirstErr(fmt.Errorf("map[%d] step %q: %w", i, s.Name, err))
					return
				}
				if s.Output != "" {
					itemCtx.SetOutput(s.Output, val)
				}
				itemCtx.SetOutput(s.Name, val)
				if s.Type == "return" {
					mu.Lock()
					results[i] = val
					mu.Unlock()
					return
				}
			}
			setFirstErr(fmt.Errorf("map[%d]: sub-pipeline completed without return step", i))
		}(i, item)
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	return results, nil
}
