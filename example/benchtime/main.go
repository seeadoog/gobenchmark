package main

import (
	"context"
	"math"
	"time"

	"github.com/seeadoog/gobenchmark"
	"github.com/seeadoog/gobenchmark/app"
)

// main
func main() {
	myMetric := gobenchmark.NewHistogram("custom cost", gobenchmark.Uniform(0, 1000000, 10000), "us")
	a := app.New("benchmark")

	bentime := a.NewApp("time")
	//count := bentime.Flags().Int64("count", 10, "count")
	//var counter *gobenchmark.Counter
	//bentime.Cmd().PersistentPreRun = func(cmd *cobra.Command, args []string) {
	//	ctx, cc := gobenchmark.NewCounterContext(context.TODO(), *count)
	//	counter = cc
	//	bentime.SetContext(ctx)
	//}
	m1 := gobenchmark.NewHistogram("11", gobenchmark.DefaultBuckets, "ms")
	bentime.SetTask(func(t context.Context, b *gobenchmark.Benchmark) (err error) {
		st := time.Now()
		math.Sqrt(1.3)
		//myMetric.Add(float64(time.Since(st).Nanoseconds() / 1000))
		tm := time.NewTimer(300 * time.Millisecond)
		select {
		//case <-t.Done():
		case <-tm.C:
		}

		m1.Add(float64(time.Since(st)) / 1e6)
		return nil
	}, gobenchmark.DefaultBuckets, m1)

	benchCost := a.NewApp("cost")
	benchCost.GoMaxProc = 1
	benchCost.SetTask(func(t context.Context, b *gobenchmark.Benchmark) (err error) {

		return nil
	}, gobenchmark.Uniform(0, 1000000, 10000), myMetric)
	a.Start()
}
