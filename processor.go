package processor

import (
	"context"
	"sync"
	"time"
)

type Processor struct {
	ctx context.Context
}

func NewProcessor(ctx context.Context) *Processor {
	processor := &Processor{
		ctx: ctx,
	}
	return processor
}

func (p *Processor) DoWithParallel(works <-chan Work, numWorkers int) Results {
	ch := make(chan Result, numWorkers)
	var wg sync.WaitGroup

	if numWorkers <= 0 {
		numWorkers = 1
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range OrDone(p.ctx, works) {
				ch <- work.Do(p.ctx)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

func (p *Processor) DoWithRate(works <-chan Work, rate time.Duration) Results {
	ch := make(chan Result)
	go func() {
		defer close(ch)
		for work := range SpeedLimit(p.ctx, works, rate) {
			ch <- work.Do(p.ctx)
		}
	}()
	return ch
}

func (p *Processor) DoWithRepeat(work Work, rate time.Duration, until time.Time) Results {
	if work.Do == nil || rate < time.Millisecond || until.Before(time.Now()) {
		ch := make(chan Result)
		close(ch)
		return ch
	}
	ch := make(chan Result)
	ticker := time.NewTicker(rate)
	ctx, cancel := context.WithDeadline(p.ctx, until)

	go func() {
		defer close(ch)
		defer ticker.Stop()
		defer cancel()

		select {
		case ch <- work.Do(p.ctx):
		case <-ctx.Done():
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				select {
				case ch <- work.Do(p.ctx):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch
}
