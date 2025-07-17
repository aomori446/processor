package processor

import (
	"context"
	"time"
)

func OrDone[T any](ctx context.Context, works <-chan T) <-chan T {
	ch := make(chan T)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case work, ok := <-works:
				if !ok {
					return
				}
				select {
				case ch <- work:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return ch
}

func SpeedLimit(ctx context.Context, works <-chan Work, speedLimit time.Duration) <-chan Work {
	out := make(chan Work)

	if speedLimit <= time.Millisecond {
		panic("最快速度為1op/ms")
	}

	ticker := time.NewTicker(speedLimit)

	go func() {
		defer close(out)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case work, ok := <-works:
				if !ok {
					return
				}

				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					select {
					case out <- work:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return out
}
