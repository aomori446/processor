package processor

import (
	"context"
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
