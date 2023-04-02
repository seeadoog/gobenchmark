package main

import (
	"context"
	"github.com/seeadoog/gobenchmark"
	"github.com/seeadoog/gobenchmark/app"
	"math"
	"math/rand"
	"time"
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

	bentime.SetTask(func(t context.Context, b *gobenchmark.Benchmark) (err error) {
		//st := time.Now()
		math.Sqrt(1.3)
		//myMetric.Add(float64(time.Since(st).Nanoseconds() / 1000))
		tm := time.NewTimer(10 * time.Millisecond)
		select {
		case <-t.Done():
		case <-tm.C:
		}
		return nil
	}, gobenchmark.DefaultBuckets)

	benchCost := a.NewApp("cost")

	benchCost.SetTask(func(t context.Context, b *gobenchmark.Benchmark) (err error) {
		st := time.Now()
		time.Sleep(time.Duration(rand.Int()%1) * time.Millisecond)
		myMetric.Add(float64(time.Since(st).Nanoseconds() / 1000))
		return nil
	}, gobenchmark.Uniform(0, 1000000, 10000), myMetric)
	a.Start()
}
