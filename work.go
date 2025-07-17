package processor

import "context"

type Work struct {
	Do func(ctx context.Context) Result
}
