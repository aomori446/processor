package processor

import (
	"context"
	"sync"
	"time"
)

type Processor struct {
	ctx context.Context
}

type Result struct {
	Val any
	Err error
}

func NewProcessor(ctx context.Context) *Processor {
	processor := &Processor{
		ctx: ctx,
	}
	return processor
}

func (p *Processor) DoWithParallel(works <-chan func(ctx context.Context) Result, numWorkers int, handleResult func(result Result)) {
	if numWorkers <= 0 {
		numWorkers = 1
	}

	if handleResult == nil {
		handleResult = func(result Result) {}
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range OrDone(p.ctx, works) {
				handleResult(work(p.ctx))
			}
		}()
	}

	wg.Wait()
}

func (p *Processor) DoWithParallel2(works []func(ctx context.Context) Result, numWorkers int, handleResult func(result Result)) {
	if numWorkers <= 0 {
		numWorkers = 1
	}

	if handleResult == nil {
		handleResult = func(result Result) {}
	}

	ws := make(chan func(ctx context.Context) Result, len(works))
	for _, work := range works {
		ws <- work
	}
	close(ws)

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for w := range ws {
				handleResult(w(p.ctx))
			}
		}()
	}

	wg.Wait()
}

func (p *Processor) DoWithRepeat(work func(ctx context.Context) Result, rate time.Duration, times int, handleResult func(result Result)) {
	if work == nil || rate < time.Millisecond {
		return
	}

	if handleResult == nil {
		handleResult = func(result Result) {}
	}

	ticker := time.NewTicker(rate)
	defer ticker.Stop()

	ctx, cancel := context.WithDeadline(p.ctx, time.Now().Add(rate*time.Duration(times)))
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go handleResult(work(ctx))
		}
	}
}

func (p *Processor) DoWithRepeat2(work func(ctx context.Context) Result, rate time.Duration, until time.Time, handleResult func(result Result)) {
	if work == nil || rate < time.Millisecond || until.Before(time.Now().Add(time.Millisecond)) {
		return
	}

	if handleResult == nil {
		handleResult = func(result Result) {}
	}

	ticker := time.NewTicker(rate)
	defer ticker.Stop()

	ctx, cancel := context.WithDeadline(p.ctx, until)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go handleResult(work(ctx))
		}
	}
}
