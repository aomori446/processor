package processor

import (
	"context"
	"sync"
)

type Result struct {
	Val any
	Err error
}

func NewResult[R any](val R, err error) Result {
	return Result{Val: val, Err: err}
}

type Results <-chan Result

func (rs Results) HandleResults(f func(result Result)) {
	for r := range rs {
		f(r)
	}
}

func CollectResults(ctx context.Context, rs ...Results) Results {
	if len(rs) == 0 {
		ch := make(chan Result)
		close(ch)
		return ch
	}
	if len(rs) == 1 {
		return rs[0]
	}

	ch := make(chan Result, len(rs))
	var wg sync.WaitGroup

	output := func(r Results) {
		defer wg.Done()
		for result := range OrDone(ctx, r) {
			select {
			case ch <- result:
			case <-ctx.Done():
				return
			}
		}
	}

	wg.Add(len(rs))
	for _, r := range rs {
		go output(r)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}
