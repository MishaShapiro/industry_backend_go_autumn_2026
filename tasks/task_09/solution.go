package main

import (
	"context"
	"errors"
	"sync"
)

type Result[R any] struct {
	Value R
	Err   error
}

var ErrInvalidWorkers = errors.New("workers must be positive")

func ParallelMap[T any, R any](ctx context.Context, workers int, in []T, fn func(context.Context, T) (R, error)) ([]Result[R], error) {
	if workers <= 0 {
		return nil, ErrInvalidWorkers
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	out := make([]Result[R], len(in))
	if len(in) == 0 {
		return out, nil
	}

	jobs := make(chan int)
	var wg sync.WaitGroup
	for range min(workers, len(in)) {
		wg.Go(func() {
			for i := range jobs {
				if ctx.Err() != nil {
					continue
				}
				v, err := fn(ctx, in[i])
				if err != nil {
					var zero R
					v = zero
				}
				out[i] = Result[R]{Value: v, Err: err}
			}
		})
	}

	func() {
		defer close(jobs)
		for i := range in {
			select {
			case jobs <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	wg.Wait()

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
