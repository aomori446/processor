package processor

import (
	"context"
	"errors"
	"log"
	"math/rand/v2"
	"testing"
	"time"
)

func WorksGenerator2(amount int) []func(ctx context.Context) Result {
	rs := make([]func(ctx context.Context) Result, amount)
	for i := range amount {
		rs[i] = func(ctx context.Context) Result {
			time.Sleep(time.Duration(i) * time.Millisecond)
			if i%2 == 0 {
				return Result{Val: i, Err: errors.New("an error on purpose")}
			}
			return Result{Val: i, Err: nil}
		}
	}
	return rs
}

func TestProcessor_DoWithParallel2(t *testing.T) {
	type fields struct {
		ctx context.Context
	}
	type args struct {
		works        []func(ctx context.Context) Result
		numWorkers   int
		handleResult func(result Result)
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "0 work,1 worker",
			fields: fields{ctx: context.Background()},
			args: args{
				works:      []func(ctx context.Context) Result{},
				numWorkers: 1,
				handleResult: func(result Result) {
					log.Println(result.Val, result.Err)
				},
			},
		},
		{
			name:   "10 work,1 worker",
			fields: fields{ctx: context.Background()},
			args: args{
				works:      WorksGenerator2(10),
				numWorkers: 1,
				handleResult: func(result Result) {
					log.Println(result.Val, result.Err)
				},
			},
		},
		{
			name:   "10 work,0 worker",
			fields: fields{ctx: context.Background()},
			args: args{
				works:      WorksGenerator2(10),
				numWorkers: 0,
				handleResult: func(result Result) {
					log.Println(result.Val, result.Err)
				},
			},
		},
		{
			name:   "10 work,1 worker,don't handle result",
			fields: fields{ctx: context.Background()},
			args: args{
				works:        WorksGenerator2(100),
				numWorkers:   0,
				handleResult: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				ctx: tt.fields.ctx,
			}
			p.DoWithParallel2(tt.args.works, tt.args.numWorkers, tt.args.handleResult)
		})
	}
}

func RepeatWork() Result {
	time.Sleep(time.Millisecond)
	r := rand.IntN(10) % 2
	if r == 0 {
		return Result{Val: r, Err: nil}
	}
	return Result{Val: r, Err: errors.New("an error on purpose")}
}

func TestProcessor_DoWithRepeat(t *testing.T) {
	type fields struct {
		ctx context.Context
	}
	type args struct {
		work         func(ctx context.Context) Result
		rate         time.Duration
		times        int
		handleResult func(result Result)
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "0.5 ops, 5 times",
			fields: fields{ctx: context.Background()},
			args: args{
				work: func(ctx context.Context) Result {
					return RepeatWork()
				},
				rate:  time.Second * 2,
				times: 5,
				handleResult: func(result Result) {
					log.Println(result)
				},
			},
		},
		{
			name:   "1 ops, 0 time",
			fields: fields{ctx: context.Background()},
			args: args{
				work: func(ctx context.Context) Result {
					return RepeatWork()
				},
				rate:  time.Second,
				times: 0,
				handleResult: func(result Result) {
					log.Println(result)
				},
			},
		},
		{
			name:   "10 ops, 100 times",
			fields: fields{ctx: context.Background()},
			args: args{
				work: func(ctx context.Context) Result {
					return RepeatWork()
				},
				rate:  time.Millisecond * 100,
				times: 100,
				handleResult: func(result Result) {
					log.Println(result)
				},
			},
		},
		{
			name:   "10 ops, 100 times, don't handle result",
			fields: fields{ctx: context.Background()},
			args: args{
				work: func(ctx context.Context) Result {
					return RepeatWork()
				},
				rate:         time.Millisecond * 100,
				times:        100,
				handleResult: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				ctx: tt.fields.ctx,
			}
			p.DoWithRepeat(tt.args.work, tt.args.rate, tt.args.times, tt.args.handleResult)
		})
	}
}
