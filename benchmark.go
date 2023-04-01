package gobenchmark

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

type Benchmark struct {
	concurrency int
	task        Task
	ctx         context.Context
	wg          sync.WaitGroup
	defaultCost *Histogram
}

func NewBenchmark(ctx context.Context, concurrency int, costBucket []float64, task Task) *Benchmark {
	return &Benchmark{
		concurrency: concurrency,
		task:        task,
		ctx:         ctx,
		wg:          sync.WaitGroup{},
		defaultCost: NewHistogram("cost", costBucket, "us"),
	}
}

func (b *Benchmark) start() {
	for i := 0; i < b.concurrency; i++ {
		b.wg.Add(1)
		go func() {
			defer b.wg.Done()
			for {
				select {
				case <-b.ctx.Done():
					return
				default:

				}
				startTime := time.Now()
				err := b.task(b.ctx, b)
				cost := time.Since(startTime)
				if err != nil {
					b.Println("error=>", err)
					time.Sleep(1 * time.Second)
					continue
				}

				b.defaultCost.Add(float64(cost.Nanoseconds() / 1e3))
			}
		}()
	}
}

func (b *Benchmark) Start() {
	b.start()
	b.wg.Wait()
}

func (b *Benchmark) String(costSecond float64) string {
	return b.defaultCost.Metrics(costSecond).String()
}

func (b *Benchmark) Metrics() *Histogram {
	return b.defaultCost
}

func (b *Benchmark) Println(args ...any) {
	fmt.Fprintln(os.Stderr, args...)
}

type Metric interface {
	Value() float64
	Name() string
}

type Task func(t context.Context, b *Benchmark) (err error)
